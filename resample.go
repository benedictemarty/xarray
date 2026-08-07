package xarray

import "fmt"

// Resample regroupe les valeurs le long d'une dimension par intervalles réguliers
// de sa coordonnée (de largeur freq), à la manière de xarray.DataArray.resample.
//
// Comme notre modèle ne gère pas encore le temps, le rééchantillonnage se fait
// sur une coordonnée numérique : le bin d'une étiquette l est
// floor((l - origine) / freq), l'origine étant la plus petite étiquette. La
// dimension est réduite aux bins (non vides), dont la coordonnée devient la borne
// gauche.
type Resample[T Number] struct {
	da     *DataArray[T]
	dim    string
	labels []T
	groups [][]int
}

// Resample construit un rééchantillonnage de pas freq le long de dim.
func (da *DataArray[T]) Resample(dim string, freq T) (*Resample[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: resample impossible : aucune coordonnée pour la dimension %q", dim)
	}
	if float64(freq) <= 0 {
		return nil, fmt.Errorf("xarray: pas de rééchantillonnage invalide %v", freq)
	}
	binLabels, groups := binGroups(cv.data, freq)
	return &Resample[T]{da: da, dim: dim, labels: binLabels, groups: groups}, nil
}

// Groups renvoie le nombre de bins non vides.
func (r *Resample[T]) Groups() int { return len(r.labels) }

// Sum agrège chaque bin par somme.
func (r *Resample[T]) Sum() (*DataArray[T], error) {
	return groupReduceOn(r.da, r.dim, r.groups, r.labels,
		func(d *DataArray[T]) (*DataArray[T], error) { return d.SumAxis(r.dim) })
}

// Mean agrège chaque bin par moyenne (float64).
func (r *Resample[T]) Mean() (*DataArray[float64], error) {
	return groupReduceOn(r.da, r.dim, r.groups, r.labels,
		func(d *DataArray[T]) (*DataArray[float64], error) { return d.MeanAxis(r.dim) })
}

// Min agrège chaque bin par minimum.
func (r *Resample[T]) Min() (*DataArray[T], error) {
	return groupReduceOn(r.da, r.dim, r.groups, r.labels,
		func(d *DataArray[T]) (*DataArray[T], error) { return d.MinAxis(r.dim) })
}

// Max agrège chaque bin par maximum.
func (r *Resample[T]) Max() (*DataArray[T], error) {
	return groupReduceOn(r.da, r.dim, r.groups, r.labels,
		func(d *DataArray[T]) (*DataArray[T], error) { return d.MaxAxis(r.dim) })
}
