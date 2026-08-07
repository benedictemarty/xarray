package xarray

import (
	"fmt"
	"math"
	"sort"
)

// Réductions : indices des extrema, quantiles, produit cumulé.

func argMinSlice[T Number](s []T) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	best := 0
	for i, x := range s[1:] {
		if x < s[best] {
			best = i + 1
		}
	}
	return float64(best)
}

func argMaxSlice[T Number](s []T) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	best := 0
	for i, x := range s[1:] {
		if x > s[best] {
			best = i + 1
		}
	}
	return float64(best)
}

// quantileSlice calcule le quantile q ∈ [0,1] par interpolation linéaire entre
// les rangs (méthode « linear » de NumPy).
func quantileSlice[T Number](s []T, q float64) float64 {
	n := len(s)
	if n == 0 {
		return math.NaN()
	}
	c := make([]float64, n)
	for i, x := range s {
		c[i] = float64(x)
	}
	sort.Float64s(c)
	if q <= 0 {
		return c[0]
	}
	if q >= 1 {
		return c[n-1]
	}
	pos := q * float64(n-1)
	lo := int(math.Floor(pos))
	frac := pos - float64(lo)
	if lo+1 >= n {
		return c[n-1]
	}
	return c[lo] + (c[lo+1]-c[lo])*frac
}

// --- Global -----------------------------------------------------------------

// Quantile renvoie le quantile q ∈ [0,1] de tous les éléments.
func (da *DataArray[T]) Quantile(q float64) float64 {
	return quantileSlice(da.variable.data, q)
}

// --- Par axe ----------------------------------------------------------------

// ArgMinAxis renvoie, le long de dim, l'indice (float64) du minimum.
func (da *DataArray[T]) ArgMinAxis(dim string) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, argMinSlice[T])
}

// ArgMaxAxis renvoie, le long de dim, l'indice (float64) du maximum.
func (da *DataArray[T]) ArgMaxAxis(dim string) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, argMaxSlice[T])
}

// QuantileAxis réduit dim par quantile q (float64).
func (da *DataArray[T]) QuantileAxis(dim string, q float64) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, func(s []T) float64 { return quantileSlice(s, q) })
}

// --- Produit cumulé ---------------------------------------------------------

// Cumprod renvoie le produit cumulé le long de dim (même forme).
func (da *DataArray[T]) Cumprod(dim string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	out := da.clone()
	stAxis := out.variable.strides()[axis]
	dimLen := out.variable.shape[axis]
	data := out.variable.data
	out.forEachLine(axis, func(base int) {
		var acc T = 1
		for k := 0; k < dimLen; k++ {
			pos := base + k*stAxis
			acc *= data[pos]
			data[pos] = acc
		}
	})
	return out, nil
}
