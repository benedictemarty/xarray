package xarray

import (
	"reflect"
	"testing"
	"time"
)

func TestExtractTime(t *testing.T) {
	times := []time.Time{
		date(2020, 1, 15),
		date(2020, 6, 10),
		date(2021, 12, 25),
	}
	coord := EpochCoord(times)
	months := ExtractTime(coord, CompMonth)
	if !reflect.DeepEqual(months, []float64{1, 6, 12}) {
		t.Errorf("mois = %v, attendu [1 6 12]", months)
	}
	years := ExtractTime(coord, CompYear)
	if !reflect.DeepEqual(years, []float64{2020, 2020, 2021}) {
		t.Errorf("années = %v", years)
	}
}

func TestGroupByTimeMonth(t *testing.T) {
	// Deux années de données mensuelles partielles : jan/fév 2020 et 2021.
	times := []time.Time{
		date(2020, 1, 10), // jan
		date(2020, 2, 10), // fév
		date(2021, 1, 20), // jan
		date(2021, 2, 20), // fév
	}
	da, _ := NewDataArray([]string{"temps"}, []int{4}, []float64{10, 20, 30, 40},
		map[string][]float64{"temps": EpochCoord(times)}, "v")

	// Climatologie mensuelle : moyenne de tous les janviers, de tous les févriers.
	g, err := GroupByTime(da, "temps", CompMonth)
	if err != nil {
		t.Fatalf("GroupByTime : %v", err)
	}
	if g.Groups() != 2 {
		t.Errorf("Groups = %d, attendu 2 (janvier, février)", g.Groups())
	}
	m, _ := g.Mean()
	// janvier : (10+30)/2 = 20 ; février : (20+40)/2 = 30
	if !reflect.DeepEqual(m.Data(), []float64{20, 30}) {
		t.Errorf("climatologie mensuelle = %v, attendu [20 30]", m.Data())
	}
	// Étiquettes = numéros de mois.
	if c, _ := m.Coord("temps"); !reflect.DeepEqual(c, []float64{1, 2}) {
		t.Errorf("étiquettes = %v, attendu [1 2]", c)
	}
}

func TestGroupByTimeSum(t *testing.T) {
	times := []time.Time{
		date(2020, 3, 1), date(2020, 3, 15), date(2020, 7, 1),
	}
	da, _ := NewDataArray([]string{"t"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"t": EpochCoord(times)}, "v")
	g, _ := GroupByTime(da, "t", CompMonth)
	s, _ := g.Sum()
	// mars : 1+2=3 ; juillet : 3
	if !reflect.DeepEqual(s.Data(), []float64{3, 3}) {
		t.Errorf("somme mensuelle = %v, attendu [3 3]", s.Data())
	}
}

func TestGroupByTimeSansCoord(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{2}, []float64{1, 2}, nil, "v")
	if _, err := GroupByTime(da, "t", CompMonth); err == nil {
		t.Error("erreur attendue : pas de coordonnée")
	}
}
