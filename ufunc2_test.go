package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestRoundFloorCeilSign(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float64{-1.6, 2.4, 2.5, -0.0}, nil, "v")
	if !reflect.DeepEqual(da.Round().Data(), []float64{-2, 2, 3, 0}) {
		t.Errorf("Round = %v", da.Round().Data())
	}
	if !reflect.DeepEqual(da.Floor().Data(), []float64{-2, 2, 2, 0}) {
		t.Errorf("Floor = %v", da.Floor().Data())
	}
	if !reflect.DeepEqual(da.Ceil().Data(), []float64{-1, 3, 3, 0}) {
		t.Errorf("Ceil = %v", da.Ceil().Data())
	}
	sg, _ := NewDataArray([]string{"x"}, []int{3}, []float64{-5, 0, 7}, nil, "s")
	if !reflect.DeepEqual(sg.Sign().Data(), []float64{-1, 0, 1}) {
		t.Errorf("Sign = %v", sg.Sign().Data())
	}
}

func TestTrig(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{2}, []float64{0, math.Pi / 2}, nil, "v")
	s := da.Sin().Data()
	if math.Abs(s[0]-0) > 1e-12 || math.Abs(s[1]-1) > 1e-12 {
		t.Errorf("Sin = %v", s)
	}
	c := da.Cos().Data()
	if math.Abs(c[0]-1) > 1e-12 || math.Abs(c[1]-0) > 1e-12 {
		t.Errorf("Cos = %v", c)
	}
}

func TestMaximumMinimum(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{4}, []float64{1, 5, 3, 8},
		map[string][]float64{"x": {0, 1, 2, 3}}, "a")
	b, _ := NewDataArray([]string{"x"}, []int{4}, []float64{4, 2, 6, 1},
		map[string][]float64{"x": {0, 1, 2, 3}}, "b")
	mx, err := a.Maximum(b)
	if err != nil {
		t.Fatalf("Maximum : %v", err)
	}
	if !reflect.DeepEqual(mx.Data(), []float64{4, 5, 6, 8}) {
		t.Errorf("Maximum = %v, attendu [4 5 6 8]", mx.Data())
	}
	mn, _ := a.Minimum(b)
	if !reflect.DeepEqual(mn.Data(), []float64{1, 2, 3, 1}) {
		t.Errorf("Minimum = %v, attendu [1 2 3 1]", mn.Data())
	}
}

func TestMaximumBroadcast(t *testing.T) {
	// broadcasting : x(2) et y(2) -> (2,2), max élément par élément
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 4}, nil, "a")
	b, _ := NewDataArray([]string{"y"}, []int{2}, []float64{3, 2}, nil, "b")
	mx, _ := a.Maximum(b)
	// [[max(1,3),max(1,2)],[max(4,3),max(4,2)]] = [3 2 4 4]
	if !reflect.DeepEqual(mx.Data(), []float64{3, 2, 4, 4}) {
		t.Errorf("Maximum broadcast = %v", mx.Data())
	}
}
