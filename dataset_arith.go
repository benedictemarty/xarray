package xarray

import "fmt"

// Arithmétique entre Datasets : opère variable par variable sur les variables de
// même nom (chacune via l'opération DataArray correspondante, avec alignement et
// broadcasting). Les scalaires s'appliquent à toutes les variables.

func (ds *Dataset[T]) binaryDS(other *Dataset[T], op func(a, b *DataArray[T]) (*DataArray[T], error)) (*Dataset[T], error) {
	next := make(map[string]*DataArray[T], len(ds.vars))
	for name, da := range ds.vars {
		od, ok := other.vars[name]
		if !ok {
			return nil, fmt.Errorf("xarray: variable %q absente de l'autre dataset", name)
		}
		r, err := op(da, od)
		if err != nil {
			return nil, fmt.Errorf("xarray: variable %q : %w", name, err)
		}
		next[name] = r
	}
	return NewDataset(next)
}

// Add renvoie ds + other (variable par variable).
func (ds *Dataset[T]) Add(other *Dataset[T]) (*Dataset[T], error) {
	return ds.binaryDS(other, func(a, b *DataArray[T]) (*DataArray[T], error) { return a.Add(b) })
}

// Sub renvoie ds - other.
func (ds *Dataset[T]) Sub(other *Dataset[T]) (*Dataset[T], error) {
	return ds.binaryDS(other, func(a, b *DataArray[T]) (*DataArray[T], error) { return a.Sub(b) })
}

// Mul renvoie ds * other.
func (ds *Dataset[T]) Mul(other *Dataset[T]) (*Dataset[T], error) {
	return ds.binaryDS(other, func(a, b *DataArray[T]) (*DataArray[T], error) { return a.Mul(b) })
}

// Div renvoie ds / other.
func (ds *Dataset[T]) Div(other *Dataset[T]) (*Dataset[T], error) {
	return ds.binaryDS(other, func(a, b *DataArray[T]) (*DataArray[T], error) { return a.Div(b) })
}

// AddScalar ajoute s à toutes les variables.
func (ds *Dataset[T]) AddScalar(s T) (*Dataset[T], error) {
	next := make(map[string]*DataArray[T], len(ds.vars))
	for name, da := range ds.vars {
		next[name] = da.AddScalar(s)
	}
	return NewDataset(next)
}

// MulScalar multiplie toutes les variables par s.
func (ds *Dataset[T]) MulScalar(s T) (*Dataset[T], error) {
	next := make(map[string]*DataArray[T], len(ds.vars))
	for name, da := range ds.vars {
		next[name] = da.MulScalar(s)
	}
	return NewDataset(next)
}
