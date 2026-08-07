package xarray

import (
	"fmt"
	"sort"
	"time"
)

// Gestion (basique) du temps. Une coordonnée temporelle est représentée par des
// **secondes depuis l'epoch Unix** (float64, en UTC). Cela s'intègre au modèle
// générique sans type dédié.
//
// Limite : float64 a ~52 bits de mantisse → précision de l'ordre de la
// microseconde sur les dates modernes, suffisante pour la plupart des usages
// mais pas pour la nanoseconde. Pas de gestion des calendriers non standard.

// EpochSeconds convertit un instant en secondes depuis l'epoch Unix (UTC).
func EpochSeconds(t time.Time) float64 {
	return float64(t.UTC().UnixNano()) / 1e9
}

// TimeFromEpoch reconstruit un instant (UTC) à partir de secondes epoch.
func TimeFromEpoch(sec float64) time.Time {
	return time.Unix(0, int64(sec*1e9)).UTC()
}

// EpochCoord convertit une suite d'instants en coordonnée (secondes epoch),
// prête à être passée comme coordonnée d'un DataArray[float64].
func EpochCoord(times []time.Time) []float64 {
	out := make([]float64, len(times))
	for i, t := range times {
		out[i] = EpochSeconds(t)
	}
	return out
}

// Period désigne une période calendaire de rééchantillonnage.
type Period int

const (
	PeriodHour Period = iota
	PeriodDay
	PeriodMonth
	PeriodYear
)

// periodStart tronque un instant au début de sa période calendaire (UTC).
func periodStart(t time.Time, p Period) time.Time {
	t = t.UTC()
	switch p {
	case PeriodHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
	case PeriodDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	case PeriodMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	default: // PeriodYear
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	}
}

// ResampleCalendar regroupe le long de dim par période calendaire (heure, jour,
// mois, année), la coordonnée de dim étant interprétée comme des secondes epoch.
// Renvoie un Resample dont les étiquettes sont les débuts de période (epoch).
func ResampleCalendar[T Number](da *DataArray[T], dim string, p Period) (*Resample[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: resample temporel impossible : aucune coordonnée %q", dim)
	}
	groupsByKey := map[int64][]int{}
	var order []int64
	for i, l := range cv.data {
		start := periodStart(TimeFromEpoch(float64(l)), p)
		k := start.Unix()
		if _, seen := groupsByKey[k]; !seen {
			order = append(order, k)
		}
		groupsByKey[k] = append(groupsByKey[k], i)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })

	labels := make([]T, len(order))
	groups := make([][]int, len(order))
	for idx, k := range order {
		labels[idx] = T(k) // début de période, en secondes epoch
		groups[idx] = groupsByKey[k]
	}
	return &Resample[T]{da: da, dim: dim, labels: labels, groups: groups}, nil
}
