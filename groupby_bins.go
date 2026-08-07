package xarray

import (
	"fmt"
	"sort"
)

// GroupByBins regroupe le long de dim par intervalles définis explicitement par
// edges (n+1 bornes → n intervalles), à la manière de `groupby_bins` de xarray.
// Un intervalle est [edges[k], edges[k+1]) ; le dernier est fermé à droite. Les
// valeurs hors des bornes sont ignorées. L'étiquette d'un bin est sa borne gauche.

func binEdgesGroups[T Number](labels []T, edges []T) ([]T, [][]int, error) {
	if len(edges) < 2 {
		return nil, nil, fmt.Errorf("xarray: au moins 2 bornes requises pour GroupByBins")
	}
	nb := len(edges) - 1
	findBin := func(v T) int {
		for b := 0; b < nb; b++ {
			if v >= edges[b] && (v < edges[b+1] || (b == nb-1 && v <= edges[nb])) {
				return b
			}
		}
		return -1
	}
	m := map[int][]int{}
	for i, v := range labels {
		if b := findBin(v); b >= 0 {
			m[b] = append(m[b], i)
		}
	}
	binIdx := make([]int, 0, len(m))
	for b := range m {
		binIdx = append(binIdx, b)
	}
	sort.Ints(binIdx)

	blabels := make([]T, len(binIdx))
	groups := make([][]int, len(binIdx))
	for i, b := range binIdx {
		blabels[i] = edges[b] // borne gauche du bin
		groups[i] = m[b]
	}
	return blabels, groups, nil
}

// GroupByBins construit un regroupement par intervalles arbitraires sur dim.
func (da *DataArray[T]) GroupByBins(dim string, edges []T) (*Resample[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée pour la dimension %q", dim)
	}
	labels, groups, err := binEdgesGroups(cv.data, edges)
	if err != nil {
		return nil, err
	}
	return &Resample[T]{da: da, dim: dim, labels: labels, groups: groups}, nil
}

// GroupByBins construit un regroupement par intervalles arbitraires sur un
// Dataset (propagé aux variables portant la dimension).
func (ds *Dataset[T]) GroupByBins(dim string, edges []T) (*DatasetGroupBy[T], error) {
	cv, ok := ds.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée %q dans le dataset", dim)
	}
	labels, groups, err := binEdgesGroups(cv.data, edges)
	if err != nil {
		return nil, err
	}
	return &DatasetGroupBy[T]{ds: ds, dim: dim, labels: labels, groups: groups}, nil
}
