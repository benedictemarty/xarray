package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestWhereFunc(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{-1, 2, -3, 4, -5}, nil, "v")
	// Garder les positifs, remplacer le reste par 0.
	r := da.WhereFunc(func(v float64) bool { return v > 0 }, 0)
	if !reflect.DeepEqual(r.Data(), []float64{0, 2, 0, 4, 0}) {
		t.Errorf("WhereFunc = %v, attendu [0 2 0 4 0]", r.Data())
	}
	// Original inchangé.
	if da.Data()[0] != -1 {
		t.Error("original modifié")
	}
}

func TestWhereMask(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{10, 20, 30, 40}, nil, "v")
	mask, _ := NewDataArray([]string{"x"}, []int{4}, []float64{1, 0, 1, 0}, nil, "m")
	r, err := da.Where(mask, -1)
	if err != nil {
		t.Fatalf("Where : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{10, -1, 30, -1}) {
		t.Errorf("Where = %v, attendu [10 -1 30 -1]", r.Data())
	}
	// Masque de mauvaise forme.
	bad, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 0, 1}, nil, "m")
	if _, err := da.Where(bad, 0); err == nil {
		t.Error("erreur attendue : forme du masque")
	}
}

func TestInterpolateNA(t *testing.T) {
	nan := math.NaN()
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{1, nan, nan, 4},
		map[string][]float64{"x": {0, 1, 2, 3}}, "v")
	r, err := da.InterpolateNA("x")
	if err != nil {
		t.Fatalf("InterpolateNA : %v", err)
	}
	// interpolation linéaire 1 -> 4 : [1 2 3 4]
	if !reflect.DeepEqual(r.Data(), []float64{1, 2, 3, 4}) {
		t.Errorf("InterpolateNA = %v, attendu [1 2 3 4]", r.Data())
	}
}

func TestInterpolateNABords(t *testing.T) {
	nan := math.NaN()
	// NaN de bord conservés ; interpolation seulement au milieu.
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{nan, 2, nan, 6, nan}, nil, "v")
	r, _ := da.InterpolateNA("x")
	got := r.Data()
	if !math.IsNaN(got[0]) || got[1] != 2 || got[2] != 4 || got[3] != 6 || !math.IsNaN(got[4]) {
		t.Errorf("InterpolateNA bords = %v, attendu [NaN 2 4 6 NaN]", got)
	}
}

func TestInterpolateNACoordNonUniforme(t *testing.T) {
	nan := math.NaN()
	// coordonnée non uniforme : interpolation pondérée par la coordonnée.
	// x=[0 1 4], valeurs [0 NaN 8] -> à x=1 : 0 + (8-0)*(1-0)/(4-0) = 2
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{0, nan, 8},
		map[string][]float64{"x": {0, 1, 4}}, "v")
	r, _ := da.InterpolateNA("x")
	if r.Data()[1] != 2 {
		t.Errorf("InterpolateNA coord = %v, attendu [0 2 8]", r.Data())
	}
}
