package xarray

import (
	"fmt"
	"math"
)

// Rolling représente une fenêtre glissante le long d'une dimension, à la manière
// de xarray.DataArray.rolling. La fenêtre est « trailing » : la valeur à
// l'indice i agrège les éléments [i-window+1 … i]. Les positions dont la fenêtre
// est incomplète (i < window-1) valent NaN.
//
// Les agrégations renvoient un DataArray[float64] de MÊME forme (les NaN de bord
// imposent le type flottant, comme dans xarray).
type Rolling[T Number] struct {
	da     *DataArray[T]
	dim    string
	window int
}

// Rolling construit une fenêtre glissante de taille window le long de dim.
func (da *DataArray[T]) Rolling(dim string, window int) (*Rolling[T], error) {
	if da.variable.dimIndex(dim) == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	if window < 1 {
		return nil, fmt.Errorf("xarray: taille de fenêtre invalide %d", window)
	}
	return &Rolling[T]{da: da, dim: dim, window: window}, nil
}

func rollingApply[T Number](da *DataArray[T], dim string, window int, reducer func([]float64) float64) (*DataArray[float64], error) {
	axis := da.variable.dimIndex(dim)
	shape := da.variable.Shape()
	dimLen := shape[axis]
	src := da.variable.data
	strides := da.variable.strides()
	stAxis := strides[axis]

	out := make([]float64, len(src))

	// Espace des dimensions autres que l'axe glissant.
	outerShape := make([]int, 0, len(shape)-1)
	for i, s := range shape {
		if i == axis {
			continue
		}
		outerShape = append(outerShape, s)
	}
	counter := make([]int, len(outerShape))
	win := make([]float64, window)
	nOuter := product(outerShape)

	for o := 0; o < nOuter; o++ {
		base := 0
		j := 0
		for i := range shape {
			if i == axis {
				continue
			}
			base += counter[j] * strides[i]
			j++
		}
		for k := 0; k < dimLen; k++ {
			pos := base + k*stAxis
			if k < window-1 {
				out[pos] = math.NaN()
				continue
			}
			for w := 0; w < window; w++ {
				win[w] = float64(src[base+(k-window+1+w)*stAxis])
			}
			out[pos] = reducer(win)
		}
		incrementCounter(counter, outerShape)
	}

	// Coordonnées conservées (converties en float64).
	coords := make(map[string][]float64, len(da.coords))
	for name, cv := range da.coords {
		lbl := make([]float64, len(cv.data))
		for i, x := range cv.data {
			lbl[i] = float64(x)
		}
		coords[name] = lbl
	}
	return NewDataArray(da.variable.Dims(), shape, out, coords, da.name)
}

// Mean : moyenne mobile.
func (r *Rolling[T]) Mean() (*DataArray[float64], error) {
	return rollingApply(r.da, r.dim, r.window, func(w []float64) float64 {
		var s float64
		for _, x := range w {
			s += x
		}
		return s / float64(len(w))
	})
}

// Sum : somme mobile.
func (r *Rolling[T]) Sum() (*DataArray[float64], error) {
	return rollingApply(r.da, r.dim, r.window, func(w []float64) float64 {
		var s float64
		for _, x := range w {
			s += x
		}
		return s
	})
}

// Min : minimum mobile.
func (r *Rolling[T]) Min() (*DataArray[float64], error) {
	return rollingApply(r.da, r.dim, r.window, func(w []float64) float64 {
		m := w[0]
		for _, x := range w[1:] {
			if x < m {
				m = x
			}
		}
		return m
	})
}

// Max : maximum mobile.
func (r *Rolling[T]) Max() (*DataArray[float64], error) {
	return rollingApply(r.da, r.dim, r.window, func(w []float64) float64 {
		m := w[0]
		for _, x := range w[1:] {
			if x > m {
				m = x
			}
		}
		return m
	})
}
