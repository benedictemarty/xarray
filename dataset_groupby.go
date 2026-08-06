package xarray

import (
	"fmt"
	"sort"
)

// DatasetGroupBy représente un regroupement d'un Dataset le long d'une dimension,
// par les valeurs (répétées) de la coordonnée partagée de cette dimension.
// L'agrégation est propagée à toutes les variables portant la dimension ; les
// autres sont conservées (converties au type de sortie si nécessaire).
type DatasetGroupBy[T Number] struct {
	ds     *Dataset[T]
	dim    string
	labels []T
	groups [][]int
}

// GroupBy construit un regroupement du Dataset le long de dim, via la
// coordonnée partagée de dim.
func (ds *Dataset[T]) GroupBy(dim string) (*DatasetGroupBy[T], error) {
	cv, ok := ds.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: groupby impossible : aucune coordonnée pour la dimension %q", dim)
	}
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
	return &DatasetGroupBy[T]{ds: ds, dim: dim, labels: labels, groups: groups}, nil
}

// Groups renvoie le nombre de groupes.
func (g *DatasetGroupBy[T]) Groups() int { return len(g.labels) }

// dsGroupReduce applique reducer aux variables portant la dimension groupée et
// convertit les autres vers R, puis reconstruit un Dataset[R].
func dsGroupReduce[T, R Number](g *DatasetGroupBy[T], reducer func(*DataArray[T]) (*DataArray[R], error)) (*Dataset[R], error) {
	next := make(map[string]*DataArray[R], len(g.ds.vars))
	for name, da := range g.ds.vars {
		if da.HasDim(g.dim) {
			r, err := groupReduceOn(da, g.dim, g.groups, g.labels, reducer)
			if err != nil {
				return nil, err
			}
			next[name] = r
		} else {
			next[name] = convertDataArray[T, R](da)
		}
	}
	return NewDataset(next)
}

// Sum agrège chaque groupe par somme.
func (g *DatasetGroupBy[T]) Sum() (*Dataset[T], error) {
	return dsGroupReduce[T, T](g, func(d *DataArray[T]) (*DataArray[T], error) { return d.SumAxis(g.dim) })
}

// Mean agrège chaque groupe par moyenne (résultat en float64).
func (g *DatasetGroupBy[T]) Mean() (*Dataset[float64], error) {
	return dsGroupReduce[T, float64](g, func(d *DataArray[T]) (*DataArray[float64], error) { return d.MeanAxis(g.dim) })
}

// Min agrège chaque groupe par minimum.
func (g *DatasetGroupBy[T]) Min() (*Dataset[T], error) {
	return dsGroupReduce[T, T](g, func(d *DataArray[T]) (*DataArray[T], error) { return d.MinAxis(g.dim) })
}

// Max agrège chaque groupe par maximum.
func (g *DatasetGroupBy[T]) Max() (*Dataset[T], error) {
	return dsGroupReduce[T, T](g, func(d *DataArray[T]) (*DataArray[T], error) { return d.MaxAxis(g.dim) })
}
