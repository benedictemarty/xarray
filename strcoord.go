package xarray

import "fmt"

// Coordonnées textuelles (non numériques). Une dimension peut porter une
// coordonnée string (ex. noms de stations, catégories) au lieu d'une coordonnée
// numérique. Les données restent de type T (numérique). L'indexation par
// étiquette texte se fait via SelStr.
//
// Portée : les coordonnées textuelles sont préservées par l'indexation
// (Isel/Sel/SelStr/SelRange/SelMany) ; elles ne le sont pas nécessairement par
// les opérations de calcul (réductions, arithmétique), comme les coordonnées
// non-index de xarray.

// WithStrCoord renvoie une copie du tableau avec une coordonnée textuelle sur la
// dimension dim (labels de longueur égale à la taille de dim).
func (da *DataArray[T]) WithStrCoord(dim string, labels []string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	if len(labels) != da.variable.shape[axis] {
		return nil, fmt.Errorf("xarray: coordonnée textuelle %q de longueur %d incompatible avec la dimension de taille %d", dim, len(labels), da.variable.shape[axis])
	}
	out := da.clone()
	if out.strCoords == nil {
		out.strCoords = map[string][]string{}
	}
	out.strCoords[dim] = append([]string(nil), labels...)
	return out, nil
}

// StrCoord renvoie les étiquettes textuelles de la dimension dim.
func (da *DataArray[T]) StrCoord(dim string) ([]string, error) {
	v, ok := da.strCoords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée textuelle pour la dimension %q", dim)
	}
	return append([]string(nil), v...), nil
}

// SelStr sélectionne par étiquette textuelle le long de dim (la dimension est
// réduite, comme Sel).
func (da *DataArray[T]) SelStr(dim string, label string) (*DataArray[T], error) {
	labels, ok := da.strCoords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée textuelle pour la dimension %q", dim)
	}
	pos := -1
	for i, l := range labels {
		if l == label {
			pos = i
			break
		}
	}
	if pos == -1 {
		return nil, fmt.Errorf("xarray: étiquette %q absente de la coordonnée %q", label, dim)
	}
	return da.Isel(dim, pos)
}

// SelStrMany conserve les positions correspondant aux étiquettes textuelles
// fournies (dimension conservée, ordre respecté).
func (da *DataArray[T]) SelStrMany(dim string, labels []string) (*DataArray[T], error) {
	coord, ok := da.strCoords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée textuelle pour la dimension %q", dim)
	}
	pos := make(map[string]int, len(coord))
	for i, l := range coord {
		pos[l] = i
	}
	idx := make([]int, len(labels))
	for i, l := range labels {
		p, found := pos[l]
		if !found {
			return nil, fmt.Errorf("xarray: étiquette %q absente de la coordonnée %q", l, dim)
		}
		idx[i] = p
	}
	return da.takeAlong(dim, idx)
}
