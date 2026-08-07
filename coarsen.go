package xarray

import "fmt"

// Coarsen agrège des blocs **non chevauchants** de `factor` éléments le long de
// dim (downsampling), à la manière de `da.coarsen(dim=factor)` de xarray. Si la
// taille n'est pas divisible, le reste est ignoré (boundary="trim").
//
// Renvoie un Resample (agrégations Sum/Mean/Min/Max). L'étiquette d'un bloc est la
// borne gauche de sa coordonnée (ou l'indice de bloc à défaut de coordonnée).

func coarsenGroups[T Number](dimLen, factor int, coord []T) ([]T, [][]int, error) {
	if factor < 1 {
		return nil, nil, fmt.Errorf("xarray: facteur de coarsen invalide %d", factor)
	}
	nB := dimLen / factor
	if nB == 0 {
		return nil, nil, fmt.Errorf("xarray: facteur %d trop grand pour la dimension de taille %d", factor, dimLen)
	}
	labels := make([]T, nB)
	groups := make([][]int, nB)
	for b := 0; b < nB; b++ {
		idx := make([]int, factor)
		for j := 0; j < factor; j++ {
			idx[j] = b*factor + j
		}
		groups[b] = idx
		if coord != nil {
			labels[b] = coord[b*factor]
		} else {
			labels[b] = T(b)
		}
	}
	return labels, groups, nil
}

// Coarsen construit un downsampling par blocs de taille factor le long de dim.
func (da *DataArray[T]) Coarsen(dim string, factor int) (*Resample[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	var coord []T
	if cv, ok := da.coords[dim]; ok {
		coord = cv.data
	}
	labels, groups, err := coarsenGroups(da.variable.shape[axis], factor, coord)
	if err != nil {
		return nil, err
	}
	return &Resample[T]{da: da, dim: dim, labels: labels, groups: groups}, nil
}

// Coarsen construit un downsampling par blocs sur un Dataset (propagé aux
// variables portant la dimension).
func (ds *Dataset[T]) Coarsen(dim string, factor int) (*DatasetGroupBy[T], error) {
	n, ok := ds.dims[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: dimension %q absente du dataset", dim)
	}
	var coord []T
	if cv, ok := ds.coords[dim]; ok {
		coord = cv.data
	}
	labels, groups, err := coarsenGroups(n, factor, coord)
	if err != nil {
		return nil, err
	}
	return &DatasetGroupBy[T]{ds: ds, dim: dim, labels: labels, groups: groups}, nil
}
