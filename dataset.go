package xarray

import (
	"fmt"
	"sort"
	"strings"
)

// Dataset est une collection de DataArrays (les « variables de données »)
// partageant un système commun de dimensions et de coordonnées.
type Dataset[T Number] struct {
	vars   map[string]*DataArray[T]
	coords map[string]*Variable[T]
	dims   map[string]int
}

// NewDataset construit un Dataset à partir d'un ensemble de variables nommées,
// en vérifiant la cohérence des dimensions et des coordonnées.
func NewDataset[T Number](vars map[string]*DataArray[T]) (*Dataset[T], error) {
	ds := &Dataset[T]{
		vars:   make(map[string]*DataArray[T], len(vars)),
		coords: map[string]*Variable[T]{},
		dims:   map[string]int{},
	}
	names := make([]string, 0, len(vars))
	for n := range vars {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		da := vars[name]
		if da == nil {
			return nil, fmt.Errorf("xarray: variable %q nulle", name)
		}
		dims := da.variable.dims
		shape := da.variable.shape
		for i, d := range dims {
			if s, ok := ds.dims[d]; ok {
				if s != shape[i] {
					return nil, fmt.Errorf("xarray: dimension %q de tailles incohérentes (%d vs %d)", d, s, shape[i])
				}
			} else {
				ds.dims[d] = shape[i]
			}
		}
		for dim, cv := range da.coords {
			if existing, ok := ds.coords[dim]; ok {
				if !sameSlice(existing.data, cv.data) {
					return nil, fmt.Errorf("xarray: coordonnées %q incohérentes entre variables", dim)
				}
			} else {
				nc, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
				ds.coords[dim] = nc
			}
		}
		ds.vars[name] = da.clone()
	}
	return ds, nil
}

func sameSlice[T Number](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// VarNames renvoie les noms des variables de données, triés.
func (ds *Dataset[T]) VarNames() []string {
	names := make([]string, 0, len(ds.vars))
	for n := range ds.vars {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Get renvoie la variable de données nommée name.
func (ds *Dataset[T]) Get(name string) (*DataArray[T], error) {
	da, ok := ds.vars[name]
	if !ok {
		return nil, fmt.Errorf("xarray: variable %q absente du dataset", name)
	}
	return da.clone(), nil
}

// Dims renvoie la taille de chaque dimension du dataset.
func (ds *Dataset[T]) Dims() map[string]int {
	out := make(map[string]int, len(ds.dims))
	for k, v := range ds.dims {
		out[k] = v
	}
	return out
}

// Coord renvoie les étiquettes de la coordonnée partagée dim.
func (ds *Dataset[T]) Coord(dim string) ([]T, error) {
	cv, ok := ds.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée %q dans le dataset", dim)
	}
	return cv.Data(), nil
}

// WithVar renvoie une copie du dataset augmentée (ou remplacée) de la variable name.
func (ds *Dataset[T]) WithVar(name string, da *DataArray[T]) (*Dataset[T], error) {
	next := ds.cloneVars()
	next[name] = da
	return NewDataset(next)
}

// DropVars renvoie une copie du dataset privée des variables indiquées.
func (ds *Dataset[T]) DropVars(names ...string) (*Dataset[T], error) {
	drop := make(map[string]struct{}, len(names))
	for _, n := range names {
		drop[n] = struct{}{}
	}
	next := make(map[string]*DataArray[T])
	for n, da := range ds.vars {
		if _, ok := drop[n]; ok {
			continue
		}
		next[n] = da.clone()
	}
	return NewDataset(next)
}

// Merge fusionne deux datasets. En cas de variable homonyme, celle de other l'emporte.
func (ds *Dataset[T]) Merge(other *Dataset[T]) (*Dataset[T], error) {
	next := ds.cloneVars()
	for n, da := range other.vars {
		next[n] = da.clone()
	}
	return NewDataset(next)
}

func (ds *Dataset[T]) cloneVars() map[string]*DataArray[T] {
	next := make(map[string]*DataArray[T], len(ds.vars))
	for n, da := range ds.vars {
		next[n] = da.clone()
	}
	return next
}

// --- Indexation propagée ----------------------------------------------------

// Isel sélectionne par position le long de dim et propage l'opération à toutes
// les variables portant cette dimension (les autres restent inchangées).
func (ds *Dataset[T]) Isel(dim string, index int) (*Dataset[T], error) {
	if _, ok := ds.dims[dim]; !ok {
		return nil, fmt.Errorf("xarray: dimension %q absente du dataset", dim)
	}
	next := make(map[string]*DataArray[T], len(ds.vars))
	for name, da := range ds.vars {
		if da.HasDim(dim) {
			sub, err := da.Isel(dim, index)
			if err != nil {
				return nil, err
			}
			next[name] = sub
		} else {
			next[name] = da.clone()
		}
	}
	return NewDataset(next)
}

// Sel sélectionne par label le long de dim (via la coordonnée partagée).
func (ds *Dataset[T]) Sel(dim string, label T) (*Dataset[T], error) {
	cv, ok := ds.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: indexation par label impossible : aucune coordonnée %q", dim)
	}
	pos := -1
	for i, l := range cv.data {
		if l == label {
			pos = i
			break
		}
	}
	if pos == -1 {
		return nil, fmt.Errorf("xarray: étiquette %v absente de la coordonnée %q", label, dim)
	}
	return ds.Isel(dim, pos)
}

// --- Réductions propagées ---------------------------------------------------

// convertDataArray convertit un DataArray d'un type numérique vers un autre.
func convertDataArray[T, R Number](da *DataArray[T]) *DataArray[R] {
	data := make([]R, len(da.variable.data))
	for i, x := range da.variable.data {
		data[i] = convertNum[T, R](x)
	}
	nv, _ := NewVariable(da.variable.Dims(), da.variable.Shape(), data)
	coords := make(map[string]*Variable[R], len(da.coords))
	for k, cv := range da.coords {
		lbl := make([]R, len(cv.data))
		for i, x := range cv.data {
			lbl[i] = convertNum[T, R](x)
		}
		nc, _ := NewVariable(cv.Dims(), cv.Shape(), lbl)
		coords[k] = nc
	}
	return &DataArray[R]{variable: nv, coords: coords, name: da.name}
}

// reduceDatasetAxis applique reducer aux variables portant dim et convertit les
// autres vers le type de sortie R, puis reconstruit un Dataset[R].
func reduceDatasetAxis[T, R Number](ds *Dataset[T], dim string, reducer func(*DataArray[T]) (*DataArray[R], error)) (*Dataset[R], error) {
	if _, ok := ds.dims[dim]; !ok {
		return nil, fmt.Errorf("xarray: dimension %q absente du dataset", dim)
	}
	next := make(map[string]*DataArray[R], len(ds.vars))
	for name, da := range ds.vars {
		if da.HasDim(dim) {
			r, err := reducer(da)
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

// SumAxis réduit la dimension dim par somme sur toutes les variables concernées.
func (ds *Dataset[T]) SumAxis(dim string) (*Dataset[T], error) {
	return reduceDatasetAxis[T, T](ds, dim, func(da *DataArray[T]) (*DataArray[T], error) { return da.SumAxis(dim) })
}

// MeanAxis réduit la dimension dim par moyenne (résultat en float64).
func (ds *Dataset[T]) MeanAxis(dim string) (*Dataset[float64], error) {
	return reduceDatasetAxis[T, float64](ds, dim, func(da *DataArray[T]) (*DataArray[float64], error) { return da.MeanAxis(dim) })
}

// MinAxis réduit la dimension dim par minimum.
func (ds *Dataset[T]) MinAxis(dim string) (*Dataset[T], error) {
	return reduceDatasetAxis[T, T](ds, dim, func(da *DataArray[T]) (*DataArray[T], error) { return da.MinAxis(dim) })
}

// MaxAxis réduit la dimension dim par maximum.
func (ds *Dataset[T]) MaxAxis(dim string) (*Dataset[T], error) {
	return reduceDatasetAxis[T, T](ds, dim, func(da *DataArray[T]) (*DataArray[T], error) { return da.MaxAxis(dim) })
}

// String fournit une représentation lisible du dataset.
func (ds *Dataset[T]) String() string {
	var b strings.Builder
	b.WriteString("<xarray.Dataset>\n")

	dimNames := make([]string, 0, len(ds.dims))
	for d := range ds.dims {
		dimNames = append(dimNames, d)
	}
	sort.Strings(dimNames)
	parts := make([]string, len(dimNames))
	for i, d := range dimNames {
		parts[i] = fmt.Sprintf("%s: %d", d, ds.dims[d])
	}
	fmt.Fprintf(&b, "Dimensions : (%s)\n", strings.Join(parts, ", "))

	if len(ds.coords) > 0 {
		b.WriteString("Coordonnées :\n")
		coordNames := make([]string, 0, len(ds.coords))
		for d := range ds.coords {
			coordNames = append(coordNames, d)
		}
		sort.Strings(coordNames)
		for _, d := range coordNames {
			fmt.Fprintf(&b, "  * %-10s %v\n", d, ds.coords[d].data)
		}
	}

	b.WriteString("Variables de données :\n")
	for _, name := range ds.VarNames() {
		da := ds.vars[name]
		fmt.Fprintf(&b, "    %-10s (%s) %v\n", name, strings.Join(da.variable.dims, ", "), da.variable.data)
	}
	return b.String()
}
