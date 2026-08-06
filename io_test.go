package xarray

import (
	"bytes"
	"reflect"
	"testing"
)

func TestJSONDataArrayAllerRetour(t *testing.T) {
	da, _ := NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")

	var buf bytes.Buffer
	if err := da.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON : %v", err)
	}
	got, err := ReadDataArrayJSON[float64](&buf)
	if err != nil {
		t.Fatalf("ReadDataArrayJSON : %v", err)
	}
	if got.Name() != "temperature" {
		t.Errorf("Name = %q", got.Name())
	}
	if !reflect.DeepEqual(got.Dims(), da.Dims()) {
		t.Errorf("Dims = %v", got.Dims())
	}
	if !reflect.DeepEqual(got.Data(), da.Data()) {
		t.Errorf("Data = %v", got.Data())
	}
	c, _ := got.Coord("temps")
	if !reflect.DeepEqual(c, []float64{2020, 2021}) {
		t.Errorf("Coord = %v", c)
	}
}

func TestJSONDatasetAllerRetour(t *testing.T) {
	temp, _ := NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")
	pluie, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{100, 200},
		map[string][]float64{"temps": {2020, 2021}}, "pluie")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"temperature": temp, "pluie": pluie})

	var buf bytes.Buffer
	if err := ds.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON : %v", err)
	}
	got, err := ReadDatasetJSON[float64](&buf)
	if err != nil {
		t.Fatalf("ReadDatasetJSON : %v", err)
	}
	if !reflect.DeepEqual(got.VarNames(), []string{"pluie", "temperature"}) {
		t.Errorf("VarNames = %v", got.VarNames())
	}
	tp, _ := got.Get("temperature")
	if !reflect.DeepEqual(tp.Data(), []float64{1, 2, 3, 4, 5, 6}) {
		t.Errorf("temperature = %v", tp.Data())
	}
	if c, _ := got.Coord("lieu"); !reflect.DeepEqual(c, []float64{10, 20, 30}) {
		t.Errorf("Coord(lieu) = %v", c)
	}
}

func TestCSVDataArrayAllerRetour(t *testing.T) {
	da, _ := NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")

	var buf bytes.Buffer
	if err := da.WriteCSV(&buf); err != nil {
		t.Fatalf("WriteCSV : %v", err)
	}
	got, err := ReadDataArrayCSV[float64](&buf)
	if err != nil {
		t.Fatalf("ReadDataArrayCSV : %v", err)
	}
	if !reflect.DeepEqual(got.Dims(), []string{"temps", "lieu"}) {
		t.Errorf("Dims = %v", got.Dims())
	}
	if !reflect.DeepEqual(got.Shape(), []int{2, 3}) {
		t.Errorf("Shape = %v", got.Shape())
	}
	if !reflect.DeepEqual(got.Data(), []float64{1, 2, 3, 4, 5, 6}) {
		t.Errorf("Data = %v", got.Data())
	}
	if got.Name() != "temperature" {
		t.Errorf("Name = %q", got.Name())
	}
	c, _ := got.Coord("lieu")
	if !reflect.DeepEqual(c, []float64{10, 20, 30}) {
		t.Errorf("Coord(lieu) = %v", c)
	}
}

func TestCSVContenu(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1.5, 2.5},
		map[string][]float64{"x": {0, 1}}, "v")
	var buf bytes.Buffer
	_ = da.WriteCSV(&buf)
	attendu := "x,v\n0,1.5\n1,2.5\n"
	if buf.String() != attendu {
		t.Errorf("CSV = %q, attendu %q", buf.String(), attendu)
	}
}

func TestCSVSansCoord(t *testing.T) {
	// Sans coordonnées : les étiquettes sont les indices 0..n-1.
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{7, 8, 9}, nil, "")
	var buf bytes.Buffer
	_ = da.WriteCSV(&buf)
	got, err := ReadDataArrayCSV[float64](&buf)
	if err != nil {
		t.Fatalf("ReadDataArrayCSV : %v", err)
	}
	if !reflect.DeepEqual(got.Data(), []float64{7, 8, 9}) {
		t.Errorf("Data = %v", got.Data())
	}
	// La colonne valeur "value" doit donner un nom vide.
	if got.Name() != "" {
		t.Errorf("Name = %q, attendu vide", got.Name())
	}
}

func TestCSVErreurs(t *testing.T) {
	if _, err := ReadDataArrayCSV[float64](bytes.NewBufferString("")); err == nil {
		t.Error("erreur attendue : CSV vide")
	}
	if _, err := ReadDataArrayCSV[float64](bytes.NewBufferString("x\n1\n")); err == nil {
		t.Error("erreur attendue : en-tête à une seule colonne")
	}
	// Grille incomplète : coord x a deux valeurs mais une seule ligne.
	if _, err := ReadDataArrayCSV[float64](bytes.NewBufferString("x,y,v\n0,0,1\n0,1,2\n1,0,3\n")); err == nil {
		t.Error("erreur attendue : grille incomplète")
	}
}
