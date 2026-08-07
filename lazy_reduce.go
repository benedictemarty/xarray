package xarray

import (
	"fmt"
	"math"
)

// Réductions par axe sur un LazyArray, en streaming (un chunk à la fois). Le
// résultat (plus petit) est matérialisé en DataArray[float64].
//
// Portée : tableaux 1D/2D (cohérent avec les sources chunk/zarr). Réduire l'axe 0
// (celui du découpage) accumule entre chunks ; réduire l'axe 1 réduit à
// l'intérieur de chaque bloc.

func indexOfDim(dims []string, dim string) int {
	for i, d := range dims {
		if d == dim {
			return i
		}
	}
	return -1
}

func (l *LazyArray) reduceAxisLazy(dim string, init float64, comb func(a, x float64) float64, mean bool) (*DataArray[float64], error) {
	shape := l.src.Shape()
	dims := l.src.Dims()
	axis := indexOfDim(dims, dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	if len(shape) > 2 {
		return nil, fmt.Errorf("xarray: réduction lazy par axe limitée aux tableaux 1D/2D")
	}
	srcCoords := l.src.Coords()
	nc := l.src.NumChunks()

	// Cas 1D : réduction totale -> scalaire (0 dimension).
	if len(shape) == 1 {
		acc := init
		for i := 0; i < nc; i++ {
			data, err := l.src.ChunkData(i)
			if err != nil {
				return nil, err
			}
			l.apply(data)
			for _, x := range data {
				acc = comb(acc, x)
			}
		}
		if mean {
			if shape[0] == 0 {
				acc = math.NaN()
			} else {
				acc /= float64(shape[0])
			}
		}
		return NewDataArray([]string{}, []int{}, []float64{acc}, nil, l.name)
	}

	// Cas 2D.
	R, C := shape[0], shape[1]

	if axis == 0 {
		// Accumulation par colonne (entre chunks).
		acc := make([]float64, C)
		for j := range acc {
			acc[j] = init
		}
		for i := 0; i < nc; i++ {
			data, err := l.src.ChunkData(i)
			if err != nil {
				return nil, err
			}
			l.apply(data)
			rows := len(data) / C
			for li := 0; li < rows; li++ {
				for j := 0; j < C; j++ {
					acc[j] = comb(acc[j], data[li*C+j])
				}
			}
		}
		if mean {
			for j := range acc {
				acc[j] /= float64(R)
			}
		}
		coords := map[string][]float64{}
		if c, ok := srcCoords[dims[1]]; ok {
			coords[dims[1]] = c
		}
		return NewDataArray([]string{dims[1]}, []int{C}, acc, coords, l.name)
	}

	// axis == 1 : réduction par ligne (à l'intérieur de chaque bloc).
	res := make([]float64, R)
	for i := 0; i < nc; i++ {
		start, _ := l.src.ChunkRows(i)
		data, err := l.src.ChunkData(i)
		if err != nil {
			return nil, err
		}
		l.apply(data)
		rows := len(data) / C
		for li := 0; li < rows; li++ {
			a := init
			for j := 0; j < C; j++ {
				a = comb(a, data[li*C+j])
			}
			if mean {
				a /= float64(C)
			}
			res[start+li] = a
		}
	}
	coords := map[string][]float64{}
	if c, ok := srcCoords[dims[0]]; ok {
		coords[dims[0]] = c
	}
	return NewDataArray([]string{dims[0]}, []int{R}, res, coords, l.name)
}

// SumAxis réduit dim par somme (streaming).
func (l *LazyArray) SumAxis(dim string) (*DataArray[float64], error) {
	return l.reduceAxisLazy(dim, 0, func(a, x float64) float64 { return a + x }, false)
}

// MeanAxis réduit dim par moyenne (streaming).
func (l *LazyArray) MeanAxis(dim string) (*DataArray[float64], error) {
	return l.reduceAxisLazy(dim, 0, func(a, x float64) float64 { return a + x }, true)
}

// MinAxis réduit dim par minimum (streaming).
func (l *LazyArray) MinAxis(dim string) (*DataArray[float64], error) {
	return l.reduceAxisLazy(dim, math.Inf(1), math.Min, false)
}

// MaxAxis réduit dim par maximum (streaming).
func (l *LazyArray) MaxAxis(dim string) (*DataArray[float64], error) {
	return l.reduceAxisLazy(dim, math.Inf(-1), math.Max, false)
}
