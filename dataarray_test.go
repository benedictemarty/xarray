package xarray

import (
	"math"
	"reflect"
	"testing"
)

func exempleDataArray(t *testing.T) *DataArray {
	t.Helper()
	da, err := NewDataArray(
		[]string{"temps", "lieu"},
		[]int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{
			"temps": {2020, 2021},
			"lieu":  {10, 20, 30},
		},
		"température",
	)
	if err != nil {
		t.Fatalf("construction impossible : %v", err)
	}
	return da
}

func TestNewDataArray(t *testing.T) {
	da := exempleDataArray(t)
	if da.Name() != "température" {
		t.Errorf("Name = %q", da.Name())
	}
	if !reflect.DeepEqual(da.Dims(), []string{"temps", "lieu"}) {
		t.Errorf("Dims = %v", da.Dims())
	}
	labels, err := da.Coord("lieu")
	if err != nil {
		t.Fatalf("Coord : %v", err)
	}
	if !reflect.DeepEqual(labels, []float64{10, 20, 30}) {
		t.Errorf("Coord(lieu) = %v", labels)
	}
}

func TestNewDataArrayCoordInvalide(t *testing.T) {
	_, err := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2},
		map[string][]float64{"x": {1, 2, 3}}, "")
	if err == nil {
		t.Error("erreur attendue : coordonnée de mauvaise longueur")
	}
	_, err = NewDataArray([]string{"x"}, []int{2}, []float64{1, 2},
		map[string][]float64{"z": {1, 2}}, "")
	if err == nil {
		t.Error("erreur attendue : coordonnée sans dimension")
	}
}

func TestDataArrayIsel(t *testing.T) {
	da := exempleDataArray(t)
	sub, err := da.Isel("temps", 1)
	if err != nil {
		t.Fatalf("Isel : %v", err)
	}
	if !reflect.DeepEqual(sub.Dims(), []string{"lieu"}) {
		t.Errorf("Dims = %v, attendu [lieu]", sub.Dims())
	}
	if !reflect.DeepEqual(sub.Data(), []float64{4, 5, 6}) {
		t.Errorf("Data = %v, attendu [4 5 6]", sub.Data())
	}
	// La coordonnée « temps » doit avoir disparu, « lieu » subsister.
	if _, err := sub.Coord("temps"); err == nil {
		t.Error("la coordonnée temps aurait dû disparaître")
	}
	if _, err := sub.Coord("lieu"); err != nil {
		t.Error("la coordonnée lieu aurait dû subsister")
	}
}

func TestDataArraySel(t *testing.T) {
	da := exempleDataArray(t)
	sub, err := da.Sel("temps", 2021)
	if err != nil {
		t.Fatalf("Sel : %v", err)
	}
	if !reflect.DeepEqual(sub.Data(), []float64{4, 5, 6}) {
		t.Errorf("Data = %v, attendu [4 5 6]", sub.Data())
	}

	if _, err := da.Sel("temps", 1999); err == nil {
		t.Error("erreur attendue : étiquette absente")
	}
	if _, err := da.Sel("inconnue", 0); err == nil {
		t.Error("erreur attendue : dimension sans coordonnée")
	}
}

func TestDataArrayReductions(t *testing.T) {
	da := exempleDataArray(t)
	if da.Sum() != 21 {
		t.Errorf("Sum = %v, attendu 21", da.Sum())
	}
	if da.Mean() != 3.5 {
		t.Errorf("Mean = %v, attendu 3.5", da.Mean())
	}
	if da.Min() != 1 {
		t.Errorf("Min = %v, attendu 1", da.Min())
	}
	if da.Max() != 6 {
		t.Errorf("Max = %v, attendu 6", da.Max())
	}
}

func TestDataArrayReductionsVide(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{0}, []float64{}, nil, "vide")
	if !math.IsNaN(da.Mean()) {
		t.Errorf("Mean d'un tableau vide devrait être NaN")
	}
	if !math.IsNaN(da.Min()) {
		t.Errorf("Min d'un tableau vide devrait être NaN")
	}
}

func TestDataArrayRename(t *testing.T) {
	da := exempleDataArray(t)
	r := da.Rename("t2m")
	if r.Name() != "t2m" {
		t.Errorf("Rename : %q", r.Name())
	}
	if da.Name() != "température" {
		t.Errorf("l'original ne doit pas être modifié : %q", da.Name())
	}
}
