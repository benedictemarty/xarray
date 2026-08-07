package xarray

import (
	"fmt"
	"math"
	"sort"
)

// Réductions statistiques et cumulatives supplémentaires.

func varSlice[T Number](s []T) float64 {
	n := len(s)
	if n == 0 {
		return math.NaN()
	}
	var mean float64
	for _, x := range s {
		mean += float64(x)
	}
	mean /= float64(n)
	var acc float64
	for _, x := range s {
		d := float64(x) - mean
		acc += d * d
	}
	return acc / float64(n) // variance de population (ddof=0), comme xarray
}

func stdSlice[T Number](s []T) float64 { return math.Sqrt(varSlice(s)) }

func medianSlice[T Number](s []T) float64 {
	n := len(s)
	if n == 0 {
		return math.NaN()
	}
	c := append([]T(nil), s...)
	sort.Slice(c, func(i, j int) bool { return c[i] < c[j] })
	if n%2 == 1 {
		return float64(c[n/2])
	}
	return (float64(c[n/2-1]) + float64(c[n/2])) / 2
}

// --- Réductions globales ----------------------------------------------------

// Var renvoie la variance (population, ddof=0) de tous les éléments.
func (da *DataArray[T]) Var() float64 { return varSlice(da.variable.data) }

// Std renvoie l'écart-type (population) de tous les éléments.
func (da *DataArray[T]) Std() float64 { return stdSlice(da.variable.data) }

// Median renvoie la médiane de tous les éléments.
func (da *DataArray[T]) Median() float64 { return medianSlice(da.variable.data) }

// --- Réductions par axe -----------------------------------------------------

// VarAxis réduit dim par variance (float64).
func (da *DataArray[T]) VarAxis(dim string) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, varSlice[T])
}

// StdAxis réduit dim par écart-type (float64).
func (da *DataArray[T]) StdAxis(dim string) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, stdSlice[T])
}

// MedianAxis réduit dim par médiane (float64).
func (da *DataArray[T]) MedianAxis(dim string) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, medianSlice[T])
}

// --- Opérations cumulatives -------------------------------------------------

// Cumsum renvoie la somme cumulée le long de dim (même forme).
func (da *DataArray[T]) Cumsum(dim string) (*DataArray[T], error) {
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
		var acc T
		for k := 0; k < dimLen; k++ {
			pos := base + k*stAxis
			acc += data[pos]
			data[pos] = acc
		}
	})
	return out, nil
}

// Diff renvoie les différences successives le long de dim : out[k] = in[k+1] -
// in[k]. La dimension est réduite de 1 (la coordonnée conserve les positions
// 1..n-1, comme xarray).
func (da *DataArray[T]) Diff(dim string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	dimLen := da.variable.shape[axis]
	if dimLen < 1 {
		return nil, fmt.Errorf("xarray: dimension %q vide", dim)
	}
	upIdx := make([]int, dimLen-1)
	loIdx := make([]int, dimLen-1)
	for i := 0; i < dimLen-1; i++ {
		upIdx[i] = i + 1
		loIdx[i] = i
	}
	up, err := da.takeAlong(dim, upIdx)
	if err != nil {
		return nil, err
	}
	lo, err := da.takeAlong(dim, loIdx)
	if err != nil {
		return nil, err
	}
	// Soustraction positionnelle (sans alignement, car coordonnées décalées).
	nv, err := binaryOp(up.variable, lo.variable, func(a, b T) T { return a - b })
	if err != nil {
		return nil, err
	}
	return &DataArray[T]{variable: nv, coords: up.cloneCoords(), name: da.name}, nil
}
