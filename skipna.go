package xarray

import "math"

// Réductions « skipna » : les valeurs NaN sont ignorées, comme le comportement
// par défaut de xarray (Python). Pour les types entiers, math.IsNaN est toujours
// faux, donc ces variantes se comportent comme les réductions ordinaires.

func isNaNT[T Number](x T) bool { return math.IsNaN(float64(x)) }

func sumSkipNA[T Number](s []T) T {
	var t T
	for _, x := range s {
		if !isNaNT(x) {
			t += x
		}
	}
	return t
}

func meanSkipNA[T Number](s []T) float64 {
	var t float64
	n := 0
	for _, x := range s {
		if !isNaNT(x) {
			t += float64(x)
			n++
		}
	}
	if n == 0 {
		return math.NaN()
	}
	return t / float64(n)
}

func minSkipNA[T Number](s []T) T {
	var m T
	seen := false
	for _, x := range s {
		if isNaNT(x) {
			continue
		}
		if !seen || x < m {
			m, seen = x, true
		}
	}
	return m
}

func maxSkipNA[T Number](s []T) T {
	var m T
	seen := false
	for _, x := range s {
		if isNaNT(x) {
			continue
		}
		if !seen || x > m {
			m, seen = x, true
		}
	}
	return m
}

// --- Réductions globales skipna ---------------------------------------------

// SumSkipNA renvoie la somme des éléments non-NaN.
func (da *DataArray[T]) SumSkipNA() T { return sumSkipNA(da.variable.data) }

// MeanSkipNA renvoie la moyenne des éléments non-NaN (NaN si aucun).
func (da *DataArray[T]) MeanSkipNA() float64 { return meanSkipNA(da.variable.data) }

// MinSkipNA renvoie le minimum des éléments non-NaN.
func (da *DataArray[T]) MinSkipNA() T { return minSkipNA(da.variable.data) }

// MaxSkipNA renvoie le maximum des éléments non-NaN.
func (da *DataArray[T]) MaxSkipNA() T { return maxSkipNA(da.variable.data) }

// --- Réductions par axe skipna ----------------------------------------------

// SumAxisSkipNA réduit dim par somme en ignorant les NaN.
func (da *DataArray[T]) SumAxisSkipNA(dim string) (*DataArray[T], error) {
	return reduceAxisDA[T, T](da, dim, sumSkipNA[T])
}

// MeanAxisSkipNA réduit dim par moyenne (float64) en ignorant les NaN.
func (da *DataArray[T]) MeanAxisSkipNA(dim string) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, meanSkipNA[T])
}

// MinAxisSkipNA réduit dim par minimum en ignorant les NaN.
func (da *DataArray[T]) MinAxisSkipNA(dim string) (*DataArray[T], error) {
	return reduceAxisDA[T, T](da, dim, minSkipNA[T])
}

// MaxAxisSkipNA réduit dim par maximum en ignorant les NaN.
func (da *DataArray[T]) MaxAxisSkipNA(dim string) (*DataArray[T], error) {
	return reduceAxisDA[T, T](da, dim, maxSkipNA[T])
}
