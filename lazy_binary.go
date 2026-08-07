package xarray

import "fmt"

// Combinaison paresseuse de deux LazyArray, élément par élément, chunk par chunk.
// Les deux sources doivent avoir la même forme et le même découpage (même nombre
// de chunks). Comme tout le reste, le calcul est différé et par bloc : deux
// tableaux hors-mémoire (fichier, Zarr) peuvent ainsi être combinés sans être
// chargés entièrement.

type binarySource struct {
	a, b *LazyArray
	op   func(x, y float64) float64
}

func (s *binarySource) Dims() []string               { return s.a.src.Dims() }
func (s *binarySource) Shape() []int                 { return s.a.src.Shape() }
func (s *binarySource) Coords() map[string][]float64 { return s.a.src.Coords() }
func (s *binarySource) NumChunks() int               { return s.a.src.NumChunks() }
func (s *binarySource) ChunkRows(i int) (int, int)   { return s.a.src.ChunkRows(i) }

func (s *binarySource) ChunkData(i int) ([]float64, error) {
	da, err := s.a.src.ChunkData(i)
	if err != nil {
		return nil, err
	}
	s.a.apply(da)
	db, err := s.b.src.ChunkData(i)
	if err != nil {
		return nil, err
	}
	s.b.apply(db)
	if len(da) != len(db) {
		return nil, fmt.Errorf("xarray: chunks de tailles différentes (%d vs %d)", len(da), len(db))
	}
	for j := range da {
		da[j] = s.op(da[j], db[j])
	}
	return da, nil
}

func sameIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (l *LazyArray) binary(other *LazyArray, op func(x, y float64) float64) (*LazyArray, error) {
	if !sameIntSlice(l.src.Shape(), other.src.Shape()) {
		return nil, fmt.Errorf("xarray: formes incompatibles pour la combinaison (%v vs %v)", l.src.Shape(), other.src.Shape())
	}
	if l.src.NumChunks() != other.src.NumChunks() {
		return nil, fmt.Errorf("xarray: découpages incompatibles (%d vs %d chunks)", l.src.NumChunks(), other.src.NumChunks())
	}
	return &LazyArray{src: &binarySource{a: l, b: other, op: op}, name: l.name}, nil
}

// Add combine deux LazyArray par addition (différée, chunk par chunk).
func (l *LazyArray) Add(other *LazyArray) (*LazyArray, error) {
	return l.binary(other, func(x, y float64) float64 { return x + y })
}

// Sub combine par soustraction.
func (l *LazyArray) Sub(other *LazyArray) (*LazyArray, error) {
	return l.binary(other, func(x, y float64) float64 { return x - y })
}

// Mul combine par multiplication.
func (l *LazyArray) Mul(other *LazyArray) (*LazyArray, error) {
	return l.binary(other, func(x, y float64) float64 { return x * y })
}

// Div combine par division.
func (l *LazyArray) Div(other *LazyArray) (*LazyArray, error) {
	return l.binary(other, func(x, y float64) float64 { return x / y })
}
