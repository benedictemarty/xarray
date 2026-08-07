package xarray

import (
	"fmt"
	"sort"
)

// Composantes temporelles et regroupement par composante (climatologie), à la
// manière de `ds.groupby("time.month")` de xarray.

// TimeComponent désigne une composante calendaire extraite d'un instant.
type TimeComponent int

const (
	CompYear TimeComponent = iota
	CompMonth
	CompDay
	CompHour
	CompMinute
	CompWeekday   // 0 = dimanche … 6 = samedi
	CompDayOfYear // 1..366
	CompSeason    // 0 = DJF (hiver), 1 = MAM, 2 = JJA, 3 = SON
)

// SeasonName renvoie le nom court d'une saison météorologique (0..3).
func SeasonName(s int) string {
	switch s {
	case 0:
		return "DJF"
	case 1:
		return "MAM"
	case 2:
		return "JJA"
	case 3:
		return "SON"
	default:
		return "?"
	}
}

func componentOf(sec float64, c TimeComponent) int {
	t := TimeFromEpoch(sec)
	switch c {
	case CompYear:
		return t.Year()
	case CompMonth:
		return int(t.Month())
	case CompDay:
		return t.Day()
	case CompHour:
		return t.Hour()
	case CompMinute:
		return t.Minute()
	case CompWeekday:
		return int(t.Weekday())
	case CompDayOfYear:
		return t.YearDay()
	default: // CompSeason : DJF=0, MAM=1, JJA=2, SON=3 (saison météorologique)
		return (int(t.Month()) % 12) / 3
	}
}

// ExtractTime renvoie la composante c de chaque instant (coordonnée en secondes
// epoch). Pratique pour construire une coordonnée dérivée (ex. le mois).
func ExtractTime(epochCoord []float64, c TimeComponent) []float64 {
	out := make([]float64, len(epochCoord))
	for i, s := range epochCoord {
		out[i] = float64(componentOf(s, c))
	}
	return out
}

// GroupByTime regroupe le long de dim par composante temporelle (la coordonnée
// de dim étant en secondes epoch). Par exemple, grouper par CompMonth réunit tous
// les mois de janvier ensemble (climatologie mensuelle), quelle que soit l'année.
// Renvoie un Resample (agrégations Sum/Mean/Min/Max) dont les étiquettes sont les
// valeurs de composante (1..12 pour les mois, etc.).
func GroupByTime[T Number](da *DataArray[T], dim string, c TimeComponent) (*Resample[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: groupby temporel impossible : aucune coordonnée %q", dim)
	}
	byComp := map[int][]int{}
	var order []int
	for i, l := range cv.data {
		k := componentOf(float64(l), c)
		if _, seen := byComp[k]; !seen {
			order = append(order, k)
		}
		byComp[k] = append(byComp[k], i)
	}
	sort.Ints(order)

	labels := make([]T, len(order))
	groups := make([][]int, len(order))
	for idx, k := range order {
		labels[idx] = T(k)
		groups[idx] = byComp[k]
	}
	return &Resample[T]{da: da, dim: dim, labels: labels, groups: groups}, nil
}
