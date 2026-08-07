package xarray

import (
	"reflect"
	"testing"
)

func TestGroupByBins(t *testing.T) {
	// coord [1 5 12 18 25], edges [0 10 20 30] -> bins {1,5},{12,18},{25}
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{10, 20, 30, 40, 50},
		map[string][]float64{"x": {1, 5, 12, 18, 25}}, "v")
	g, err := da.GroupByBins("x", []float64{0, 10, 20, 30})
	if err != nil {
		t.Fatalf("GroupByBins : %v", err)
	}
	if g.Groups() != 3 {
		t.Errorf("Groups = %d, attendu 3", g.Groups())
	}
	m, _ := g.Mean()
	// bin0 (1,5) -> (10+20)/2=15 ; bin1 (12,18) -> (30+40)/2=35 ; bin2 (25) -> 50
	if !reflect.DeepEqual(m.Data(), []float64{15, 35, 50}) {
		t.Errorf("GroupByBins mean = %v, attendu [15 35 50]", m.Data())
	}
	// étiquettes = bornes gauches
	if c, _ := m.Coord("x"); !reflect.DeepEqual(c, []float64{0, 10, 20}) {
		t.Errorf("coord = %v, attendu [0 10 20]", c)
	}
}

func TestGroupByBinsHorsBornes(t *testing.T) {
	// Valeurs hors [0,20] ignorées.
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{1, 2, 3, 4},
		map[string][]float64{"x": {-5, 5, 15, 100}}, "v")
	g, _ := da.GroupByBins("x", []float64{0, 10, 20})
	s, _ := g.Sum()
	// -5 et 100 ignorés ; bin0 (5)->2 ; bin1 (15)->3
	if !reflect.DeepEqual(s.Data(), []float64{2, 3}) {
		t.Errorf("Sum = %v, attendu [2 3]", s.Data())
	}
}

func TestGroupByBinsDataset(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{4}, []float64{10, 20, 30, 40},
		map[string][]float64{"x": {1, 2, 11, 12}}, "a")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a})
	g, _ := ds.GroupByBins("x", []float64{0, 10, 20})
	m, _ := g.Mean()
	av, _ := m.Get("a")
	// bin0 (1,2)->15 ; bin1 (11,12)->35
	if !reflect.DeepEqual(av.Data(), []float64{15, 35}) {
		t.Errorf("GroupByBins dataset = %v, attendu [15 35]", av.Data())
	}
}

func TestGroupByBinsBornesInsuffisantes(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2},
		map[string][]float64{"x": {0, 1}}, "v")
	if _, err := da.GroupByBins("x", []float64{5}); err == nil {
		t.Error("erreur attendue : bornes insuffisantes")
	}
}
