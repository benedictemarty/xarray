package xarray

import (
	"fmt"
	"math"
	"runtime"
	"sync"
)

// Évaluation paresseuse par blocs (« chunks »), dans l'esprit de dask/xarray.
//
// Un LazyArray décrit un calcul différé sur un tableau float64 découpé en chunks
// le long de l'axe 0. Les opérations (Map/AddScalar…) empilent des transformations
// sans rien calculer ; l'exécution n'a lieu qu'au Compute() (matérialisation
// parallèle) ou lors d'une réduction (Sum/Mean… en streaming, un chunk à la fois).
//
// Intérêt : traiter des tableaux plus grands que la RAM lorsqu'ils proviennent
// d'une source hors-mémoire (voir ChunkFile), sans jamais tout charger.

// ChunkSource fournit les données d'un tableau, bloc par bloc, le long de l'axe 0.
type ChunkSource interface {
	Dims() []string
	Shape() []int
	Coords() map[string][]float64
	NumChunks() int
	ChunkRows(i int) (start, end int) // plage de lignes (axe 0) du chunk i
	ChunkData(i int) ([]float64, error)
}

// LazyArray représente un calcul différé sur une ChunkSource.
type LazyArray struct {
	src  ChunkSource
	ops  []func(float64) float64
	name string
}

// --- Source mémoire ---------------------------------------------------------

type memSource struct {
	da        *DataArray[float64]
	chunkSize int
}

func (m *memSource) Dims() []string { return m.da.variable.Dims() }
func (m *memSource) Shape() []int   { return m.da.variable.Shape() }
func (m *memSource) Coords() map[string][]float64 {
	out := map[string][]float64{}
	for k, cv := range m.da.coords {
		out[k] = cv.Data()
	}
	return out
}
func (m *memSource) rowLen() int { s := m.Shape(); return product(s[1:]) }
func (m *memSource) NumChunks() int {
	n := m.Shape()[0]
	return (n + m.chunkSize - 1) / m.chunkSize
}
func (m *memSource) ChunkRows(i int) (int, int) {
	start := i * m.chunkSize
	end := start + m.chunkSize
	if n := m.Shape()[0]; end > n {
		end = n
	}
	return start, end
}
func (m *memSource) ChunkData(i int) ([]float64, error) {
	s, e := m.ChunkRows(i)
	rl := m.rowLen()
	return append([]float64(nil), m.da.variable.data[s*rl:e*rl]...), nil
}

// Chunk crée un LazyArray adossé à un DataArray en mémoire, découpé en blocs de
// chunkSize lignes (le long de l'axe 0).
func Chunk(da *DataArray[float64], chunkSize int) (*LazyArray, error) {
	if chunkSize < 1 {
		return nil, fmt.Errorf("xarray: taille de chunk invalide %d", chunkSize)
	}
	if da.Ndim() == 0 {
		return nil, fmt.Errorf("xarray: impossible de chunker un scalaire")
	}
	return &LazyArray{src: &memSource{da: da, chunkSize: chunkSize}, name: da.name}, nil
}

// --- Construction du graphe (paresseux) -------------------------------------

func (l *LazyArray) with(op func(float64) float64) *LazyArray {
	ops := make([]func(float64) float64, len(l.ops)+1)
	copy(ops, l.ops)
	ops[len(l.ops)] = op
	return &LazyArray{src: l.src, ops: ops, name: l.name}
}

// Map empile une transformation élément par élément (différée).
func (l *LazyArray) Map(fn func(float64) float64) *LazyArray { return l.with(fn) }

// AddScalar empile l'ajout d'un scalaire.
func (l *LazyArray) AddScalar(s float64) *LazyArray {
	return l.with(func(x float64) float64 { return x + s })
}

// MulScalar empile la multiplication par un scalaire.
func (l *LazyArray) MulScalar(s float64) *LazyArray {
	return l.with(func(x float64) float64 { return x * s })
}

// NumChunks renvoie le nombre de blocs.
func (l *LazyArray) NumChunks() int { return l.src.NumChunks() }

func (l *LazyArray) apply(chunk []float64) {
	for _, op := range l.ops {
		for i, x := range chunk {
			chunk[i] = op(x)
		}
	}
}

// --- Exécution --------------------------------------------------------------

// forEachChunk exécute fn(i, data) pour chaque chunk, en parallèle. Les opérations
// différées sont appliquées avant fn. Renvoie la première erreur rencontrée.
func (l *LazyArray) forEachChunk(fn func(i int, data []float64)) error {
	nc := l.src.NumChunks()
	nw := runtime.GOMAXPROCS(0)
	if nw > nc {
		nw = nc
	}
	if nw < 1 {
		nw = 1
	}
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		firstEr error
		next    = make(chan int)
	)
	go func() {
		for i := 0; i < nc; i++ {
			next <- i
		}
		close(next)
	}()
	for w := 0; w < nw; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				data, err := l.src.ChunkData(i)
				if err != nil {
					mu.Lock()
					if firstEr == nil {
						firstEr = err
					}
					mu.Unlock()
					continue
				}
				l.apply(data)
				fn(i, data)
			}
		}()
	}
	wg.Wait()
	return firstEr
}

// Compute matérialise le résultat en DataArray[float64] (exécution parallèle par
// chunk).
func (l *LazyArray) Compute() (*DataArray[float64], error) {
	shape := l.src.Shape()
	rl := product(shape[1:])
	out := make([]float64, product(shape))
	err := l.forEachChunk(func(i int, data []float64) {
		start, _ := l.src.ChunkRows(i)
		copy(out[start*rl:], data) // plages disjointes -> pas de course
	})
	if err != nil {
		return nil, err
	}
	return NewDataArray(l.src.Dims(), shape, out, l.src.Coords(), l.name)
}

// --- Réductions en streaming (un chunk à la fois) ---------------------------

func (l *LazyArray) reduce(init float64, combine func(acc, x float64) float64) (float64, int, error) {
	acc := init
	count := 0
	var mu sync.Mutex
	err := l.forEachChunk(func(i int, data []float64) {
		local := init
		for _, x := range data {
			local = combine(local, x)
		}
		mu.Lock()
		acc = combine(acc, local)
		count += len(data)
		mu.Unlock()
	})
	return acc, count, err
}

// Sum agrège toutes les valeurs sans charger l'ensemble en mémoire.
func (l *LazyArray) Sum() (float64, error) {
	s, _, err := l.reduce(0, func(a, x float64) float64 { return a + x })
	return s, err
}

// Mean renvoie la moyenne (streaming).
func (l *LazyArray) Mean() (float64, error) {
	s, n, err := l.reduce(0, func(a, x float64) float64 { return a + x })
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return math.NaN(), nil
	}
	return s / float64(n), nil
}

// Min renvoie le minimum (streaming).
func (l *LazyArray) Min() (float64, error) {
	m, _, err := l.reduce(math.Inf(1), math.Min)
	return m, err
}

// Max renvoie le maximum (streaming).
func (l *LazyArray) Max() (float64, error) {
	m, _, err := l.reduce(math.Inf(-1), math.Max)
	return m, err
}
