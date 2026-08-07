package xarray

import (
	"reflect"
	"testing"
	"time"
)

func TestGroupBySeason(t *testing.T) {
	// Un mois représentatif de chaque saison, + un doublon hivernal.
	times := []time.Time{
		date(2020, 1, 15),  // DJF (0)
		date(2020, 12, 20), // DJF (0)
		date(2020, 4, 10),  // MAM (1)
		date(2020, 7, 5),   // JJA (2)
		date(2020, 10, 1),  // SON (3)
	}
	da, _ := NewDataArray([]string{"t"}, []int{5}, []float64{2, 4, 10, 20, 30},
		map[string][]float64{"t": EpochCoord(times)}, "v")

	g, err := GroupByTime(da, "t", CompSeason)
	if err != nil {
		t.Fatalf("GroupByTime saison : %v", err)
	}
	if g.Groups() != 4 {
		t.Errorf("Groups = %d, attendu 4", g.Groups())
	}
	m, _ := g.Mean()
	// DJF : (2+4)/2=3 ; MAM : 10 ; JJA : 20 ; SON : 30
	if !reflect.DeepEqual(m.Data(), []float64{3, 10, 20, 30}) {
		t.Errorf("moyennes saisonnières = %v, attendu [3 10 20 30]", m.Data())
	}
	// Étiquettes = indices de saison 0..3.
	if c, _ := m.Coord("t"); !reflect.DeepEqual(c, []float64{0, 1, 2, 3}) {
		t.Errorf("étiquettes = %v", c)
	}
}

func TestSeasonName(t *testing.T) {
	got := []string{SeasonName(0), SeasonName(1), SeasonName(2), SeasonName(3)}
	if !reflect.DeepEqual(got, []string{"DJF", "MAM", "JJA", "SON"}) {
		t.Errorf("noms de saison = %v", got)
	}
}

func TestExtractSeason(t *testing.T) {
	times := []time.Time{date(2020, 12, 1), date(2020, 3, 1), date(2020, 8, 1)}
	s := ExtractTime(EpochCoord(times), CompSeason)
	// déc -> DJF(0), mars -> MAM(1), août -> JJA(2)
	if !reflect.DeepEqual(s, []float64{0, 1, 2}) {
		t.Errorf("saisons = %v, attendu [0 1 2]", s)
	}
}
