package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestVarStdMedian(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{2, 4, 4, 6}, nil, "v")
	// moyenne 4 ; variance = ((−2)²+0+0+2²)/4 = 8/4 = 2 ; écart-type = √2
	if da.Var() != 2 {
		t.Errorf("Var = %v, attendu 2", da.Var())
	}
	if math.Abs(da.Std()-math.Sqrt2) > 1e-12 {
		t.Errorf("Std = %v, attendu √2", da.Std())
	}
	// médiane de [2 4 4 6] = (4+4)/2 = 4
	if da.Median() != 4 {
		t.Errorf("Median = %v, attendu 4", da.Median())
	}
}

func TestMedianImpair(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{5, 1, 3, 2, 4}, nil, "v")
	if da.Median() != 3 {
		t.Errorf("Median = %v, attendu 3", da.Median())
	}
}

func TestVarAxis(t *testing.T) {
	// [[2 4 4 6],[1 1 1 1]] dims g,x ; variance le long de x
	da, _ := NewDataArray([]string{"g", "x"}, []int{2, 4},
		[]float64{2, 4, 4, 6, 1, 1, 1, 1}, nil, "v")
	v, err := da.VarAxis("x")
	if err != nil {
		t.Fatalf("VarAxis : %v", err)
	}
	// var ligne 0 = 2 ; var ligne 1 = 0
	if !reflect.DeepEqual(v.Data(), []float64{2, 0}) {
		t.Errorf("VarAxis = %v, attendu [2 0]", v.Data())
	}
}

func TestCumsum(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{1, 2, 3, 4}, nil, "v")
	c, err := da.Cumsum("x")
	if err != nil {
		t.Fatalf("Cumsum : %v", err)
	}
	if !reflect.DeepEqual(c.Data(), []float64{1, 3, 6, 10}) {
		t.Errorf("Cumsum = %v, attendu [1 3 6 10]", c.Data())
	}
}

func TestCumsum2D(t *testing.T) {
	// dims [t, x], cumul sur t.
	da, _ := NewDataArray([]string{"t", "x"}, []int{3, 2},
		[]float64{1, 10, 2, 20, 3, 30}, nil, "v")
	c, _ := da.Cumsum("t")
	// colonnes cumulées : [1 10 ; 3 30 ; 6 60]
	if !reflect.DeepEqual(c.Data(), []float64{1, 10, 3, 30, 6, 60}) {
		t.Errorf("Cumsum 2D = %v", c.Data())
	}
}

func TestDiff(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{1, 3, 6, 10},
		map[string][]float64{"x": {0, 1, 2, 3}}, "v")
	d, err := da.Diff("x")
	if err != nil {
		t.Fatalf("Diff : %v", err)
	}
	// différences : [2 3 4]
	if !reflect.DeepEqual(d.Data(), []float64{2, 3, 4}) {
		t.Errorf("Diff = %v, attendu [2 3 4]", d.Data())
	}
	// coordonnée = positions 1..n-1
	if c, _ := d.Coord("x"); !reflect.DeepEqual(c, []float64{1, 2, 3}) {
		t.Errorf("coord = %v, attendu [1 2 3]", c)
	}
}
