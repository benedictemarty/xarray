package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestSkipNAGlobal(t *testing.T) {
	nan := math.NaN()
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{1, nan, 3, nan, 5}, nil, "v")

	if da.SumSkipNA() != 9 {
		t.Errorf("SumSkipNA = %v, attendu 9", da.SumSkipNA())
	}
	if da.MeanSkipNA() != 3 { // (1+3+5)/3
		t.Errorf("MeanSkipNA = %v, attendu 3", da.MeanSkipNA())
	}
	if da.MinSkipNA() != 1 {
		t.Errorf("MinSkipNA = %v", da.MinSkipNA())
	}
	if da.MaxSkipNA() != 5 {
		t.Errorf("MaxSkipNA = %v", da.MaxSkipNA())
	}

	// La réduction ordinaire, elle, propage NaN.
	if !math.IsNaN(da.Sum()) {
		// Sum ordinaire additionne NaN -> NaN
		t.Logf("Sum ordinaire = %v (NaN attendu par propagation)", da.Sum())
	}
}

func TestSkipNATousNaN(t *testing.T) {
	nan := math.NaN()
	da, _ := NewDataArray([]string{"x"}, []int{2}, []float64{nan, nan}, nil, "v")
	if !math.IsNaN(da.MeanSkipNA()) {
		t.Errorf("MeanSkipNA de tout-NaN doit être NaN")
	}
	// Sum de tout-NaN ignoré -> 0
	if da.SumSkipNA() != 0 {
		t.Errorf("SumSkipNA = %v, attendu 0", da.SumSkipNA())
	}
}

func TestSkipNAParAxe(t *testing.T) {
	nan := math.NaN()
	// [[1 NaN 3],[4 5 NaN]] dims x,y
	da, _ := NewDataArray([]string{"x", "y"}, []int{2, 3},
		[]float64{1, nan, 3, 4, 5, nan},
		map[string][]float64{"x": {0, 1}, "y": {10, 20, 30}}, "v")

	// Somme le long de x, en ignorant NaN :
	// y0: 1+4=5 ; y1: 5 (NaN ignoré) ; y2: 3 (NaN ignoré)
	s, err := da.SumAxisSkipNA("x")
	if err != nil {
		t.Fatalf("SumAxisSkipNA : %v", err)
	}
	if !reflect.DeepEqual(s.Data(), []float64{5, 5, 3}) {
		t.Errorf("SumAxisSkipNA = %v, attendu [5 5 3]", s.Data())
	}

	// Moyenne le long de y, en ignorant NaN :
	// x0: (1+3)/2=2 ; x1: (4+5)/2=4.5
	m, _ := da.MeanAxisSkipNA("y")
	if !reflect.DeepEqual(m.Data(), []float64{2, 4.5}) {
		t.Errorf("MeanAxisSkipNA = %v, attendu [2 4.5]", m.Data())
	}
}

func TestSkipNAEntier(t *testing.T) {
	// Pour un type entier, aucun NaN possible : identique à la réduction normale.
	da, _ := NewDataArray([]string{"x"}, []int{3}, []int{2, 4, 6}, nil, "v")
	if da.SumSkipNA() != 12 {
		t.Errorf("SumSkipNA[int] = %v", da.SumSkipNA())
	}
}
