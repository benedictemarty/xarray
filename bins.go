package xarray

import (
	"math"
	"sort"
)

// Helpers de regroupement : à partir d'une coordonnée (étiquettes), calculent la
// liste des étiquettes de groupe et les indices de chaque groupe. Mutualisés
// entre les niveaux DataArray et Dataset (Resample, ResampleCalendar, GroupByTime).

// binGroups regroupe par intervalles réguliers de largeur freq
// (floor((l-origine)/freq)). Étiquettes = bornes gauches des bins.
func binGroups[T Number](labels []T, freq T) ([]T, [][]int) {
	origin := labels[0]
	for _, l := range labels {
		if l < origin {
			origin = l
		}
	}
	m := map[int64][]int{}
	var order []int64
	for i, l := range labels {
		b := int64(math.Floor(float64(l-origin) / float64(freq)))
		if _, seen := m[b]; !seen {
			order = append(order, b)
		}
		m[b] = append(m[b], i)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	blabels := make([]T, len(order))
	groups := make([][]int, len(order))
	for k, b := range order {
		blabels[k] = origin + T(b)*freq
		groups[k] = m[b]
	}
	return blabels, groups
}

// calendarGroups regroupe par période calendaire (la coordonnée étant en secondes
// epoch). Étiquettes = débuts de période (epoch).
func calendarGroups[T Number](labels []T, p Period) ([]T, [][]int) {
	m := map[int64][]int{}
	var order []int64
	for i, l := range labels {
		start := periodStart(TimeFromEpoch(float64(l)), p)
		k := start.Unix()
		if _, seen := m[k]; !seen {
			order = append(order, k)
		}
		m[k] = append(m[k], i)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	blabels := make([]T, len(order))
	groups := make([][]int, len(order))
	for idx, k := range order {
		blabels[idx] = T(k)
		groups[idx] = m[k]
	}
	return blabels, groups
}

// componentGroups regroupe par composante temporelle (mois, saison, etc.).
// Étiquettes = valeurs de composante.
func componentGroups[T Number](labels []T, c TimeComponent) ([]T, [][]int) {
	m := map[int][]int{}
	var order []int
	for i, l := range labels {
		k := componentOf(float64(l), c)
		if _, seen := m[k]; !seen {
			order = append(order, k)
		}
		m[k] = append(m[k], i)
	}
	sort.Ints(order)
	blabels := make([]T, len(order))
	groups := make([][]int, len(order))
	for idx, k := range order {
		blabels[idx] = T(k)
		groups[idx] = m[k]
	}
	return blabels, groups
}
