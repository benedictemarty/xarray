package xarray

import "fmt"

// Restructuration des dimensions : suppression/ajout de dimensions de taille 1,
// renommage.

// Squeeze supprime la dimension dim si elle est de taille 1 (ainsi que sa
// coordonnée). Les données ne changent pas.
func (da *DataArray[T]) Squeeze(dim string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	if da.variable.shape[axis] != 1 {
		return nil, fmt.Errorf("xarray: la dimension %q n'est pas de taille 1 (%d)", dim, da.variable.shape[axis])
	}
	newDims := make([]string, 0, len(da.variable.dims)-1)
	newShape := make([]int, 0, len(da.variable.shape)-1)
	for i := range da.variable.dims {
		if i == axis {
			continue
		}
		newDims = append(newDims, da.variable.dims[i])
		newShape = append(newShape, da.variable.shape[i])
	}
	coords := map[string][]T{}
	for k, cv := range da.coords {
		if k == dim {
			continue
		}
		coords[k] = cv.Data()
	}
	return NewDataArray(newDims, newShape, da.variable.Data(), coords, da.name)
}

// ExpandDims insère une nouvelle dimension de taille 1 en tête (sans
// coordonnée). Les données ne changent pas.
func (da *DataArray[T]) ExpandDims(dim string) (*DataArray[T], error) {
	if da.variable.dimIndex(dim) != -1 {
		return nil, fmt.Errorf("xarray: la dimension %q existe déjà", dim)
	}
	newDims := append([]string{dim}, da.variable.Dims()...)
	newShape := append([]int{1}, da.variable.Shape()...)
	coords := map[string][]T{}
	for k, cv := range da.coords {
		coords[k] = cv.Data()
	}
	return NewDataArray(newDims, newShape, da.variable.Data(), coords, da.name)
}

// RenameDim renomme une dimension (et sa coordonnée éventuelle).
func (da *DataArray[T]) RenameDim(old, newName string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(old)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", old)
	}
	if da.variable.dimIndex(newName) != -1 {
		return nil, fmt.Errorf("xarray: la dimension %q existe déjà", newName)
	}
	newDims := da.variable.Dims()
	newDims[axis] = newName
	coords := map[string][]T{}
	for k, cv := range da.coords {
		key := k
		if k == old {
			key = newName
		}
		coords[key] = cv.Data()
	}
	return NewDataArray(newDims, da.variable.Shape(), da.variable.Data(), coords, da.name)
}
