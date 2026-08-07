package xarray

import "fmt"

// Dot contracte deux DataArrays sur une dimension commune dim (produit
// tensoriel), à la manière de `xr.dot(a, b, dims=dim)`. Le résultat porte les
// dimensions restantes de a puis celles de b ; on somme sur dim :
//
//	c[i, j] = Σ_k a[i, k] · b[k, j]
//
// Pour ce MVP, dim doit être la seule dimension commune à a et b.
func Dot[T Number](a, b *DataArray[T], dim string) (*DataArray[T], error) {
	axisA := a.variable.dimIndex(dim)
	axisB := b.variable.dimIndex(dim)
	if axisA == -1 || axisB == -1 {
		return nil, fmt.Errorf("xarray: dimension de contraction %q absente d'un des tableaux", dim)
	}
	if a.variable.shape[axisA] != b.variable.shape[axisB] {
		return nil, fmt.Errorf("xarray: tailles de %q incompatibles (%d vs %d)", dim, a.variable.shape[axisA], b.variable.shape[axisB])
	}
	// Vérifie que dim est la seule dimension commune.
	for _, d := range a.variable.dims {
		if d == dim {
			continue
		}
		if b.variable.dimIndex(d) != -1 {
			return nil, fmt.Errorf("xarray: Dot ne gère qu'une seule dimension commune (%q aussi partagée)", d)
		}
	}

	K := a.variable.shape[axisA]
	kStrideA := a.variable.strides()[axisA]
	kStrideB := b.variable.strides()[axisB]

	aRestDims, aRestShape, aRestStrides := restOf(a.variable, axisA)
	bRestDims, bRestShape, bRestStrides := restOf(b.variable, axisB)

	aBases := basesOf(aRestShape, aRestStrides)
	bBases := basesOf(bRestShape, bRestStrides)

	resDims := append(append([]string(nil), aRestDims...), bRestDims...)
	resShape := append(append([]int(nil), aRestShape...), bRestShape...)
	nB := len(bBases)
	resData := make([]T, len(aBases)*nB)

	ad, bd := a.variable.data, b.variable.data
	for ia, aBase := range aBases {
		for ib, bBase := range bBases {
			var s T
			for k := 0; k < K; k++ {
				s += ad[aBase+k*kStrideA] * bd[bBase+k*kStrideB]
			}
			resData[ia*nB+ib] = s
		}
	}

	// Coordonnées : restantes de a puis de b (la dimension contractée disparaît).
	coords := map[string][]T{}
	for _, d := range aRestDims {
		if cv, ok := a.coords[d]; ok {
			coords[d] = cv.Data()
		}
	}
	for _, d := range bRestDims {
		if cv, ok := b.coords[d]; ok {
			coords[d] = cv.Data()
		}
	}
	return NewDataArray(resDims, resShape, resData, coords, a.name)
}

// restOf renvoie dimensions, forme et strides d'une variable en excluant l'axe.
func restOf[T Number](v *Variable[T], axis int) ([]string, []int, []int) {
	st := v.strides()
	var dims []string
	var shape, strides []int
	for i := range v.dims {
		if i == axis {
			continue
		}
		dims = append(dims, v.dims[i])
		shape = append(shape, v.shape[i])
		strides = append(strides, st[i])
	}
	return dims, shape, strides
}

// basesOf énumère les offsets plats de toutes les positions d'une sous-forme.
func basesOf(shape, strides []int) []int {
	n := product(shape)
	out := make([]int, n)
	counter := make([]int, len(shape))
	for p := 0; p < n; p++ {
		base := 0
		for i := range shape {
			base += counter[i] * strides[i]
		}
		out[p] = base
		for k := len(shape) - 1; k >= 0; k-- {
			counter[k]++
			if counter[k] < shape[k] {
				break
			}
			counter[k] = 0
		}
	}
	return out
}
