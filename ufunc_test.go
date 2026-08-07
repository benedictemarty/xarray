package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestApplyEtAbs(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{-2, 3, -4, 5},
		map[string][]float64{"x": {0, 1, 2, 3}}, "v")
	ab := da.Abs()
	if !reflect.DeepEqual(ab.Data(), []float64{2, 3, 4, 5}) {
		t.Errorf("Abs = %v", ab.Data())
	}
	// Coordonnées préservées.
	if c, _ := ab.Coord("x"); !reflect.DeepEqual(c, []float64{0, 1, 2, 3}) {
		t.Errorf("coord x = %v", c)
	}
	// Apply personnalisé : x -> x+100.
	r := da.Apply(func(x float64) float64 { return x + 100 })
	if !reflect.DeepEqual(r.Data(), []float64{98, 103, 96, 105}) {
		t.Errorf("Apply = %v", r.Data())
	}
	// Original inchangé.
	if da.Data()[0] != -2 {
		t.Error("original modifié")
	}
}

func TestClip(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{-1, 2, 5, 8, 11}, nil, "v")
	c := da.Clip(0, 10)
	if !reflect.DeepEqual(c.Data(), []float64{0, 2, 5, 8, 10}) {
		t.Errorf("Clip = %v, attendu [0 2 5 8 10]", c.Data())
	}
}

func TestSqrtExpLogPow(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 4, 9}, nil, "v")
	if !reflect.DeepEqual(da.Sqrt().Data(), []float64{1, 2, 3}) {
		t.Errorf("Sqrt = %v", da.Sqrt().Data())
	}
	if !reflect.DeepEqual(da.Pow(2).Data(), []float64{1, 16, 81}) {
		t.Errorf("Pow(2) = %v", da.Pow(2).Data())
	}
	// Exp(Log(x)) ≈ x.
	el := da.Log().Exp().Data()
	for i, x := range []float64{1, 4, 9} {
		if math.Abs(el[i]-x) > 1e-9 {
			t.Errorf("Exp(Log) [%d] = %v, attendu %v", i, el[i], x)
		}
	}
}

func TestUfuncInt(t *testing.T) {
	// Type entier : Abs/Clip cohérents ; Sqrt tronqué.
	da, _ := NewDataArray([]string{"x"}, []int{3}, []int{-5, 3, -1}, nil, "v")
	if !reflect.DeepEqual(da.Abs().Data(), []int{5, 3, 1}) {
		t.Errorf("Abs[int] = %v", da.Abs().Data())
	}
}
