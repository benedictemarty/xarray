package xarray

// Propagation d'opérations supplémentaires au niveau Dataset (cohérence avec les
// méthodes de DataArray). Les réductions par axe s'appliquent aux variables
// portant la dimension ; les autres sont conservées (converties si besoin).

// VarAxis réduit dim par variance (float64) sur toutes les variables concernées.
func (ds *Dataset[T]) VarAxis(dim string) (*Dataset[float64], error) {
	return reduceDatasetAxis[T, float64](ds, dim,
		func(da *DataArray[T]) (*DataArray[float64], error) { return da.VarAxis(dim) })
}

// StdAxis réduit dim par écart-type (float64).
func (ds *Dataset[T]) StdAxis(dim string) (*Dataset[float64], error) {
	return reduceDatasetAxis[T, float64](ds, dim,
		func(da *DataArray[T]) (*DataArray[float64], error) { return da.StdAxis(dim) })
}

// MedianAxis réduit dim par médiane (float64).
func (ds *Dataset[T]) MedianAxis(dim string) (*Dataset[float64], error) {
	return reduceDatasetAxis[T, float64](ds, dim,
		func(da *DataArray[T]) (*DataArray[float64], error) { return da.MedianAxis(dim) })
}

// FillNA renvoie une copie du dataset où les NaN de chaque variable sont
// remplacés par value.
func (ds *Dataset[T]) FillNA(value T) (*Dataset[T], error) {
	next := make(map[string]*DataArray[T], len(ds.vars))
	for name, da := range ds.vars {
		next[name] = da.FillNA(value)
	}
	return NewDataset(next)
}

// Cumsum applique la somme cumulée le long de dim à chaque variable la portant
// (les autres sont conservées).
func (ds *Dataset[T]) Cumsum(dim string) (*Dataset[T], error) {
	next := make(map[string]*DataArray[T], len(ds.vars))
	for name, da := range ds.vars {
		if da.HasDim(dim) {
			c, err := da.Cumsum(dim)
			if err != nil {
				return nil, err
			}
			next[name] = c
		} else {
			next[name] = da.clone()
		}
	}
	return NewDataset(next)
}
