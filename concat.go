package xarray

import (
	"fmt"
	"reflect"
)

// Concat concatène plusieurs DataArrays le long d'une dimension existante dim.
// Tous les tableaux doivent avoir les mêmes dimensions (noms et ordre) et les
// mêmes tailles sur toutes les dimensions autres que dim. La coordonnée de dim
// résultante est la concaténation des coordonnées correspondantes (si toutes
// présentes) ; les autres coordonnées proviennent du premier tableau.
func Concat[T Number](arrays []*DataArray[T], dim string) (*DataArray[T], error) {
	if len(arrays) == 0 {
		return nil, fmt.Errorf("xarray: concaténation d'un ensemble vide")
	}
	first := arrays[0]
	axis := first.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}

	// Validation de compatibilité et calcul de la taille cumulée sur l'axe.
	totalAxis := 0
	haveAllCoords := true
	for _, a := range arrays {
		if !reflect.DeepEqual(a.variable.dims, first.variable.dims) {
			return nil, fmt.Errorf("xarray: dimensions incompatibles pour la concaténation (%v vs %v)", a.variable.dims, first.variable.dims)
		}
		for i := range first.variable.shape {
			if i == axis {
				continue
			}
			if a.variable.shape[i] != first.variable.shape[i] {
				return nil, fmt.Errorf("xarray: taille incompatible sur %q pour la concaténation", first.variable.dims[i])
			}
		}
		totalAxis += a.variable.shape[axis]
		if _, ok := a.coords[dim]; !ok {
			haveAllCoords = false
		}
	}

	newShape := append([]int(nil), first.variable.shape...)
	newShape[axis] = totalAxis

	out := &Variable[T]{
		dims:  first.variable.Dims(),
		shape: newShape,
		data:  make([]T, product(newShape)),
		attrs: map[string]string{},
	}
	newStrides := out.strides()

	// Copie de chaque tableau à son décalage le long de l'axe.
	offset := 0
	for _, a := range arrays {
		counter := make([]int, len(a.variable.shape))
		for flatOld := range a.variable.data {
			flatNew := 0
			for i, c := range counter {
				pos := c
				if i == axis {
					pos = c + offset
				}
				flatNew += pos * newStrides[i]
			}
			out.data[flatNew] = a.variable.data[flatOld]
			for k := len(counter) - 1; k >= 0; k-- {
				counter[k]++
				if counter[k] < a.variable.shape[k] {
					break
				}
				counter[k] = 0
			}
		}
		offset += a.variable.shape[axis]
	}

	// Coordonnées.
	coords := map[string][]T{}
	if haveAllCoords {
		merged := make([]T, 0, totalAxis)
		for _, a := range arrays {
			merged = append(merged, a.coords[dim].data...)
		}
		coords[dim] = merged
	}
	for k, cv := range first.coords {
		if k == dim {
			continue
		}
		coords[k] = cv.Data()
	}

	da := &DataArray[T]{variable: out, name: first.name}
	// Reconstruit proprement via NewDataArray pour valider les coordonnées.
	return NewDataArray(da.variable.Dims(), da.variable.Shape(), da.variable.Data(), coords, first.name)
}

// Stack empile plusieurs DataArrays de forme identique le long d'une NOUVELLE
// dimension newDim placée en tête, dont la coordonnée est labels (une étiquette
// par tableau).
func Stack[T Number](arrays []*DataArray[T], newDim string, labels []T) (*DataArray[T], error) {
	if len(labels) != len(arrays) {
		return nil, fmt.Errorf("xarray: %d étiquette(s) pour %d tableau(x)", len(labels), len(arrays))
	}
	if len(arrays) > 0 && arrays[0].variable.dimIndex(newDim) != -1 {
		return nil, fmt.Errorf("xarray: la dimension %q existe déjà", newDim)
	}
	return stackDim(arrays, newDim, labels)
}
