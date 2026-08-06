package xarray

import (
	"fmt"
	"runtime"
	"sync"
)

// Transpose renvoie une nouvelle Variable dont les axes sont réordonnés selon
// newDims. newDims doit être une permutation des dimensions courantes.
func (v *Variable[T]) Transpose(newDims ...string) (*Variable[T], error) {
	if len(newDims) != len(v.dims) {
		return nil, fmt.Errorf("xarray: transpose attend %d dimension(s), %d fournie(s)", len(v.dims), len(newDims))
	}
	perm := make([]int, len(newDims))
	for i, d := range newDims {
		ax := v.dimIndex(d)
		if ax == -1 {
			return nil, fmt.Errorf("xarray: dimension %q inconnue", d)
		}
		perm[i] = ax
	}
	seen := make(map[int]struct{}, len(perm))
	for _, p := range perm {
		if _, ok := seen[p]; ok {
			return nil, fmt.Errorf("xarray: dimension %q répétée dans transpose", v.dims[p])
		}
		seen[p] = struct{}{}
	}

	newShape := make([]int, len(perm))
	for i, p := range perm {
		newShape[i] = v.shape[p]
	}
	oldStrides := v.strides()

	out := &Variable[T]{
		dims:  append([]string(nil), newDims...),
		shape: newShape,
		data:  make([]T, v.Size()),
		attrs: v.Attrs(),
	}

	counter := make([]int, len(newShape))
	for flatOut := range out.data {
		flatIn := 0
		for i, c := range counter {
			flatIn += c * oldStrides[perm[i]]
		}
		out.data[flatOut] = v.data[flatIn]
		for k := len(counter) - 1; k >= 0; k-- {
			counter[k]++
			if counter[k] < newShape[k] {
				break
			}
			counter[k] = 0
		}
	}
	return out, nil
}

// take sélectionne plusieurs positions le long de l'axe correspondant à dim,
// sans supprimer la dimension (contrairement à Isel). L'ordre est conservé.
func (v *Variable[T]) take(dim string, indices []int) (*Variable[T], error) {
	axis := v.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	for _, idx := range indices {
		if idx < 0 || idx >= v.shape[axis] {
			return nil, fmt.Errorf("xarray: indice %d hors bornes sur %q", idx, dim)
		}
	}
	newShape := append([]int(nil), v.shape...)
	newShape[axis] = len(indices)

	out := &Variable[T]{
		dims:  v.Dims(),
		shape: newShape,
		data:  make([]T, product(newShape)),
		attrs: v.Attrs(),
	}
	oldStrides := v.strides()

	counter := make([]int, len(newShape))
	for flatOut := range out.data {
		flatIn := 0
		for i, c := range counter {
			src := c
			if i == axis {
				src = indices[c]
			}
			flatIn += src * oldStrides[i]
		}
		out.data[flatOut] = v.data[flatIn]
		for k := len(newShape) - 1; k >= 0; k-- {
			counter[k]++
			if counter[k] < newShape[k] {
				break
			}
			counter[k] = 0
		}
	}
	return out, nil
}

// mapScalar applique fn(x) à chaque élément et renvoie une nouvelle Variable.
func (v *Variable[T]) mapScalar(fn func(T) T) *Variable[T] {
	out := &Variable[T]{
		dims:  v.Dims(),
		shape: v.Shape(),
		data:  make([]T, len(v.data)),
		attrs: v.Attrs(),
	}
	for i, x := range v.data {
		out.data[i] = fn(x)
	}
	return out
}

// reduceAxisVar réduit la dimension dim en appliquant reducer sur chaque vecteur
// le long de cet axe. Le type de sortie R peut différer du type d'entrée T
// (ex. moyenne d'entiers -> float64). C'est une fonction libre car une méthode
// Go ne peut pas introduire de paramètre de type supplémentaire.
func reduceAxisVar[T, R Number](v *Variable[T], dim string, reducer func([]T) R) (*Variable[R], error) {
	axis := v.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	newDims := make([]string, 0, len(v.dims)-1)
	newShape := make([]int, 0, len(v.shape)-1)
	for i := range v.dims {
		if i == axis {
			continue
		}
		newDims = append(newDims, v.dims[i])
		newShape = append(newShape, v.shape[i])
	}

	out := &Variable[R]{
		dims:  newDims,
		shape: newShape,
		data:  make([]R, product(newShape)),
		attrs: v.Attrs(),
	}
	oldStrides := v.strides()
	n := v.shape[axis]

	counter := make([]int, len(newShape))
	buf := make([]T, n)
	for flatOut := range out.data {
		base := 0
		j := 0
		for i := range v.dims {
			if i == axis {
				continue
			}
			base += counter[j] * oldStrides[i]
			j++
		}
		for k := 0; k < n; k++ {
			buf[k] = v.data[base+k*oldStrides[axis]]
		}
		out.data[flatOut] = reducer(buf)

		for k := len(newShape) - 1; k >= 0; k-- {
			counter[k]++
			if counter[k] < newShape[k] {
				break
			}
			counter[k] = 0
		}
	}
	return out, nil
}

// binaryOp combine deux Variables élément par élément avec broadcasting par nom
// de dimension. Les dimensions communes doivent avoir la même taille ; une
// dimension présente dans un seul opérande est diffusée sur l'autre.
//
// L'ordre des dimensions du résultat est : celles de a, puis celles de b
// absentes de a.
func binaryOp[T Number](a, b *Variable[T], fn func(x, y T) T) (*Variable[T], error) {
	// Chemin rapide : dimensions identiques (mêmes noms, même ordre, mêmes
	// tailles) — aucun broadcasting, boucle directe sans calcul d'indices.
	if sameDimsShape(a, b) {
		out := &Variable[T]{
			dims:  a.Dims(),
			shape: a.Shape(),
			data:  make([]T, len(a.data)),
			attrs: map[string]string{},
		}
		for i := range a.data {
			out.data[i] = fn(a.data[i], b.data[i])
		}
		return out, nil
	}

	resDims := append([]string(nil), a.dims...)
	sizeByDim := make(map[string]int, len(a.dims)+len(b.dims))
	for i, d := range a.dims {
		sizeByDim[d] = a.shape[i]
	}
	for i, d := range b.dims {
		if s, ok := sizeByDim[d]; ok {
			if s != b.shape[i] {
				return nil, fmt.Errorf("xarray: tailles incompatibles pour la dimension %q (%d vs %d)", d, s, b.shape[i])
			}
			continue
		}
		sizeByDim[d] = b.shape[i]
		resDims = append(resDims, d)
	}
	resShape := make([]int, len(resDims))
	for i, d := range resDims {
		resShape[i] = sizeByDim[d]
	}

	out := &Variable[T]{
		dims:  resDims,
		shape: resShape,
		data:  make([]T, product(resShape)),
		attrs: map[string]string{},
	}

	// Pré-calcul des strides par position du résultat (0 si la dimension est
	// absente de l'opérande, ce qui la neutralise). Évite tout accès à une map
	// dans la boucle interne parcourue à chaque élément.
	aStrides := strideByDim(a)
	bStrides := strideByDim(b)
	aSt := make([]int, len(resDims))
	bSt := make([]int, len(resDims))
	for i, d := range resDims {
		aSt[i] = aStrides[d]
		bSt[i] = bStrides[d]
	}

	resStrides := out.strides()

	// Au-delà d'un seuil, on répartit le remplissage sur plusieurs cœurs : chaque
	// travailleur écrit une plage disjointe de out.data (aucune course de données).
	const seuilParallele = 1 << 15
	size := len(out.data)
	if size >= seuilParallele {
		nw := runtime.GOMAXPROCS(0)
		if nw > size {
			nw = size
		}
		chunk := (size + nw - 1) / nw
		var wg sync.WaitGroup
		for w := 0; w < nw; w++ {
			lo := w * chunk
			if lo >= size {
				break
			}
			hi := lo + chunk
			if hi > size {
				hi = size
			}
			wg.Add(1)
			go func(lo, hi int) {
				defer wg.Done()
				fillBinaryRange(out.data, a.data, b.data, fn, aSt, bSt, resShape, resStrides, lo, hi)
			}(lo, hi)
		}
		wg.Wait()
		return out, nil
	}

	fillBinaryRange(out.data, a.data, b.data, fn, aSt, bSt, resShape, resStrides, 0, size)
	return out, nil
}

// fillBinaryRange remplit out[lo:hi] en itérant de façon incrémentale sur les
// indices multidimensionnels. flatA/flatB sont maintenus par pas (O(1) amorti)
// plutôt que recalculés à chaque élément ; leur état initial est reconstruit à
// partir de lo via les strides du résultat.
func fillBinaryRange[T Number](out, adata, bdata []T, fn func(x, y T) T, aSt, bSt, resShape, resStrides []int, lo, hi int) {
	n := len(resShape)
	counter := make([]int, n)
	flatA, flatB := 0, 0
	rem := lo
	for i := 0; i < n; i++ {
		counter[i] = rem / resStrides[i]
		rem %= resStrides[i]
		flatA += counter[i] * aSt[i]
		flatB += counter[i] * bSt[i]
	}

	for flat := lo; flat < hi; flat++ {
		out[flat] = fn(adata[flatA], bdata[flatB])
		for k := n - 1; k >= 0; k-- {
			counter[k]++
			flatA += aSt[k]
			flatB += bSt[k]
			if counter[k] < resShape[k] {
				break
			}
			counter[k] = 0
			flatA -= aSt[k] * resShape[k]
			flatB -= bSt[k] * resShape[k]
		}
	}
}

// strideByDim associe à chaque nom de dimension son stride (ordre C).
func strideByDim[T Number](v *Variable[T]) map[string]int {
	st := v.strides()
	m := make(map[string]int, len(v.dims))
	for i, d := range v.dims {
		m[d] = st[i]
	}
	return m
}

// sameDimsShape indique si deux variables ont exactement les mêmes dimensions
// (noms, ordre) et la même forme.
func sameDimsShape[T Number](a, b *Variable[T]) bool {
	if len(a.dims) != len(b.dims) {
		return false
	}
	for i := range a.dims {
		if a.dims[i] != b.dims[i] || a.shape[i] != b.shape[i] {
			return false
		}
	}
	return true
}

func product(shape []int) int {
	p := 1
	for _, s := range shape {
		p *= s
	}
	return p
}
