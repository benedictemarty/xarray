package xarray

import "fmt"

// Transpose renvoie une nouvelle Variable dont les axes sont réordonnés selon
// newDims. newDims doit être une permutation des dimensions courantes.
func (v *Variable) Transpose(newDims ...string) (*Variable, error) {
	if len(newDims) != len(v.dims) {
		return nil, fmt.Errorf("xarray: transpose attend %d dimension(s), %d fournie(s)", len(v.dims), len(newDims))
	}
	// perm[i] = ancien axe placé en position i.
	perm := make([]int, len(newDims))
	for i, d := range newDims {
		ax := v.dimIndex(d)
		if ax == -1 {
			return nil, fmt.Errorf("xarray: dimension %q inconnue", d)
		}
		perm[i] = ax
	}
	// Vérifie que c'est bien une permutation (pas de doublon).
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

	out := &Variable{
		dims:  append([]string(nil), newDims...),
		shape: newShape,
		data:  make([]float64, v.Size()),
		attrs: v.Attrs(),
	}
	// Pour chaque position de sortie, retrouver la position source.
	counter := make([]int, len(newShape))
	for flatOut := range out.data {
		// counter est le multi-indice de sortie.
		flatIn := 0
		for i, c := range counter {
			flatIn += c * oldStrides[perm[i]]
		}
		out.data[flatOut] = v.data[flatIn]
		// Incrément du compteur (ordre C).
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
// sans supprimer la dimension (contrairement à Isel). L'ordre des indices est
// conservé.
func (v *Variable) take(dim string, indices []int) (*Variable, error) {
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

	out := &Variable{
		dims:  v.Dims(),
		shape: newShape,
		data:  make([]float64, product(newShape)),
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

// reduceAxis réduit la dimension dim en appliquant reducer sur chaque vecteur
// le long de cet axe. La dimension disparaît du résultat.
func (v *Variable) reduceAxis(dim string, reducer func([]float64) float64) (*Variable, error) {
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

	out := &Variable{
		dims:  newDims,
		shape: newShape,
		data:  make([]float64, product(newShape)),
		attrs: v.Attrs(),
	}
	oldStrides := v.strides()
	n := v.shape[axis]

	// counter parcourt l'espace de sortie (dimensions sans l'axe réduit).
	counter := make([]int, len(newShape))
	buf := make([]float64, n)
	for flatOut := range out.data {
		// Base : position dans le tableau source avec l'axe réduit à 0.
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
func binaryOp(a, b *Variable, fn func(x, y float64) float64) (*Variable, error) {
	// Construction des dimensions et tailles du résultat.
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

	out := &Variable{
		dims:  resDims,
		shape: resShape,
		data:  make([]float64, product(resShape)),
		attrs: map[string]string{},
	}

	aStrides := strideByDim(a)
	bStrides := strideByDim(b)

	counter := make([]int, len(resDims))
	for flat := range out.data {
		flatA, flatB := 0, 0
		for i, d := range resDims {
			if st, ok := aStrides[d]; ok {
				flatA += counter[i] * st
			}
			if st, ok := bStrides[d]; ok {
				flatB += counter[i] * st
			}
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

// mapScalar applique fn(x) à chaque élément et renvoie une nouvelle Variable.
func (v *Variable) mapScalar(fn func(float64) float64) *Variable {
	out := &Variable{
		dims:  v.Dims(),
		shape: v.Shape(),
		data:  make([]float64, len(v.data)),
		attrs: v.Attrs(),
	}
	for i, x := range v.data {
		out.data[i] = fn(x)
	}
	return out
}

// strideByDim associe à chaque nom de dimension son stride (ordre C).
func strideByDim(v *Variable) map[string]int {
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
