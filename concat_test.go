package xarray

import (
	"reflect"
	"testing"
)

func TestConcat1D(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2},
		map[string][]float64{"x": {0, 1}}, "v")
	b, _ := NewDataArray([]string{"x"}, []int{3}, []float64{3, 4, 5},
		map[string][]float64{"x": {2, 3, 4}}, "v")

	c, err := Concat([]*DataArray[float64]{a, b}, "x")
	if err != nil {
		t.Fatalf("Concat : %v", err)
	}
	if !reflect.DeepEqual(c.Data(), []float64{1, 2, 3, 4, 5}) {
		t.Errorf("Data = %v", c.Data())
	}
	if crd, _ := c.Coord("x"); !reflect.DeepEqual(crd, []float64{0, 1, 2, 3, 4}) {
		t.Errorf("Coord = %v", crd)
	}
}

func TestConcat2DAxe0(t *testing.T) {
	// Concat sur la première dimension (lignes).
	a, _ := NewDataArray([]string{"r", "c"}, []int{1, 3}, []float64{1, 2, 3},
		map[string][]float64{"r": {0}, "c": {10, 20, 30}}, "v")
	b, _ := NewDataArray([]string{"r", "c"}, []int{2, 3}, []float64{4, 5, 6, 7, 8, 9},
		map[string][]float64{"r": {1, 2}, "c": {10, 20, 30}}, "v")
	c, err := Concat([]*DataArray[float64]{a, b}, "r")
	if err != nil {
		t.Fatalf("Concat : %v", err)
	}
	if !reflect.DeepEqual(c.Shape(), []int{3, 3}) {
		t.Errorf("Shape = %v", c.Shape())
	}
	if !reflect.DeepEqual(c.Data(), []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}) {
		t.Errorf("Data = %v", c.Data())
	}
	if crd, _ := c.Coord("r"); !reflect.DeepEqual(crd, []float64{0, 1, 2}) {
		t.Errorf("Coord(r) = %v", crd)
	}
}

func TestConcat2DAxe1(t *testing.T) {
	// Concat sur la seconde dimension (colonnes) : entrelacement.
	a, _ := NewDataArray([]string{"r", "c"}, []int{2, 1}, []float64{1, 3},
		map[string][]float64{"r": {0, 1}, "c": {0}}, "v")
	b, _ := NewDataArray([]string{"r", "c"}, []int{2, 2}, []float64{2, 9, 4, 8},
		map[string][]float64{"r": {0, 1}, "c": {1, 2}}, "v")
	// a = [[1],[3]] ; b = [[2 9],[4 8]] -> [[1 2 9],[3 4 8]]
	c, err := Concat([]*DataArray[float64]{a, b}, "c")
	if err != nil {
		t.Fatalf("Concat : %v", err)
	}
	if !reflect.DeepEqual(c.Shape(), []int{2, 3}) {
		t.Errorf("Shape = %v", c.Shape())
	}
	if !reflect.DeepEqual(c.Data(), []float64{1, 2, 9, 3, 4, 8}) {
		t.Errorf("Data = %v, attendu [1 2 9 3 4 8]", c.Data())
	}
}

func TestConcatIncompatible(t *testing.T) {
	a, _ := NewDataArray([]string{"r", "c"}, []int{2, 2}, []float64{1, 2, 3, 4}, nil, "v")
	b, _ := NewDataArray([]string{"r", "c"}, []int{2, 3}, []float64{1, 2, 3, 4, 5, 6}, nil, "v")
	// Concat sur r : les tailles de c diffèrent (2 vs 3) -> erreur.
	if _, err := Concat([]*DataArray[float64]{a, b}, "r"); err == nil {
		t.Error("erreur attendue : tailles incompatibles")
	}
	if _, err := Concat([]*DataArray[float64]{a, b}, "z"); err == nil {
		t.Error("erreur attendue : dimension inconnue")
	}
	if _, err := Concat([]*DataArray[float64]{}, "r"); err == nil {
		t.Error("erreur attendue : ensemble vide")
	}
}

func TestStack(t *testing.T) {
	// Empile trois tranches 1D sur une nouvelle dimension.
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2},
		map[string][]float64{"x": {0, 1}}, "v")
	b, _ := NewDataArray([]string{"x"}, []int{2}, []float64{3, 4},
		map[string][]float64{"x": {0, 1}}, "v")
	c, _ := NewDataArray([]string{"x"}, []int{2}, []float64{5, 6},
		map[string][]float64{"x": {0, 1}}, "v")

	s, err := Stack([]*DataArray[float64]{a, b, c}, "essai", []float64{10, 20, 30})
	if err != nil {
		t.Fatalf("Stack : %v", err)
	}
	if !reflect.DeepEqual(s.Dims(), []string{"essai", "x"}) {
		t.Errorf("Dims = %v", s.Dims())
	}
	if !reflect.DeepEqual(s.Shape(), []int{3, 2}) {
		t.Errorf("Shape = %v", s.Shape())
	}
	if !reflect.DeepEqual(s.Data(), []float64{1, 2, 3, 4, 5, 6}) {
		t.Errorf("Data = %v", s.Data())
	}
	if crd, _ := s.Coord("essai"); !reflect.DeepEqual(crd, []float64{10, 20, 30}) {
		t.Errorf("Coord(essai) = %v", crd)
	}
	// La coordonnée x est conservée.
	if crd, _ := s.Coord("x"); !reflect.DeepEqual(crd, []float64{0, 1}) {
		t.Errorf("Coord(x) = %v", crd)
	}
}

func TestStackErreurs(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2}, nil, "v")
	if _, err := Stack([]*DataArray[float64]{a}, "x", []float64{0}); err == nil {
		t.Error("erreur attendue : la dimension x existe déjà")
	}
	if _, err := Stack([]*DataArray[float64]{a}, "n", []float64{0, 1}); err == nil {
		t.Error("erreur attendue : nombre d'étiquettes incohérent")
	}
}
