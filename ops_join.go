package xarray

import "fmt"

// JoinType désigne la stratégie d'alignement des coordonnées avant une opération
// binaire entre deux DataArrays.
type JoinType int

const (
	// JoinInner ne conserve que les étiquettes communes (jointure interne).
	JoinInner JoinType = iota
	// JoinOuter conserve l'union des étiquettes ; les valeurs manquantes sont
	// remplies par la valeur de remplissage fournie.
	JoinOuter
	// JoinLeft conserve les étiquettes de l'opérande gauche.
	JoinLeft
	// JoinRight conserve les étiquettes de l'opérande droit.
	JoinRight
)

func (j JoinType) String() string {
	switch j {
	case JoinInner:
		return "inner"
	case JoinOuter:
		return "outer"
	case JoinLeft:
		return "left"
	case JoinRight:
		return "right"
	default:
		return "inconnu"
	}
}

// joinLabels calcule l'ensemble ordonné des étiquettes cibles selon la stratégie.
func joinLabels[T Number](a, b []T, join JoinType) []T {
	inA := make(map[T]struct{}, len(a))
	for _, l := range a {
		inA[l] = struct{}{}
	}
	inB := make(map[T]struct{}, len(b))
	for _, l := range b {
		inB[l] = struct{}{}
	}
	switch join {
	case JoinLeft:
		return append([]T(nil), a...)
	case JoinRight:
		return append([]T(nil), b...)
	case JoinOuter:
		out := append([]T(nil), a...)
		for _, l := range b {
			if _, ok := inA[l]; !ok {
				out = append(out, l)
			}
		}
		return out
	default: // JoinInner
		var out []T
		for _, l := range a {
			if _, ok := inB[l]; ok {
				out = append(out, l)
			}
		}
		return out
	}
}

// takeFill sélectionne des positions le long de dim ; un indice -1 produit une
// tranche entièrement remplie par fill (utile pour les jointures externes).
func (v *Variable[T]) takeFill(dim string, indices []int, fill T) (*Variable[T], error) {
	axis := v.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	for _, idx := range indices {
		if idx < -1 || idx >= v.shape[axis] {
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
		missing := false
		flatIn := 0
		for i, c := range counter {
			src := c
			if i == axis {
				if indices[c] == -1 {
					missing = true
					break
				}
				src = indices[c]
			}
			flatIn += src * oldStrides[i]
		}
		if missing {
			out.data[flatOut] = fill
		} else {
			out.data[flatOut] = v.data[flatIn]
		}
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

// reindex réaligne le DataArray sur les étiquettes cibles de la dimension dim ;
// les étiquettes absentes sont remplies par fill. La coordonnée de dim devient
// la liste cible.
func (da *DataArray[T]) reindex(dim string, target []T, fill T) (*DataArray[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: reindex impossible : aucune coordonnée %q", dim)
	}
	pos := make(map[T]int, len(cv.data))
	for i, l := range cv.data {
		pos[l] = i
	}
	indices := make([]int, len(target))
	for i, l := range target {
		if p, ok := pos[l]; ok {
			indices[i] = p
		} else {
			indices[i] = -1
		}
	}
	nv, err := da.variable.takeFill(dim, indices, fill)
	if err != nil {
		return nil, err
	}
	coords := make(map[string]*Variable[T], len(da.coords))
	for k, c := range da.coords {
		if k == dim {
			nc, _ := NewVariable([]string{dim}, []int{len(target)}, append([]T(nil), target...))
			coords[k] = nc
			continue
		}
		nc, _ := NewVariable(c.Dims(), c.Shape(), c.Data())
		coords[k] = nc
	}
	return &DataArray[T]{variable: nv, coords: coords, name: da.name}, nil
}

// alignJoin aligne deux DataArrays selon la stratégie de jointure sur chaque
// dimension commune disposant de coordonnées des deux côtés.
func alignJoin[T Number](a, b *DataArray[T], join JoinType, fill T) (*DataArray[T], *DataArray[T], error) {
	a2, b2 := a, b
	for _, dim := range a.variable.dims {
		ca, okA := a2.coords[dim]
		cb, okB := b2.coords[dim]
		if !okA || !okB {
			continue
		}
		target := joinLabels(ca.data, cb.data, join)
		if len(target) == 0 {
			return nil, nil, fmt.Errorf("xarray: aucune étiquette après jointure %s sur %q", join, dim)
		}
		na, err := a2.reindex(dim, target, fill)
		if err != nil {
			return nil, nil, err
		}
		nb, err := b2.reindex(dim, target, fill)
		if err != nil {
			return nil, nil, err
		}
		a2, b2 = na, nb
	}
	return a2, b2, nil
}

// binaryJoin applique une opération binaire avec une stratégie de jointure et
// une valeur de remplissage explicites.
func (da *DataArray[T]) binaryJoin(other *DataArray[T], fn func(x, y T) T, join JoinType, fill T) (*DataArray[T], error) {
	a, b, err := alignJoin(da, other, join, fill)
	if err != nil {
		return nil, err
	}
	nv, err := binaryOp(a.variable, b.variable, fn)
	if err != nil {
		return nil, err
	}
	coords := make(map[string]*Variable[T], len(nv.dims))
	for _, dim := range nv.dims {
		if cv, ok := a.coords[dim]; ok {
			nc, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
			coords[dim] = nc
		} else if cv, ok := b.coords[dim]; ok {
			nc, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
			coords[dim] = nc
		}
	}
	return &DataArray[T]{variable: nv, coords: coords, name: da.name}, nil
}

// AddJoin renvoie da + other avec la stratégie de jointure et le remplissage donnés.
func (da *DataArray[T]) AddJoin(other *DataArray[T], join JoinType, fill T) (*DataArray[T], error) {
	return da.binaryJoin(other, func(x, y T) T { return x + y }, join, fill)
}

// SubJoin renvoie da - other avec jointure et remplissage.
func (da *DataArray[T]) SubJoin(other *DataArray[T], join JoinType, fill T) (*DataArray[T], error) {
	return da.binaryJoin(other, func(x, y T) T { return x - y }, join, fill)
}

// MulJoin renvoie da * other avec jointure et remplissage.
func (da *DataArray[T]) MulJoin(other *DataArray[T], join JoinType, fill T) (*DataArray[T], error) {
	return da.binaryJoin(other, func(x, y T) T { return x * y }, join, fill)
}

// DivJoin renvoie da / other avec jointure et remplissage.
func (da *DataArray[T]) DivJoin(other *DataArray[T], join JoinType, fill T) (*DataArray[T], error) {
	return da.binaryJoin(other, func(x, y T) T { return x / y }, join, fill)
}
