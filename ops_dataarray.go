package xarray

import (
	"fmt"
	"math"
)

// --- Réducteurs de base sur une tranche -------------------------------------

func sumSlice(s []float64) float64 {
	var t float64
	for _, x := range s {
		t += x
	}
	return t
}

func meanSlice(s []float64) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	return sumSlice(s) / float64(len(s))
}

func minSlice(s []float64) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	m := s[0]
	for _, x := range s[1:] {
		if x < m {
			m = x
		}
	}
	return m
}

func maxSlice(s []float64) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	m := s[0]
	for _, x := range s[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

// --- Transpose --------------------------------------------------------------

// Transpose réordonne les dimensions du DataArray selon newDims (permutation).
// Les coordonnées sont conservées.
func (da *DataArray) Transpose(newDims ...string) (*DataArray, error) {
	nv, err := da.variable.Transpose(newDims...)
	if err != nil {
		return nil, err
	}
	return &DataArray{variable: nv, coords: da.cloneCoords(), name: da.name}, nil
}

func (da *DataArray) cloneCoords() map[string]*Variable {
	coords := make(map[string]*Variable, len(da.coords))
	for k, cv := range da.coords {
		nc, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
		coords[k] = nc
	}
	return coords
}

// --- Réductions par axe -----------------------------------------------------

// reduceAxis applique un réducteur le long d'une dimension et retire sa
// coordonnée éventuelle du résultat.
func (da *DataArray) reduceAxis(dim string, reducer func([]float64) float64) (*DataArray, error) {
	nv, err := da.variable.reduceAxis(dim, reducer)
	if err != nil {
		return nil, err
	}
	coords := make(map[string]*Variable, len(da.coords))
	for k, cv := range da.coords {
		if k == dim {
			continue
		}
		nc, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
		coords[k] = nc
	}
	return &DataArray{variable: nv, coords: coords, name: da.name}, nil
}

// SumAxis réduit la dimension dim par somme.
func (da *DataArray) SumAxis(dim string) (*DataArray, error) {
	return da.reduceAxis(dim, sumSlice)
}

// MeanAxis réduit la dimension dim par moyenne.
func (da *DataArray) MeanAxis(dim string) (*DataArray, error) {
	return da.reduceAxis(dim, meanSlice)
}

// MinAxis réduit la dimension dim par minimum.
func (da *DataArray) MinAxis(dim string) (*DataArray, error) {
	return da.reduceAxis(dim, minSlice)
}

// MaxAxis réduit la dimension dim par maximum.
func (da *DataArray) MaxAxis(dim string) (*DataArray, error) {
	return da.reduceAxis(dim, maxSlice)
}

// --- Alignement -------------------------------------------------------------

// align aligne deux DataArrays sur leurs dimensions communes disposant de
// coordonnées des deux côtés : seules les étiquettes présentes dans les deux
// (jointure interne) sont conservées, dans l'ordre du premier opérande.
func align(a, b *DataArray) (*DataArray, *DataArray, error) {
	a2, b2 := a, b
	for _, dim := range a.variable.dims {
		ca, okA := a2.coords[dim]
		cb, okB := b2.coords[dim]
		if !okA || !okB {
			continue // pas de coordonnées des deux côtés : pas d'alignement
		}
		// Positions communes.
		posB := make(map[float64]int, len(cb.data))
		for i, l := range cb.data {
			posB[l] = i
		}
		var idxA, idxB []int
		var labels []float64
		for i, l := range ca.data {
			if j, ok := posB[l]; ok {
				idxA = append(idxA, i)
				idxB = append(idxB, j)
				labels = append(labels, l)
			}
		}
		if len(labels) == 0 {
			return nil, nil, fmt.Errorf("xarray: aucune étiquette commune sur la dimension %q", dim)
		}
		na, err := a2.takeAlong(dim, idxA)
		if err != nil {
			return nil, nil, err
		}
		nb, err := b2.takeAlong(dim, idxB)
		if err != nil {
			return nil, nil, err
		}
		a2, b2 = na, nb
	}
	return a2, b2, nil
}

// takeAlong sélectionne plusieurs positions le long de dim en conservant la
// dimension et en réindexant sa coordonnée.
func (da *DataArray) takeAlong(dim string, indices []int) (*DataArray, error) {
	nv, err := da.variable.take(dim, indices)
	if err != nil {
		return nil, err
	}
	coords := make(map[string]*Variable, len(da.coords))
	for k, cv := range da.coords {
		if k == dim {
			nc, err := cv.take(dim, indices)
			if err != nil {
				return nil, err
			}
			coords[k] = nc
			continue
		}
		nc, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
		coords[k] = nc
	}
	return &DataArray{variable: nv, coords: coords, name: da.name}, nil
}

// --- Arithmétique -----------------------------------------------------------

// binary applique une opération binaire élément par élément entre deux
// DataArrays, avec alignement automatique sur les coordonnées puis broadcasting
// par nom de dimension.
func (da *DataArray) binary(other *DataArray, fn func(x, y float64) float64) (*DataArray, error) {
	a, b, err := align(da, other)
	if err != nil {
		return nil, err
	}
	nv, err := binaryOp(a.variable, b.variable, fn)
	if err != nil {
		return nil, err
	}
	// Coordonnées du résultat : celles de a en priorité, complétées par b.
	coords := make(map[string]*Variable, len(nv.dims))
	for _, dim := range nv.dims {
		if cv, ok := a.coords[dim]; ok {
			nc, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
			coords[dim] = nc
		} else if cv, ok := b.coords[dim]; ok {
			nc, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
			coords[dim] = nc
		}
	}
	return &DataArray{variable: nv, coords: coords, name: da.name}, nil
}

// Add renvoie da + other (avec alignement et broadcasting).
func (da *DataArray) Add(other *DataArray) (*DataArray, error) {
	return da.binary(other, func(x, y float64) float64 { return x + y })
}

// Sub renvoie da - other.
func (da *DataArray) Sub(other *DataArray) (*DataArray, error) {
	return da.binary(other, func(x, y float64) float64 { return x - y })
}

// Mul renvoie da * other.
func (da *DataArray) Mul(other *DataArray) (*DataArray, error) {
	return da.binary(other, func(x, y float64) float64 { return x * y })
}

// Div renvoie da / other.
func (da *DataArray) Div(other *DataArray) (*DataArray, error) {
	return da.binary(other, func(x, y float64) float64 { return x / y })
}

// --- Opérations scalaires ---------------------------------------------------

// AddScalar ajoute s à chaque élément.
func (da *DataArray) AddScalar(s float64) *DataArray {
	return da.mapScalar(func(x float64) float64 { return x + s })
}

// MulScalar multiplie chaque élément par s.
func (da *DataArray) MulScalar(s float64) *DataArray {
	return da.mapScalar(func(x float64) float64 { return x * s })
}

// mapScalar applique fn à chaque élément en conservant dimensions et coordonnées.
func (da *DataArray) mapScalar(fn func(float64) float64) *DataArray {
	nv := da.variable.mapScalar(fn)
	return &DataArray{variable: nv, coords: da.cloneCoords(), name: da.name}
}
