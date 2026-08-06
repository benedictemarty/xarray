package xarray

import (
	"reflect"
	"testing"
)

func exempleDataset(t *testing.T) *Dataset[float64] {
	t.Helper()
	// Deux variables partageant la dimension "temps" (coord 2020,2021).
	temp, _ := NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")
	pluie, _ := NewDataArray([]string{"temps"}, []int{2},
		[]float64{100, 200},
		map[string][]float64{"temps": {2020, 2021}}, "pluie")

	ds, err := NewDataset(map[string]*DataArray[float64]{"temperature": temp, "pluie": pluie})
	if err != nil {
		t.Fatalf("NewDataset : %v", err)
	}
	return ds
}

func TestNewDataset(t *testing.T) {
	ds := exempleDataset(t)
	if !reflect.DeepEqual(ds.VarNames(), []string{"pluie", "temperature"}) {
		t.Errorf("VarNames = %v", ds.VarNames())
	}
	dims := ds.Dims()
	if dims["temps"] != 2 || dims["lieu"] != 3 {
		t.Errorf("Dims = %v", dims)
	}
	if c, _ := ds.Coord("temps"); !reflect.DeepEqual(c, []float64{2020, 2021}) {
		t.Errorf("Coord(temps) = %v", c)
	}
}

func TestNewDatasetIncoherent(t *testing.T) {
	// Même dimension "temps" mais tailles différentes -> erreur.
	a, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{1, 2}, nil, "a")
	b, _ := NewDataArray([]string{"temps"}, []int{3}, []float64{1, 2, 3}, nil, "b")
	if _, err := NewDataset(map[string]*DataArray[float64]{"a": a, "b": b}); err == nil {
		t.Error("erreur attendue : dimension temps de tailles incohérentes")
	}

	// Coordonnées incohérentes pour la même dimension.
	c, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{1, 2},
		map[string][]float64{"temps": {2020, 2021}}, "c")
	d, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{1, 2},
		map[string][]float64{"temps": {1999, 2000}}, "d")
	if _, err := NewDataset(map[string]*DataArray[float64]{"c": c, "d": d}); err == nil {
		t.Error("erreur attendue : coordonnées temps incohérentes")
	}
}

func TestDatasetIsel(t *testing.T) {
	ds := exempleDataset(t)
	// Sélection temps=index 1 : propagée aux deux variables.
	sub, err := ds.Isel("temps", 1)
	if err != nil {
		t.Fatalf("Isel : %v", err)
	}
	tp, _ := sub.Get("temperature")
	if !reflect.DeepEqual(tp.Data(), []float64{4, 5, 6}) {
		t.Errorf("temperature = %v, attendu [4 5 6]", tp.Data())
	}
	pl, _ := sub.Get("pluie")
	if !reflect.DeepEqual(pl.Data(), []float64{200}) {
		t.Errorf("pluie = %v, attendu [200]", pl.Data())
	}
	// La dimension temps a disparu.
	if _, ok := sub.Dims()["temps"]; ok {
		t.Error("la dimension temps aurait dû disparaître")
	}
}

func TestDatasetSel(t *testing.T) {
	ds := exempleDataset(t)
	sub, err := ds.Sel("temps", 2020)
	if err != nil {
		t.Fatalf("Sel : %v", err)
	}
	tp, _ := sub.Get("temperature")
	if !reflect.DeepEqual(tp.Data(), []float64{1, 2, 3}) {
		t.Errorf("temperature = %v, attendu [1 2 3]", tp.Data())
	}
	if _, err := ds.Sel("temps", 1900); err == nil {
		t.Error("erreur attendue : étiquette absente")
	}
}

func TestDatasetReductionPropagee(t *testing.T) {
	ds := exempleDataset(t)
	// Moyenne sur temps : temperature [(1+4)/2,(2+5)/2,(3+6)/2]=[2.5 3.5 4.5]
	//                     pluie [(100+200)/2]=[150]
	m, err := ds.MeanAxis("temps")
	if err != nil {
		t.Fatalf("MeanAxis : %v", err)
	}
	tp, _ := m.Get("temperature")
	if !reflect.DeepEqual(tp.Data(), []float64{2.5, 3.5, 4.5}) {
		t.Errorf("temperature = %v", tp.Data())
	}
	pl, _ := m.Get("pluie")
	if !reflect.DeepEqual(pl.Data(), []float64{150}) {
		t.Errorf("pluie = %v", pl.Data())
	}
}

func TestDatasetReductionDimPartielle(t *testing.T) {
	ds := exempleDataset(t)
	// Réduction sur "lieu" : ne concerne que temperature ; pluie inchangée.
	s, err := ds.SumAxis("lieu")
	if err != nil {
		t.Fatalf("SumAxis : %v", err)
	}
	tp, _ := s.Get("temperature") // [1+2+3, 4+5+6] = [6 15]
	if !reflect.DeepEqual(tp.Data(), []float64{6, 15}) {
		t.Errorf("temperature = %v, attendu [6 15]", tp.Data())
	}
	pl, _ := s.Get("pluie") // inchangée
	if !reflect.DeepEqual(pl.Data(), []float64{100, 200}) {
		t.Errorf("pluie = %v, attendu [100 200]", pl.Data())
	}
}

func TestDatasetWithVarDrop(t *testing.T) {
	ds := exempleDataset(t)
	vent, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{5, 7},
		map[string][]float64{"temps": {2020, 2021}}, "vent")
	ds2, err := ds.WithVar("vent", vent)
	if err != nil {
		t.Fatalf("WithVar : %v", err)
	}
	if !reflect.DeepEqual(ds2.VarNames(), []string{"pluie", "temperature", "vent"}) {
		t.Errorf("VarNames = %v", ds2.VarNames())
	}
	ds3, _ := ds2.DropVars("pluie", "vent")
	if !reflect.DeepEqual(ds3.VarNames(), []string{"temperature"}) {
		t.Errorf("VarNames = %v", ds3.VarNames())
	}
}

func TestDatasetMerge(t *testing.T) {
	ds := exempleDataset(t)
	autre, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{9, 9},
		map[string][]float64{"temps": {2020, 2021}}, "humidite")
	dsB, _ := NewDataset(map[string]*DataArray[float64]{"humidite": autre})
	m, err := ds.Merge(dsB)
	if err != nil {
		t.Fatalf("Merge : %v", err)
	}
	if !reflect.DeepEqual(m.VarNames(), []string{"humidite", "pluie", "temperature"}) {
		t.Errorf("VarNames = %v", m.VarNames())
	}
}

func TestDatasetIselDimInconnue(t *testing.T) {
	ds := exempleDataset(t)
	if _, err := ds.Isel("inconnue", 0); err == nil {
		t.Error("erreur attendue : dimension inconnue")
	}
}
