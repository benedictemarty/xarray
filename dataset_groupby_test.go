package xarray

import (
	"reflect"
	"testing"
)

func TestDatasetGroupBy(t *testing.T) {
	// Coordonnée t répétée : groupes {1:[0,1], 2:[2,3,4]}.
	// temperature (t, lieu) 5x2 ; pluie (t) 5.
	temp, _ := NewDataArray([]string{"t", "lieu"}, []int{5, 2},
		[]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		map[string][]float64{"t": {1, 1, 2, 2, 2}, "lieu": {10, 20}}, "temperature")
	pluie, _ := NewDataArray([]string{"t"}, []int{5},
		[]float64{100, 200, 300, 400, 500},
		map[string][]float64{"t": {1, 1, 2, 2, 2}}, "pluie")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"temperature": temp, "pluie": pluie})

	g, err := ds.GroupBy("t")
	if err != nil {
		t.Fatalf("GroupBy : %v", err)
	}
	if g.Groups() != 2 {
		t.Errorf("Groups = %d", g.Groups())
	}

	s, err := g.Sum()
	if err != nil {
		t.Fatalf("Sum : %v", err)
	}

	// pluie : groupe1=100+200=300 ; groupe2=300+400+500=1200
	pl, _ := s.Get("pluie")
	if !reflect.DeepEqual(pl.Data(), []float64{300, 1200}) {
		t.Errorf("pluie groupée = %v, attendu [300 1200]", pl.Data())
	}

	// temperature (t,lieu) : lignes t regroupées, somme sur t par lieu.
	// groupe1 (lignes 0,1) : [1+3, 2+4] = [4 6]
	// groupe2 (lignes 2,3,4) : [5+7+9, 6+8+10] = [21 24]
	tp, _ := s.Get("temperature")
	if !reflect.DeepEqual(tp.Shape(), []int{2, 2}) {
		t.Errorf("temperature shape = %v", tp.Shape())
	}
	if !reflect.DeepEqual(tp.Data(), []float64{4, 6, 21, 24}) {
		t.Errorf("temperature groupée = %v, attendu [4 6 21 24]", tp.Data())
	}

	// La coordonnée t du résultat = groupes uniques.
	if c, _ := s.Coord("t"); !reflect.DeepEqual(c, []float64{1, 2}) {
		t.Errorf("coord t = %v", c)
	}
}

func TestDatasetGroupByMean(t *testing.T) {
	pluie, _ := NewDataArray([]string{"t"}, []int{4}, []float64{10, 20, 30, 50},
		map[string][]float64{"t": {1, 1, 2, 2}}, "pluie")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"pluie": pluie})
	g, _ := ds.GroupBy("t")
	m, err := g.Mean()
	if err != nil {
		t.Fatalf("Mean : %v", err)
	}
	pl, _ := m.Get("pluie")
	// groupe1: (10+20)/2=15 ; groupe2: (30+50)/2=40
	if !reflect.DeepEqual(pl.Data(), []float64{15, 40}) {
		t.Errorf("Mean groupée = %v, attendu [15 40]", pl.Data())
	}
}

func TestDatasetGroupBySansCoord(t *testing.T) {
	a, _ := NewDataArray([]string{"t"}, []int{2}, []float64{1, 2}, nil, "a")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a})
	if _, err := ds.GroupBy("t"); err == nil {
		t.Error("erreur attendue : dimension sans coordonnée")
	}
}
