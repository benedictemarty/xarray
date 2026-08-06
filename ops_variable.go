package xarray

import "fmt"

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

	counter := make([]int, len(resDims))
	for flat := range out.data {
		flatA, flatB := 0, 0
		for i := range resDims {
			flatA += counter[i] * aSt[i]
			flatB += counter[i] * bSt[i]
		}
		out.data[flat] = fn(a.data[flatA], b.data[flatB])

		for k := len(resDims) - 1; k >= 0; k-- {
			counter[k]++
			if counter[k] < resShape[k] {
				break
			}
			counter[k] = 0
		}
	}
	return out, nil
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

func product(shape []int) int {
	p := 1
	for _, s := range shape {
		p *= s
	}
	return p
}
