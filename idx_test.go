package xarray

import (
	"reflect"
	"testing"
)

func TestIdxMinMax1D(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{3, 1, 2},
		map[string][]float64{"x": {10, 20, 30}}, "v")
	mn, err := da.IdxMinAxis("x")
	if err != nil {
		t.Fatalf("IdxMinAxis : %v", err)
	}
	// min à l'indice 1 -> étiquette 20
	if len(mn.Data()) != 1 || mn.Data()[0] != 20 {
		t.Errorf("IdxMin = %v, attendu [20]", mn.Data())
	}
	mx, _ := da.IdxMaxAxis("x")
	// max à l'indice 0 -> étiquette 10
	if mx.Data()[0] != 10 {
		t.Errorf("IdxMax = %v, attendu [10]", mx.Data())
	}
}

func TestIdxMinMax2D(t *testing.T) {
	// [[3 1 2],[5 9 4]] dims g,x ; coord x=[10 20 30]
	da, _ := NewDataArray([]string{"g", "x"}, []int{2, 3},
		[]float64{3, 1, 2, 5, 9, 4},
		map[string][]float64{"x": {10, 20, 30}}, "v")
	mn, _ := da.IdxMinAxis("x")
	// ligne 0 : min idx1 -> 20 ; ligne 1 : min idx2 -> 30
	if !reflect.DeepEqual(mn.Data(), []float64{20, 30}) {
		t.Errorf("IdxMin 2D = %v, attendu [20 30]", mn.Data())
	}
}

func TestIdxSansCoord(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{3, 1, 2}, nil, "v")
	if _, err := da.IdxMinAxis("x"); err == nil {
		t.Error("erreur attendue : pas de coordonnée")
	}
}
