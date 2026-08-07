package xarray

import (
	"reflect"
	"testing"
)

func TestSqueeze(t *testing.T) {
	// (1, 3) -> (3) en supprimant la dimension t de taille 1.
	da, _ := NewDataArray([]string{"t", "x"}, []int{1, 3}, []float64{1, 2, 3},
		map[string][]float64{"t": {2020}, "x": {10, 20, 30}}, "v")
	s, err := da.Squeeze("t")
	if err != nil {
		t.Fatalf("Squeeze : %v", err)
	}
	if !reflect.DeepEqual(s.Dims(), []string{"x"}) {
		t.Errorf("Dims = %v", s.Dims())
	}
	if !reflect.DeepEqual(s.Data(), []float64{1, 2, 3}) {
		t.Errorf("Data = %v", s.Data())
	}
	// La coordonnée t a disparu, x subsiste.
	if _, err := s.Coord("t"); err == nil {
		t.Error("coord t aurait dû disparaître")
	}
	if c, _ := s.Coord("x"); !reflect.DeepEqual(c, []float64{10, 20, 30}) {
		t.Errorf("coord x = %v", c)
	}
	// Squeeze d'une dimension non unitaire -> erreur.
	if _, err := da.Squeeze("x"); err == nil {
		t.Error("erreur attendue : dimension non unitaire")
	}
}

func TestExpandDims(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"x": {10, 20, 30}}, "v")
	e, err := da.ExpandDims("membre")
	if err != nil {
		t.Fatalf("ExpandDims : %v", err)
	}
	if !reflect.DeepEqual(e.Dims(), []string{"membre", "x"}) {
		t.Errorf("Dims = %v", e.Dims())
	}
	if !reflect.DeepEqual(e.Shape(), []int{1, 3}) {
		t.Errorf("Shape = %v", e.Shape())
	}
	if !reflect.DeepEqual(e.Data(), []float64{1, 2, 3}) {
		t.Errorf("Data = %v", e.Data())
	}
	// Aller-retour ExpandDims -> Squeeze.
	back, _ := e.Squeeze("membre")
	if !reflect.DeepEqual(back.Dims(), []string{"x"}) {
		t.Errorf("Squeeze après Expand : %v", back.Dims())
	}
	// Dimension déjà présente.
	if _, err := da.ExpandDims("x"); err == nil {
		t.Error("erreur attendue : dimension existante")
	}
}

func TestRenameDim(t *testing.T) {
	da, _ := NewDataArray([]string{"t", "x"}, []int{2, 2}, []float64{1, 2, 3, 4},
		map[string][]float64{"t": {0, 1}, "x": {10, 20}}, "v")
	r, err := da.RenameDim("t", "temps")
	if err != nil {
		t.Fatalf("RenameDim : %v", err)
	}
	if !reflect.DeepEqual(r.Dims(), []string{"temps", "x"}) {
		t.Errorf("Dims = %v", r.Dims())
	}
	// La coordonnée a suivi le renommage.
	if c, _ := r.Coord("temps"); !reflect.DeepEqual(c, []float64{0, 1}) {
		t.Errorf("coord temps = %v", c)
	}
	if _, err := r.Coord("t"); err == nil {
		t.Error("l'ancienne coord t ne devrait plus exister")
	}
	// Sel par label sur la dimension renommée.
	sub, _ := r.Sel("temps", 1)
	if !reflect.DeepEqual(sub.Data(), []float64{3, 4}) {
		t.Errorf("Sel après rename = %v", sub.Data())
	}
	// Renommage vers un nom existant.
	if _, err := da.RenameDim("t", "x"); err == nil {
		t.Error("erreur attendue : nom existant")
	}
}
