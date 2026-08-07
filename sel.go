package xarray

import (
	"fmt"
	"math"
)

// Indexation par label enrichie, complétant Sel (match exact) : plus proche
// voisin, plage d'étiquettes, liste d'étiquettes.

// SelNearest sélectionne, le long de dim, la position dont l'étiquette est la
// plus proche de label (la dimension est réduite, comme Sel/Isel).
func (da *DataArray[T]) SelNearest(dim string, label T) (*DataArray[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée pour la dimension %q", dim)
	}
	if len(cv.data) == 0 {
		return nil, fmt.Errorf("xarray: coordonnée %q vide", dim)
	}
	best := 0
	bestDist := math.Abs(float64(cv.data[0]) - float64(label))
	for i, l := range cv.data[1:] {
		d := math.Abs(float64(l) - float64(label))
		if d < bestDist {
			bestDist, best = d, i+1
		}
	}
	return da.Isel(dim, best)
}

// SelRange conserve, le long de dim, les positions dont l'étiquette est dans
// l'intervalle [lo, hi] (bornes incluses). La dimension est conservée.
func (da *DataArray[T]) SelRange(dim string, lo, hi T) (*DataArray[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée pour la dimension %q", dim)
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	var idx []int
	for i, l := range cv.data {
		if l >= lo && l <= hi {
			idx = append(idx, i)
		}
	}
	if len(idx) == 0 {
		return nil, fmt.Errorf("xarray: aucune étiquette de %q dans [%v, %v]", dim, lo, hi)
	}
	return da.takeAlong(dim, idx)
}

// SelMany conserve, le long de dim, les positions correspondant exactement aux
// étiquettes fournies (dans l'ordre donné). La dimension est conservée.
func (da *DataArray[T]) SelMany(dim string, labels []T) (*DataArray[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée pour la dimension %q", dim)
	}
	pos := make(map[T]int, len(cv.data))
	for i, l := range cv.data {
		pos[l] = i
	}
	idx := make([]int, len(labels))
	for i, l := range labels {
		p, found := pos[l]
		if !found {
			return nil, fmt.Errorf("xarray: étiquette %v absente de la coordonnée %q", l, dim)
		}
		idx[i] = p
	}
	return da.takeAlong(dim, idx)
}
