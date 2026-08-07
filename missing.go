package xarray

import "fmt"

// Gestion des valeurs manquantes (NaN). Pour les types entiers, math.IsNaN est
// toujours faux : ces opérations sont alors sans effet.

// CountNA renvoie le nombre de valeurs NaN.
func (da *DataArray[T]) CountNA() int {
	n := 0
	for _, x := range da.variable.data {
		if isNaNT(x) {
			n++
		}
	}
	return n
}

// FillNA renvoie une copie où chaque NaN est remplacé par value.
func (da *DataArray[T]) FillNA(value T) *DataArray[T] {
	out := da.clone()
	for i, x := range out.variable.data {
		if isNaNT(x) {
			out.variable.data[i] = value
		}
	}
	return out
}

// forEachLine appelle fn(base) pour chaque position des dimensions autres que
// l'axe, base étant l'offset plat du premier élément (axe = 0) de la ligne.
func (da *DataArray[T]) forEachLine(axis int, fn func(base int)) {
	shape := da.variable.shape
	strides := da.variable.strides()
	outerShape := make([]int, 0, len(shape)-1)
	outerStrides := make([]int, 0, len(shape)-1)
	for i := range shape {
		if i == axis {
			continue
		}
		outerShape = append(outerShape, shape[i])
		outerStrides = append(outerStrides, strides[i])
	}
	counter := make([]int, len(outerShape))
	nOuter := product(outerShape)
	for o := 0; o < nOuter; o++ {
		base := 0
		for j := range outerShape {
			base += counter[j] * outerStrides[j]
		}
		fn(base)
		incrementCounter(counter, outerShape)
	}
}

// DropNA supprime, le long de dim, les positions dont la tranche contient au
// moins un NaN (how = "any", comme le défaut de xarray).
func (da *DataArray[T]) DropNA(dim string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	shape := da.variable.shape
	strides := da.variable.strides()
	stAxis := strides[axis]
	dimLen := shape[axis]
	src := da.variable.data

	hasNaN := make([]bool, dimLen)
	da.forEachLine(axis, func(base int) {
		for k := 0; k < dimLen; k++ {
			if isNaNT(src[base+k*stAxis]) {
				hasNaN[k] = true
			}
		}
	})

	keep := make([]int, 0, dimLen)
	for k := 0; k < dimLen; k++ {
		if !hasNaN[k] {
			keep = append(keep, k)
		}
	}
	if len(keep) == dimLen {
		return da.clone(), nil
	}
	return da.takeAlong(dim, keep)
}

// FFill propage la dernière valeur non-NaN vers l'avant, le long de dim (forward
// fill). Les NaN en tête (avant toute valeur valide) restent NaN.
func (da *DataArray[T]) FFill(dim string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	out := da.clone()
	shape := out.variable.shape
	stAxis := out.variable.strides()[axis]
	dimLen := shape[axis]
	data := out.variable.data

	out.forEachLine(axis, func(base int) {
		var last T
		haveLast := false
		for k := 0; k < dimLen; k++ {
			pos := base + k*stAxis
			if isNaNT(data[pos]) {
				if haveLast {
					data[pos] = last
				}
			} else {
				last = data[pos]
				haveLast = true
			}
		}
	})
	return out, nil
}

// BFill propage la prochaine valeur non-NaN vers l'arrière, le long de dim
// (backward fill). Les NaN en fin restent NaN.
func (da *DataArray[T]) BFill(dim string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	out := da.clone()
	shape := out.variable.shape
	stAxis := out.variable.strides()[axis]
	dimLen := shape[axis]
	data := out.variable.data

	out.forEachLine(axis, func(base int) {
		var next T
		haveNext := false
		for k := dimLen - 1; k >= 0; k-- {
			pos := base + k*stAxis
			if isNaNT(data[pos]) {
				if haveNext {
					data[pos] = next
				}
			} else {
				next = data[pos]
				haveNext = true
			}
		}
	})
	return out, nil
}
