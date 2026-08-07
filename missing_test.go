package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestFillNAetCount(t *testing.T) {
	nan := math.NaN()
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{1, nan, 3, nan, 5}, nil, "v")
	if da.CountNA() != 2 {
		t.Errorf("CountNA = %d, attendu 2", da.CountNA())
	}
	f := da.FillNA(0)
	if !reflect.DeepEqual(f.Data(), []float64{1, 0, 3, 0, 5}) {
		t.Errorf("FillNA = %v", f.Data())
	}
	// L'original ne doit pas être modifié.
	if da.CountNA() != 2 {
		t.Errorf("original modifié")
	}
}

func TestDropNA1D(t *testing.T) {
	nan := math.NaN()
	da, _ := NewDataArray([]string{"x"}, []int{5}, []float64{1, nan, 3, nan, 5},
		map[string][]float64{"x": {0, 1, 2, 3, 4}}, "v")
	d, err := da.DropNA("x")
	if err != nil {
		t.Fatalf("DropNA : %v", err)
	}
	if !reflect.DeepEqual(d.Data(), []float64{1, 3, 5}) {
		t.Errorf("DropNA = %v, attendu [1 3 5]", d.Data())
	}
	// Coordonnée réduite en conséquence.
	if c, _ := d.Coord("x"); !reflect.DeepEqual(c, []float64{0, 2, 4}) {
		t.Errorf("coord = %v, attendu [0 2 4]", c)
	}
}

func TestDropNA2D(t *testing.T) {
	nan := math.NaN()
	// dims [t, x] 3×2 ; la ligne t=1 contient un NaN -> supprimée.
	da, _ := NewDataArray([]string{"t", "x"}, []int{3, 2},
		[]float64{1, 2, 3, nan, 5, 6},
		map[string][]float64{"t": {0, 1, 2}, "x": {10, 20}}, "v")
	d, err := da.DropNA("t")
	if err != nil {
		t.Fatalf("DropNA : %v", err)
	}
	if !reflect.DeepEqual(d.Shape(), []int{2, 2}) {
		t.Errorf("Shape = %v, attendu [2 2]", d.Shape())
	}
	// Restent t=0 [1 2] et t=2 [5 6].
	if !reflect.DeepEqual(d.Data(), []float64{1, 2, 5, 6}) {
		t.Errorf("DropNA 2D = %v", d.Data())
	}
	if c, _ := d.Coord("t"); !reflect.DeepEqual(c, []float64{0, 2}) {
		t.Errorf("coord t = %v", c)
	}
}

func TestFFillBFill(t *testing.T) {
	nan := math.NaN()
	da, _ := NewDataArray([]string{"x"}, []int{6}, []float64{nan, 1, nan, nan, 4, nan}, nil, "v")

	f, _ := da.FFill("x")
	// [NaN 1 1 1 4 4]
	got := f.Data()
	if !math.IsNaN(got[0]) || got[1] != 1 || got[2] != 1 || got[3] != 1 || got[4] != 4 || got[5] != 4 {
		t.Errorf("FFill = %v", got)
	}

	b, _ := da.BFill("x")
	// [1 1 4 4 4 NaN]
	gb := b.Data()
	if gb[0] != 1 || gb[1] != 1 || gb[2] != 4 || gb[3] != 4 || gb[4] != 4 || !math.IsNaN(gb[5]) {
		t.Errorf("BFill = %v", gb)
	}
}

func TestMissingEntier(t *testing.T) {
	// Type entier : aucun NaN possible, opérations sans effet.
	da, _ := NewDataArray([]string{"x"}, []int{3}, []int{1, 2, 3}, nil, "v")
	if da.CountNA() != 0 {
		t.Errorf("CountNA[int] = %d", da.CountNA())
	}
	if !reflect.DeepEqual(da.FillNA(9).Data(), []int{1, 2, 3}) {
		t.Errorf("FillNA[int] a modifié les données")
	}
}
