package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestDatasetStatsAxis(t *testing.T) {
	// temperature (t,lieu) 2×3 ; pluie (t) 2
	temp, _ := NewDataArray([]string{"t", "lieu"}, []int{2, 3},
		[]float64{2, 4, 6, 4, 8, 12},
		map[string][]float64{"t": {0, 1}, "lieu": {10, 20, 30}}, "temperature")
	pluie, _ := NewDataArray([]string{"t"}, []int{2}, []float64{2, 6},
		map[string][]float64{"t": {0, 1}}, "pluie")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"temperature": temp, "pluie": pluie})

	// Écart-type le long de t.
	s, err := ds.StdAxis("t")
	if err != nil {
		t.Fatalf("StdAxis : %v", err)
	}
	// temperature par lieu : std([2,4])=1, std([4,8])=2, std([6,12])=3
	tp, _ := s.Get("temperature")
	if !reflect.DeepEqual(tp.Data(), []float64{1, 2, 3}) {
		t.Errorf("StdAxis temperature = %v, attendu [1 2 3]", tp.Data())
	}
	// pluie : std([2,6]) = 2
	pl, _ := s.Get("pluie")
	if pl.Data()[0] != 2 {
		t.Errorf("StdAxis pluie = %v, attendu [2]", pl.Data())
	}
}

func TestDatasetFillNA(t *testing.T) {
	nan := math.NaN()
	a, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, nan, 3}, nil, "a")
	b, _ := NewDataArray([]string{"x"}, []int{3}, []float64{nan, 5, 6}, nil, "b")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a, "b": b})

	f, err := ds.FillNA(0)
	if err != nil {
		t.Fatalf("FillNA : %v", err)
	}
	fa, _ := f.Get("a")
	fb, _ := f.Get("b")
	if !reflect.DeepEqual(fa.Data(), []float64{1, 0, 3}) {
		t.Errorf("FillNA a = %v", fa.Data())
	}
	if !reflect.DeepEqual(fb.Data(), []float64{0, 5, 6}) {
		t.Errorf("FillNA b = %v", fb.Data())
	}
}

func TestDatasetCumsum(t *testing.T) {
	a, _ := NewDataArray([]string{"t"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"t": {0, 1, 2}}, "a")
	// b n'a pas la dimension t -> conservée telle quelle.
	b, _ := NewDataArray([]string{"z"}, []int{2}, []float64{10, 20},
		map[string][]float64{"z": {0, 1}}, "b")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a, "b": b})

	c, err := ds.Cumsum("t")
	if err != nil {
		t.Fatalf("Cumsum : %v", err)
	}
	ca, _ := c.Get("a")
	if !reflect.DeepEqual(ca.Data(), []float64{1, 3, 6}) {
		t.Errorf("Cumsum a = %v, attendu [1 3 6]", ca.Data())
	}
	cb, _ := c.Get("b")
	if !reflect.DeepEqual(cb.Data(), []float64{10, 20}) {
		t.Errorf("b modifiée : %v", cb.Data())
	}
}
