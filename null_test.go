package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestIsNullNotNull(t *testing.T) {
	nan := math.NaN()
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{1, nan, 3, nan, 5}, nil, "v")
	if !reflect.DeepEqual(da.IsNull().Data(), []float64{0, 1, 0, 1, 0}) {
		t.Errorf("IsNull = %v", da.IsNull().Data())
	}
	if !reflect.DeepEqual(da.NotNull().Data(), []float64{1, 0, 1, 0, 1}) {
		t.Errorf("NotNull = %v", da.NotNull().Data())
	}
	if da.Count() != 3 {
		t.Errorf("Count = %d, attendu 3", da.Count())
	}
}

func TestCountAxis(t *testing.T) {
	nan := math.NaN()
	// [[1 NaN 3],[NaN NaN 6]] dims g,x
	da, _ := NewDataArray([]string{"g", "x"}, []int{2, 3},
		[]float64{1, nan, 3, nan, nan, 6}, nil, "v")
	// non-NaN le long de x : ligne0 -> 2, ligne1 -> 1
	cx, err := da.CountAxis("x")
	if err != nil {
		t.Fatalf("CountAxis : %v", err)
	}
	if !reflect.DeepEqual(cx.Data(), []float64{2, 1}) {
		t.Errorf("CountAxis(x) = %v, attendu [2 1]", cx.Data())
	}
	// non-NaN le long de g : colonnes -> [1, 0, 2]
	cg, _ := da.CountAxis("g")
	if !reflect.DeepEqual(cg.Data(), []float64{1, 0, 2}) {
		t.Errorf("CountAxis(g) = %v, attendu [1 0 2]", cg.Data())
	}
}

func TestNullEntier(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []int{1, 2, 3}, nil, "v")
	if da.Count() != 3 {
		t.Errorf("Count[int] = %d", da.Count())
	}
	if !reflect.DeepEqual(da.IsNull().Data(), []int{0, 0, 0}) {
		t.Errorf("IsNull[int] = %v", da.IsNull().Data())
	}
}
