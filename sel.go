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
	return da.Isel(dim, nearestIndex(cv.data, label))
}

// nearestIndex renvoie l'index de l'étiquette la plus proche de label.
// Précondition : data non vide.
func nearestIndex[T Number](data []T, label T) int {
	best := 0
	bestDist := math.Abs(float64(data[0]) - float64(label))
	for i, l := range data[1:] {
		d := math.Abs(float64(l) - float64(label))
		if d < bestDist {
			bestDist, best = d, i+1
		}
	}
	return best
}

// SelNearestMany sélectionne, le long de dim, la position la plus proche pour
// chacune des étiquettes fournies (dans l'ordre), en CONSERVANT la dimension.
// Équivalent de xarray sel(dim=[l1, l2, ...], method="nearest"), là où
// SelNearest (label scalaire) réduit la dimension comme sel(dim=l).
func (da *DataArray[T]) SelNearestMany(dim string, labels []T) (*DataArray[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée pour la dimension %q", dim)
	}
	if len(cv.data) == 0 {
		return nil, fmt.Errorf("xarray: coordonnée %q vide", dim)
	}
	if len(labels) == 0 {
		return nil, fmt.Errorf("xarray: aucune étiquette fournie pour %q", dim)
	}
	idx := make([]int, len(labels))
	for i, l := range labels {
		idx[i] = nearestIndex(cv.data, l)
	}
	return da.takeAlong(dim, idx)
}

// SelNearestKeep sélectionne la position la plus proche de label le long de dim
// en CONSERVANT la dimension (taille 1). Équivalent de xarray
// sel(dim=[label], method="nearest"). Utile pour les exports (CoverageJSON,
// EDR) qui exigent des axes explicites.
func (da *DataArray[T]) SelNearestKeep(dim string, label T) (*DataArray[T], error) {
	return da.SelNearestMany(dim, []T{label})
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
