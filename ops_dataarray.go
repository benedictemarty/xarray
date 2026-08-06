package xarray

import (
	"fmt"
	"math"
)

// --- Réducteurs de base sur une tranche -------------------------------------

func sumSliceG[T Number](s []T) T {
	var t T
	for _, x := range s {
		t += x
	}
	return t
}

func meanSliceG[T Number](s []T) float64 {
	if len(s) == 0 {
		return math.NaN()
	}
	var t float64
	for _, x := range s {
		t += float64(x)
	}
	return t / float64(len(s))
}

func minSliceG[T Number](s []T) T {
	var m T
	if len(s) == 0 {
		return m
	}
	m = s[0]
	for _, x := range s[1:] {
		if x < m {
			m = x
		}
	}
	return m
}

func maxSliceG[T Number](s []T) T {
	var m T
	if len(s) == 0 {
		return m
	}
	m = s[0]
	for _, x := range s[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

// convertNum convertit une valeur numérique d'un type vers un autre, via float64.
// Une perte de précision est possible pour les très grands entiers (documenté).
func convertNum[T, R Number](x T) R { return R(float64(x)) }

// --- Transpose --------------------------------------------------------------

// Transpose réordonne les dimensions du DataArray selon newDims (permutation).
func (da *DataArray[T]) Transpose(newDims ...string) (*DataArray[T], error) {
	nv, err := da.variable.Transpose(newDims...)
	if err != nil {
		return nil, err
	}
	return &DataArray[T]{variable: nv, coords: da.cloneCoords(), name: da.name}, nil
}

func (da *DataArray[T]) cloneCoords() map[string]*Variable[T] {
	coords := make(map[string]*Variable[T], len(da.coords))
	for k, cv := range da.coords {
		nc := cv.cloneVar()
		coords[k] = nc
	}
	return coords
}

// --- Réductions par axe -----------------------------------------------------

// reduceAxisDA réduit la dimension dim et retire sa coordonnée. Le type de
// sortie R peut différer de T (ex. moyenne d'entiers -> float64) ; les
// coordonnées restantes sont converties de T vers R.
func reduceAxisDA[T, R Number](da *DataArray[T], dim string, reducer func([]T) R) (*DataArray[R], error) {
	nv, err := reduceAxisVar(da.variable, dim, reducer)
	if err != nil {
		return nil, err
	}
	coords := make(map[string]*Variable[R], len(da.coords))
	for k, cv := range da.coords {
		if k == dim {
			continue
		}
		labels := make([]R, len(cv.data))
		for i, x := range cv.data {
			labels[i] = convertNum[T, R](x)
		}
		nc, _ := NewVariable(cv.Dims(), cv.Shape(), labels)
		coords[k] = nc
	}
	return &DataArray[R]{variable: nv, coords: coords, name: da.name}, nil
}

// SumAxis réduit la dimension dim par somme.
func (da *DataArray[T]) SumAxis(dim string) (*DataArray[T], error) {
	return reduceAxisDA[T, T](da, dim, sumSliceG[T])
}

// MeanAxis réduit la dimension dim par moyenne (résultat en float64).
func (da *DataArray[T]) MeanAxis(dim string) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, meanSliceG[T])
}

// MinAxis réduit la dimension dim par minimum.
func (da *DataArray[T]) MinAxis(dim string) (*DataArray[T], error) {
	return reduceAxisDA[T, T](da, dim, minSliceG[T])
}

// MaxAxis réduit la dimension dim par maximum.
func (da *DataArray[T]) MaxAxis(dim string) (*DataArray[T], error) {
	return reduceAxisDA[T, T](da, dim, maxSliceG[T])
}

// --- Alignement -------------------------------------------------------------

// align aligne deux DataArrays sur leurs dimensions communes disposant de
// coordonnées des deux côtés : seules les étiquettes communes (jointure interne)
// sont conservées, dans l'ordre du premier opérande.
func align[T Number](a, b *DataArray[T]) (*DataArray[T], *DataArray[T], error) {
	a2, b2 := a, b
	for _, dim := range a.variable.dims {
		ca, okA := a2.coords[dim]
		cb, okB := b2.coords[dim]
		if !okA || !okB {
			continue
		}
		// Chemin rapide : coordonnées déjà identiques -> aucune réindexation
		// (évite deux copies via takeAlong sur ce cas très fréquent).
		if sameSlice(ca.data, cb.data) {
			continue
		}
		posB := make(map[T]int, len(cb.data))
		for i, l := range cb.data {
			posB[l] = i
		}
		var idxA, idxB []int
		var labels []T
		for i, l := range ca.data {
			if j, ok := posB[l]; ok {
				idxA = append(idxA, i)
				idxB = append(idxB, j)
				labels = append(labels, l)
			}
		}
		if len(labels) == 0 {
			return nil, nil, fmt.Errorf("xarray: aucune étiquette commune sur la dimension %q", dim)
		}
		na, err := a2.takeAlong(dim, idxA)
		if err != nil {
			return nil, nil, err
		}
		nb, err := b2.takeAlong(dim, idxB)
		if err != nil {
			return nil, nil, err
		}
		a2, b2 = na, nb
	}
	return a2, b2, nil
}

// takeAlong sélectionne plusieurs positions le long de dim en conservant la
// dimension et en réindexant sa coordonnée.
func (da *DataArray[T]) takeAlong(dim string, indices []int) (*DataArray[T], error) {
	nv, err := da.variable.take(dim, indices)
	if err != nil {
		return nil, err
	}
	coords := make(map[string]*Variable[T], len(da.coords))
	for k, cv := range da.coords {
		if k == dim {
			nc, err := cv.take(dim, indices)
			if err != nil {
				return nil, err
			}
			coords[k] = nc
			continue
		}
		nc := cv.cloneVar()
		coords[k] = nc
	}
	return &DataArray[T]{variable: nv, coords: coords, name: da.name}, nil
}

// --- Arithmétique -----------------------------------------------------------

// binary applique une opération binaire élément par élément entre deux
// DataArrays, avec alignement automatique puis broadcasting par nom.
func (da *DataArray[T]) binary(other *DataArray[T], fn func(x, y T) T) (*DataArray[T], error) {
	a, b, err := align(da, other)
	if err != nil {
		return nil, err
	}
	nv, err := binaryOp(a.variable, b.variable, fn)
	if err != nil {
		return nil, err
	}
	coords := make(map[string]*Variable[T], len(nv.dims))
	for _, dim := range nv.dims {
		if cv, ok := a.coords[dim]; ok {
			nc := cv.cloneVar()
			coords[dim] = nc
		} else if cv, ok := b.coords[dim]; ok {
			nc := cv.cloneVar()
			coords[dim] = nc
		}
	}
	return &DataArray[T]{variable: nv, coords: coords, name: da.name}, nil
}

// Add renvoie da + other (avec alignement et broadcasting).
//
// Optimisation : pour le cas fréquent « float64, mêmes dimensions après
// alignement », on utilise une addition directe (addFloat64) sans closure. Le
// chemin générique passe par une closure func(T,T) T qui n'est pas inlinée (un
// appel par élément) ; l'éviter accélère l'opération d'un ordre de grandeur sur
// les grands tableaux (voir docs/BENCHMARKS.md).
func (da *DataArray[T]) Add(other *DataArray[T]) (*DataArray[T], error) {
	if r, ok, err := da.addFast(other); err != nil {
		return nil, err
	} else if ok {
		return r, nil
	}
	return da.binary(other, func(x, y T) T { return x + y })
}

// addFast : chemin rapide float64 à formes identiques (sans closure). Renvoie
// ok=false pour signaler le repli sur le chemin générique (autre type, ou
// broadcasting requis).
func (da *DataArray[T]) addFast(other *DataArray[T]) (*DataArray[T], bool, error) {
	if _, ok := any(da.variable.data).([]float64); !ok {
		return nil, false, nil
	}
	a, b, err := align(da, other)
	if err != nil {
		return nil, false, err
	}
	if !sameDimsShape(a.variable, b.variable) {
		return nil, false, nil
	}
	ad := any(a.variable.data).([]float64)
	bd := any(b.variable.data).([]float64)
	dst := make([]float64, len(ad))
	addFloat64(dst, ad, bd)

	nv := &Variable[T]{
		dims:  a.variable.Dims(),
		shape: a.variable.Shape(),
		data:  any(dst).([]T),
		attrs: map[string]string{},
	}
	coords := make(map[string]*Variable[T], len(nv.dims))
	for _, dim := range nv.dims {
		if cv, ok := a.coords[dim]; ok {
			coords[dim] = cv.cloneVar()
		} else if cv, ok := b.coords[dim]; ok {
			coords[dim] = cv.cloneVar()
		}
	}
	return &DataArray[T]{variable: nv, coords: coords, name: da.name}, true, nil
}

// Sub renvoie da - other.
func (da *DataArray[T]) Sub(other *DataArray[T]) (*DataArray[T], error) {
	return da.binary(other, func(x, y T) T { return x - y })
}

// Mul renvoie da * other.
func (da *DataArray[T]) Mul(other *DataArray[T]) (*DataArray[T], error) {
	return da.binary(other, func(x, y T) T { return x * y })
}

// Div renvoie da / other.
func (da *DataArray[T]) Div(other *DataArray[T]) (*DataArray[T], error) {
	return da.binary(other, func(x, y T) T { return x / y })
}

// --- Opérations scalaires ---------------------------------------------------

// AddScalar ajoute s à chaque élément.
func (da *DataArray[T]) AddScalar(s T) *DataArray[T] {
	return da.mapScalar(func(x T) T { return x + s })
}

// MulScalar multiplie chaque élément par s.
func (da *DataArray[T]) MulScalar(s T) *DataArray[T] {
	return da.mapScalar(func(x T) T { return x * s })
}

// mapScalar applique fn à chaque élément en conservant dimensions et coordonnées.
func (da *DataArray[T]) mapScalar(fn func(T) T) *DataArray[T] {
	nv := da.variable.mapScalar(fn)
	return &DataArray[T]{variable: nv, coords: da.cloneCoords(), name: da.name}
}
