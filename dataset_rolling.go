package xarray

import "fmt"

// DatasetRolling propage une fenêtre glissante à toutes les variables d'un
// Dataset portant la dimension visée.
type DatasetRolling[T Number] struct {
	ds     *Dataset[T]
	dim    string
	window int
}

// Rolling construit une fenêtre glissante de taille window le long de dim sur le
// Dataset.
func (ds *Dataset[T]) Rolling(dim string, window int) (*DatasetRolling[T], error) {
	if _, ok := ds.dims[dim]; !ok {
		return nil, fmt.Errorf("xarray: dimension %q absente du dataset", dim)
	}
	if window < 1 {
		return nil, fmt.Errorf("xarray: taille de fenêtre invalide %d", window)
	}
	return &DatasetRolling[T]{ds: ds, dim: dim, window: window}, nil
}

// dsRolling applique un agrégat glissant aux variables portant la dimension (les
// autres sont converties en float64), puis reconstruit un Dataset[float64].
func dsRolling[T Number](r *DatasetRolling[T], agg func(*Rolling[T]) (*DataArray[float64], error)) (*Dataset[float64], error) {
	next := make(map[string]*DataArray[float64], len(r.ds.vars))
	for name, da := range r.ds.vars {
		if da.HasDim(r.dim) {
			rr, err := da.Rolling(r.dim, r.window)
			if err != nil {
				return nil, err
			}
			m, err := agg(rr)
			if err != nil {
				return nil, err
			}
			next[name] = m
		} else {
			next[name] = convertDataArray[T, float64](da)
		}
	}
	return NewDataset(next)
}

// Mean : moyenne mobile sur le Dataset.
func (r *DatasetRolling[T]) Mean() (*Dataset[float64], error) {
	return dsRolling(r, func(rr *Rolling[T]) (*DataArray[float64], error) { return rr.Mean() })
}

// Sum : somme mobile.
func (r *DatasetRolling[T]) Sum() (*Dataset[float64], error) {
	return dsRolling(r, func(rr *Rolling[T]) (*DataArray[float64], error) { return rr.Sum() })
}

// Min : minimum mobile.
func (r *DatasetRolling[T]) Min() (*Dataset[float64], error) {
	return dsRolling(r, func(rr *Rolling[T]) (*DataArray[float64], error) { return rr.Min() })
}

// Max : maximum mobile.
func (r *DatasetRolling[T]) Max() (*Dataset[float64], error) {
	return dsRolling(r, func(rr *Rolling[T]) (*DataArray[float64], error) { return rr.Max() })
}
