package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestDatasetRolling(t *testing.T) {
	// temperature (t, x) 4×2 ; pluie (t) 4. Fenêtre de 2 sur t.
	temp, _ := NewDataArray([]string{"t", "x"}, []int{4, 2},
		[]float64{1, 10, 2, 20, 3, 30, 4, 40},
		map[string][]float64{"t": {0, 1, 2, 3}, "x": {100, 200}}, "temperature")
	pluie, _ := NewDataArray([]string{"t"}, []int{4}, []float64{5, 7, 9, 11},
		map[string][]float64{"t": {0, 1, 2, 3}}, "pluie")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"temperature": temp, "pluie": pluie})

	r, err := ds.Rolling("t", 2)
	if err != nil {
		t.Fatalf("Rolling : %v", err)
	}
	s, err := r.Sum()
	if err != nil {
		t.Fatalf("Sum : %v", err)
	}

	// pluie somme mobile : [NaN, 12, 16, 20]
	pl, _ := s.Get("pluie")
	pd := pl.Data()
	if !math.IsNaN(pd[0]) || pd[1] != 12 || pd[2] != 16 || pd[3] != 20 {
		t.Errorf("pluie mobile = %v", pd)
	}

	// temperature somme mobile sur t, par colonne x :
	// ligne0 NaN NaN ; ligne1 [1+2,10+20]=[3 30] ; ligne2 [2+3,20+30]=[5 50] ; ligne3 [3+4,30+40]=[7 70]
	tp, _ := s.Get("temperature")
	td := tp.Data()
	if !math.IsNaN(td[0]) || !math.IsNaN(td[1]) {
		t.Errorf("temperature ligne0 devrait être NaN : %v", td)
	}
	if td[2] != 3 || td[3] != 30 || td[4] != 5 || td[5] != 50 || td[6] != 7 || td[7] != 70 {
		t.Errorf("temperature mobile = %v", td)
	}
}

func TestDatasetRollingVarPartielle(t *testing.T) {
	// Variable ne portant pas la dimension roulée : conservée (convertie).
	a, _ := NewDataArray([]string{"t"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"t": {0, 1, 2}}, "a")
	b, _ := NewDataArray([]string{"z"}, []int{2}, []float64{9, 8},
		map[string][]float64{"z": {0, 1}}, "b")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a, "b": b})
	r, _ := ds.Rolling("t", 2)
	m, err := r.Mean()
	if err != nil {
		t.Fatalf("Mean : %v", err)
	}
	// b inchangée.
	mb, _ := m.Get("b")
	if !reflect.DeepEqual(mb.Data(), []float64{9, 8}) {
		t.Errorf("b modifiée : %v", mb.Data())
	}
	// a moyenne mobile : [NaN, 1.5, 2.5]
	ma, _ := m.Get("a")
	if ma.Data()[1] != 1.5 || ma.Data()[2] != 2.5 {
		t.Errorf("a mobile = %v", ma.Data())
	}
}
