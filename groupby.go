package xarray

import (
	"fmt"
	"reflect"
	"sort"
)

// GroupBy représente un regroupement d'un DataArray le long d'une dimension,
// par les valeurs (éventuellement répétées) de la coordonnée de cette dimension.
//
// Exemple : une dimension « temps » dont la coordonnée vaut [1, 1, 2, 2] (des
// numéros de mois répétés) peut être regroupée par mois, chaque groupe étant
// ensuite agrégé (somme, moyenne…).
type GroupBy[T Number] struct {
	da     *DataArray[T]
	dim    string
	labels []T     // valeurs de groupe uniques, triées
	groups [][]int // positions le long de dim pour chaque groupe
}

// GroupBy construit un regroupement le long de dim, en s'appuyant sur la
// coordonnée de dim. La dimension résultante conserve le nom dim mais ne porte
// plus que les étiquettes uniques.
func (da *DataArray[T]) GroupBy(dim string) (*GroupBy[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: groupby impossible : aucune coordonnée pour la dimension %q", dim)
	}
	// Regroupement des positions par valeur de coordonnée.
	positions := map[T][]int{}
	for i, l := range cv.data {
		positions[l] = append(positions[l], i)
	}
	labels := make([]T, 0, len(positions))
	for l := range positions {
		labels = append(labels, l)
	}
	sort.Slice(labels, func(i, j int) bool { return labels[i] < labels[j] })

	groups := make([][]int, len(labels))
	for i, l := range labels {
		groups[i] = positions[l]
	}
	return &GroupBy[T]{da: da, dim: dim, labels: labels, groups: groups}, nil
}

// Groups renvoie le nombre de groupes.
func (g *GroupBy[T]) Groups() int { return len(g.labels) }

// Labels renvoie les étiquettes de groupe (uniques, triées).
func (g *GroupBy[T]) Labels() []T { return append([]T(nil), g.labels...) }

// groupReduce applique un réducteur (réduisant la dimension groupée) à chaque
// groupe puis empile les résultats sur la dimension groupée. Le type de sortie R
// peut différer de T (ex. moyenne -> float64).
func groupReduce[T, R Number](g *GroupBy[T], reducer func(*DataArray[T]) (*DataArray[R], error)) (*DataArray[R], error) {
	slices := make([]*DataArray[R], len(g.groups))
	for i, idxs := range g.groups {
		sub, err := g.da.takeAlong(g.dim, idxs)
		if err != nil {
			return nil, err
		}
		r, err := reducer(sub)
		if err != nil {
			return nil, err
		}
		slices[i] = r
	}
	labelsR := make([]R, len(g.labels))
	for i, l := range g.labels {
		labelsR[i] = convertNum[T, R](l)
	}
	return stackDim(slices, g.dim, labelsR)
}

// stackDim empile des tranches de forme identique le long d'une nouvelle
// dimension placée en tête, dont la coordonnée est labels.
func stackDim[R Number](slices []*DataArray[R], newDim string, labels []R) (*DataArray[R], error) {
	if len(slices) == 0 {
		return nil, fmt.Errorf("xarray: empilement d'un ensemble vide")
	}
	first := slices[0]
	innerDims := first.variable.dims
	innerShape := first.variable.shape

	dims := append([]string{newDim}, innerDims...)
	shape := append([]int{len(slices)}, innerShape...)

	data := make([]R, 0, len(slices)*first.variable.Size())
	for _, s := range slices {
		if !reflect.DeepEqual(s.variable.dims, innerDims) || !reflect.DeepEqual(s.variable.shape, innerShape) {
			return nil, fmt.Errorf("xarray: empilement de formes incompatibles")
		}
		data = append(data, s.variable.data...)
	}

	coords := map[string][]R{newDim: labels}
	for k, cv := range first.coords {
		coords[k] = cv.Data()
	}
	return NewDataArray(dims, shape, data, coords, first.name)
}

// Sum agrège chaque groupe par somme.
func (g *GroupBy[T]) Sum() (*DataArray[T], error) {
	return groupReduce[T, T](g, func(d *DataArray[T]) (*DataArray[T], error) { return d.SumAxis(g.dim) })
}

// Mean agrège chaque groupe par moyenne (résultat en float64).
func (g *GroupBy[T]) Mean() (*DataArray[float64], error) {
	return groupReduce[T, float64](g, func(d *DataArray[T]) (*DataArray[float64], error) { return d.MeanAxis(g.dim) })
}

// Min agrège chaque groupe par minimum.
func (g *GroupBy[T]) Min() (*DataArray[T], error) {
	return groupReduce[T, T](g, func(d *DataArray[T]) (*DataArray[T], error) { return d.MinAxis(g.dim) })
}

// Max agrège chaque groupe par maximum.
func (g *GroupBy[T]) Max() (*DataArray[T], error) {
	return groupReduce[T, T](g, func(d *DataArray[T]) (*DataArray[T], error) { return d.MaxAxis(g.dim) })
}
