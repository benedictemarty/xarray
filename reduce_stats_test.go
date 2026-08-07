package xarray

import (
	"reflect"
	"testing"
)

func TestArgMinMaxAxis(t *testing.T) {
	// [[3 1 2],[5 9 4]] dims g,x
	da, _ := NewDataArray([]string{"g", "x"}, []int{2, 3},
		[]float64{3, 1, 2, 5, 9, 4}, nil, "v")
	mn, err := da.ArgMinAxis("x")
	if err != nil {
		t.Fatalf("ArgMinAxis : %v", err)
	}
	// ligne 0 : min à l'indice 1 ; ligne 1 : min à l'indice 2
	if !reflect.DeepEqual(mn.Data(), []float64{1, 2}) {
		t.Errorf("ArgMinAxis = %v, attendu [1 2]", mn.Data())
	}
	mx, _ := da.ArgMaxAxis("x")
	// ligne 0 : max à l'indice 0 ; ligne 1 : max à l'indice 1
	if !reflect.DeepEqual(mx.Data(), []float64{0, 1}) {
		t.Errorf("ArgMaxAxis = %v, attendu [0 1]", mx.Data())
	}
}

func TestQuantile(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{0, 1, 2, 3, 4}, nil, "v")
	if da.Quantile(0.5) != 2 {
		t.Errorf("médiane (q=0.5) = %v, attendu 2", da.Quantile(0.5))
	}
	if da.Quantile(0) != 0 || da.Quantile(1) != 4 {
		t.Errorf("q=0/1 = %v/%v", da.Quantile(0), da.Quantile(1))
	}
	// q=0.25 : position 0.25*4=1 -> valeur 1
	if da.Quantile(0.25) != 1 {
		t.Errorf("q=0.25 = %v, attendu 1", da.Quantile(0.25))
	}
	// q=0.75 : position 0.75*4=3 -> valeur 3
	if da.Quantile(0.75) != 3 {
		t.Errorf("q=0.75 = %v, attendu 3", da.Quantile(0.75))
	}
}

func TestQuantileInterpolation(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{0, 1, 2, 3}, nil, "v")
	// q=0.5 : position 1.5 -> (1+2)/2 = 1.5
	if da.Quantile(0.5) != 1.5 {
		t.Errorf("q=0.5 = %v, attendu 1.5", da.Quantile(0.5))
	}
}

func TestCumprod(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{1, 2, 3, 4}, nil, "v")
	c, err := da.Cumprod("x")
	if err != nil {
		t.Fatalf("Cumprod : %v", err)
	}
	if !reflect.DeepEqual(c.Data(), []float64{1, 2, 6, 24}) {
		t.Errorf("Cumprod = %v, attendu [1 2 6 24]", c.Data())
	}
}
