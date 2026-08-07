package xarray

import (
	"reflect"
	"testing"
)

func exempleSel(t *testing.T) *DataArray[float64] {
	t.Helper()
	da, _ := NewDataArray([]string{"lieu"}, []int{5}, []float64{10, 20, 30, 40, 50},
		map[string][]float64{"lieu": {0, 10, 20, 30, 40}}, "v")
	return da
}

func TestSelNearest(t *testing.T) {
	da := exempleSel(t)
	// étiquette 23 -> plus proche de 20 (indice 2) -> valeur 30
	r, err := da.SelNearest("lieu", 23)
	if err != nil {
		t.Fatalf("SelNearest : %v", err)
	}
	if len(r.Data()) != 1 || r.Data()[0] != 30 {
		t.Errorf("SelNearest(23) = %v, attendu [30]", r.Data())
	}
	// étiquette 100 -> plus proche de 40 -> valeur 50
	r2, _ := da.SelNearest("lieu", 100)
	if r2.Data()[0] != 50 {
		t.Errorf("SelNearest(100) = %v, attendu [50]", r2.Data())
	}
}

func TestSelNearestKeep(t *testing.T) {
	da := exempleSel(t)
	// 23 -> plus proche de 20 -> valeur 30, mais dimension CONSERVÉE (taille 1)
	r, err := da.SelNearestKeep("lieu", 23)
	if err != nil {
		t.Fatalf("SelNearestKeep : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{30}) {
		t.Errorf("data = %v, attendu [30]", r.Data())
	}
	// la dimension et sa coordonnée doivent survivre (contrairement à SelNearest)
	if c, _ := r.Coord("lieu"); !reflect.DeepEqual(c, []float64{20}) {
		t.Errorf("coord = %v, attendu [20]", c)
	}
	if len(r.Dims()) != 1 {
		t.Errorf("dims = %v, dimension supprimée à tort", r.Dims())
	}
}

func TestSelNearestMany(t *testing.T) {
	da := exempleSel(t)
	// [23, 100] -> plus proches 20 et 40 -> valeurs 30 et 50, ordre conservé
	r, err := da.SelNearestMany("lieu", []float64{23, 100})
	if err != nil {
		t.Fatalf("SelNearestMany : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{30, 50}) {
		t.Errorf("data = %v, attendu [30 50]", r.Data())
	}
	if c, _ := r.Coord("lieu"); !reflect.DeepEqual(c, []float64{20, 40}) {
		t.Errorf("coord = %v, attendu [20 40]", c)
	}
	// liste vide -> erreur
	if _, err := da.SelNearestMany("lieu", nil); err == nil {
		t.Error("erreur attendue : liste vide")
	}
}

func TestSelRange(t *testing.T) {
	da := exempleSel(t)
	// [10, 30] -> étiquettes 10,20,30 -> valeurs 20,30,40
	r, err := da.SelRange("lieu", 10, 30)
	if err != nil {
		t.Fatalf("SelRange : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{20, 30, 40}) {
		t.Errorf("SelRange = %v, attendu [20 30 40]", r.Data())
	}
	if c, _ := r.Coord("lieu"); !reflect.DeepEqual(c, []float64{10, 20, 30}) {
		t.Errorf("coord = %v", c)
	}
	// bornes inversées : même résultat
	r2, _ := da.SelRange("lieu", 30, 10)
	if !reflect.DeepEqual(r2.Data(), []float64{20, 30, 40}) {
		t.Errorf("SelRange inversé = %v", r2.Data())
	}
	// intervalle vide
	if _, err := da.SelRange("lieu", 100, 200); err == nil {
		t.Error("erreur attendue : intervalle vide")
	}
}

func TestSelMany(t *testing.T) {
	da := exempleSel(t)
	// étiquettes [40, 0] -> valeurs [50, 10] (ordre respecté)
	r, err := da.SelMany("lieu", []float64{40, 0})
	if err != nil {
		t.Fatalf("SelMany : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{50, 10}) {
		t.Errorf("SelMany = %v, attendu [50 10]", r.Data())
	}
	if c, _ := r.Coord("lieu"); !reflect.DeepEqual(c, []float64{40, 0}) {
		t.Errorf("coord = %v", c)
	}
	// étiquette absente
	if _, err := da.SelMany("lieu", []float64{99}); err == nil {
		t.Error("erreur attendue : étiquette absente")
	}
}

func TestSelAdvancedSansCoord(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2}, nil, "v")
	if _, err := da.SelNearest("x", 1); err == nil {
		t.Error("erreur attendue : pas de coordonnée")
	}
	if _, err := da.SelRange("x", 0, 1); err == nil {
		t.Error("erreur attendue : pas de coordonnée")
	}
}
