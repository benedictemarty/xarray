package xarray

import (
	"fmt"
	"sort"
	"strings"
)

// Dataset est une collection de DataArrays (les « variables de données »)
// partageant un système commun de dimensions et de coordonnées.
//
// Toutes les variables portant une même dimension doivent lui donner la même
// taille ; si plusieurs variables définissent une coordonnée pour une même
// dimension, ces coordonnées doivent être identiques.
type Dataset struct {
	vars   map[string]*DataArray // variables de données, par nom
	coords map[string]*Variable  // coordonnées de dimension partagées
	dims   map[string]int        // taille de chaque dimension
}

// NewDataset construit un Dataset à partir d'un ensemble de variables nommées,
// en vérifiant la cohérence des dimensions et des coordonnées.
func NewDataset(vars map[string]*DataArray) (*Dataset, error) {
	ds := &Dataset{
		vars:   make(map[string]*DataArray, len(vars)),
		coords: map[string]*Variable{},
		dims:   map[string]int{},
	}
	// Ordre déterministe pour des messages d'erreur stables.
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
		// Coordonnées : agrégation avec vérification de cohérence.
		for dim, cv := range da.coords {
			if existing, ok := ds.coords[dim]; ok {
				if !sameFloats(existing.data, cv.data) {
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

func sameFloats(a, b []float64) bool {
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
func (ds *Dataset) VarNames() []string {
	names := make([]string, 0, len(ds.vars))
	for n := range ds.vars {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Get renvoie la variable de données nommée name.
func (ds *Dataset) Get(name string) (*DataArray, error) {
	da, ok := ds.vars[name]
	if !ok {
		return nil, fmt.Errorf("xarray: variable %q absente du dataset", name)
	}
	return da.clone(), nil
}

// Dims renvoie la taille de chaque dimension du dataset.
func (ds *Dataset) Dims() map[string]int {
	out := make(map[string]int, len(ds.dims))
	for k, v := range ds.dims {
		out[k] = v
	}
	return out
}

// Coord renvoie les étiquettes de la coordonnée partagée dim.
func (ds *Dataset) Coord(dim string) ([]float64, error) {
	cv, ok := ds.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée %q dans le dataset", dim)
	}
	return cv.Data(), nil
}

// WithVar renvoie une copie du dataset augmentée (ou remplacée) de la variable
// name. La cohérence dimensions/coordonnées est revérifiée.
func (ds *Dataset) WithVar(name string, da *DataArray) (*Dataset, error) {
	next := ds.cloneVars()
	next[name] = da
	return NewDataset(next)
}

// DropVars renvoie une copie du dataset privée des variables indiquées.
func (ds *Dataset) DropVars(names ...string) (*Dataset, error) {
	drop := make(map[string]struct{}, len(names))
	for _, n := range names {
		drop[n] = struct{}{}
	}
	next := make(map[string]*DataArray)
	for n, da := range ds.vars {
		if _, ok := drop[n]; ok {
			continue
		}
		next[n] = da.clone()
	}
	return NewDataset(next)
}

// Merge fusionne deux datasets. En cas de variable homonyme, celle de other
// l'emporte ; la cohérence globale est revérifiée.
func (ds *Dataset) Merge(other *Dataset) (*Dataset, error) {
	next := ds.cloneVars()
	for n, da := range other.vars {
		next[n] = da.clone()
	}
	return NewDataset(next)
}

func (ds *Dataset) cloneVars() map[string]*DataArray {
	next := make(map[string]*DataArray, len(ds.vars))
	for n, da := range ds.vars {
		next[n] = da.clone()
	}
	return next
}

// --- Indexation propagée ----------------------------------------------------

// Isel sélectionne par position le long de dim et propage l'opération à toutes
// les variables portant cette dimension (les autres restent inchangées).
func (ds *Dataset) Isel(dim string, index int) (*Dataset, error) {
	if _, ok := ds.dims[dim]; !ok {
		return nil, fmt.Errorf("xarray: dimension %q absente du dataset", dim)
	}
	next := make(map[string]*DataArray, len(ds.vars))
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

// Sel sélectionne par label le long de dim (via la coordonnée partagée) et
// propage l'opération à toutes les variables portant cette dimension.
func (ds *Dataset) Sel(dim string, label float64) (*Dataset, error) {
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

func (ds *Dataset) reduceAxis(dim string, reducer func(*DataArray) (*DataArray, error)) (*Dataset, error) {
	if _, ok := ds.dims[dim]; !ok {
		return nil, fmt.Errorf("xarray: dimension %q absente du dataset", dim)
	}
	next := make(map[string]*DataArray, len(ds.vars))
	for name, da := range ds.vars {
		if da.HasDim(dim) {
			r, err := reducer(da)
			if err != nil {
				return nil, err
			}
			next[name] = r
		} else {
			next[name] = da.clone()
		}
	}
	return NewDataset(next)
}

// SumAxis réduit la dimension dim par somme sur toutes les variables concernées.
func (ds *Dataset) SumAxis(dim string) (*Dataset, error) {
	return ds.reduceAxis(dim, func(da *DataArray) (*DataArray, error) { return da.SumAxis(dim) })
}

// MeanAxis réduit la dimension dim par moyenne.
func (ds *Dataset) MeanAxis(dim string) (*Dataset, error) {
	return ds.reduceAxis(dim, func(da *DataArray) (*DataArray, error) { return da.MeanAxis(dim) })
}

// MinAxis réduit la dimension dim par minimum.
func (ds *Dataset) MinAxis(dim string) (*Dataset, error) {
	return ds.reduceAxis(dim, func(da *DataArray) (*DataArray, error) { return da.MinAxis(dim) })
}

// MaxAxis réduit la dimension dim par maximum.
func (ds *Dataset) MaxAxis(dim string) (*Dataset, error) {
	return ds.reduceAxis(dim, func(da *DataArray) (*DataArray, error) { return da.MaxAxis(dim) })
}

// String fournit une représentation lisible du dataset.
func (ds *Dataset) String() string {
	var b strings.Builder
	b.WriteString("<xarray.Dataset>\n")

	// Dimensions.
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

	// Coordonnées.
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

	// Variables de données.
	b.WriteString("Variables de données :\n")
	for _, name := range ds.VarNames() {
		da := ds.vars[name]
		dp := make([]string, len(da.variable.dims))
		for i, d := range da.variable.dims {
			dp[i] = d
		}
		fmt.Fprintf(&b, "    %-10s (%s) %v\n", name, strings.Join(dp, ", "), da.variable.data)
	}
	return b.String()
}
