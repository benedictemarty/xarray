package xarray

import (
	"bytes"
	"reflect"
	"testing"
)

// Ces tests valident que la bibliothèque fonctionne avec un type numérique
// autre que float64 (ici int), objectif du Sprint 5 (dette T-01).

func TestGenericInt(t *testing.T) {
	da, err := NewDataArray([]string{"x", "y"}, []int{2, 3},
		[]int{1, 2, 3, 4, 5, 6},
		map[string][]int{"x": {0, 1}, "y": {10, 20, 30}}, "compteur")
	if err != nil {
		t.Fatalf("NewDataArray[int] : %v", err)
	}
	if da.Sum() != 21 {
		t.Errorf("Sum = %d, attendu 21", da.Sum())
	}
	if da.Min() != 1 || da.Max() != 6 {
		t.Errorf("Min/Max = %d/%d", da.Min(), da.Max())
	}
	// Mean renvoie float64 même pour un type entier.
	if da.Mean() != 3.5 {
		t.Errorf("Mean = %v, attendu 3.5", da.Mean())
	}

	// Sélection par label (int).
	sub, err := da.Sel("y", 20)
	if err != nil {
		t.Fatalf("Sel : %v", err)
	}
	if !reflect.DeepEqual(sub.Data(), []int{2, 5}) {
		t.Errorf("Sel(y,20) = %v, attendu [2 5]", sub.Data())
	}

	// SumAxis conserve le type int ; MeanAxis passe en float64.
	sx, _ := da.SumAxis("x")
	if !reflect.DeepEqual(sx.Data(), []int{5, 7, 9}) {
		t.Errorf("SumAxis(x) = %v", sx.Data())
	}
	mx, _ := da.MeanAxis("x")
	if !reflect.DeepEqual(mx.Data(), []float64{2.5, 3.5, 4.5}) {
		t.Errorf("MeanAxis(x) = %v", mx.Data())
	}
}

func TestGenericIntArithmetique(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{3}, []int{1, 2, 3}, nil, "a")
	b, _ := NewDataArray([]string{"x"}, []int{3}, []int{10, 20, 30}, nil, "b")
	s, _ := a.Add(b)
	if !reflect.DeepEqual(s.Data(), []int{11, 22, 33}) {
		t.Errorf("Add = %v", s.Data())
	}
	// Division entière : comportement propre au type int.
	d, _ := b.Div(a)
	if !reflect.DeepEqual(d.Data(), []int{10, 10, 10}) {
		t.Errorf("Div = %v", d.Data())
	}
}

func TestGenericIntIO(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []int{7, 8, 9},
		map[string][]int{"x": {0, 1, 2}}, "v")

	// JSON aller-retour.
	var jbuf bytes.Buffer
	if err := da.WriteJSON(&jbuf); err != nil {
		t.Fatalf("WriteJSON : %v", err)
	}
	gj, err := ReadDataArrayJSON[int](&jbuf)
	if err != nil {
		t.Fatalf("ReadDataArrayJSON[int] : %v", err)
	}
	if !reflect.DeepEqual(gj.Data(), []int{7, 8, 9}) {
		t.Errorf("JSON Data = %v", gj.Data())
	}

	// CSV aller-retour.
	var cbuf bytes.Buffer
	if err := da.WriteCSV(&cbuf); err != nil {
		t.Fatalf("WriteCSV : %v", err)
	}
	gc, err := ReadDataArrayCSV[int](&cbuf)
	if err != nil {
		t.Fatalf("ReadDataArrayCSV[int] : %v", err)
	}
	if !reflect.DeepEqual(gc.Data(), []int{7, 8, 9}) {
		t.Errorf("CSV Data = %v", gc.Data())
	}
}

func TestGenericFloat32Dataset(t *testing.T) {
	a, _ := NewDataArray([]string{"t"}, []int{2}, []float32{1.5, 2.5},
		map[string][]float32{"t": {0, 1}}, "a")
	ds, err := NewDataset(map[string]*DataArray[float32]{"a": a})
	if err != nil {
		t.Fatalf("NewDataset[float32] : %v", err)
	}
	sub, _ := ds.Sel("t", 1)
	got, _ := sub.Get("a")
	if !reflect.DeepEqual(got.Data(), []float32{2.5}) {
		t.Errorf("Sel = %v", got.Data())
	}
}
