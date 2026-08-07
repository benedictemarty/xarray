package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestRollingMean1D(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{5}, []float64{1, 2, 3, 4, 5},
		map[string][]float64{"t": {0, 1, 2, 3, 4}}, "v")
	r, err := da.Rolling("t", 3)
	if err != nil {
		t.Fatalf("Rolling : %v", err)
	}
	m, err := r.Mean()
	if err != nil {
		t.Fatalf("Mean : %v", err)
	}
	got := m.Data()
	// Positions 0,1 : NaN ; puis moyennes de [1,2,3]=2, [2,3,4]=3, [3,4,5]=4.
	if !math.IsNaN(got[0]) || !math.IsNaN(got[1]) {
		t.Errorf("les 2 premières valeurs devraient être NaN : %v", got)
	}
	if got[2] != 2 || got[3] != 3 || got[4] != 4 {
		t.Errorf("moyennes mobiles = %v, attendu [_ _ 2 3 4]", got)
	}
	// Même forme que l'entrée.
	if !reflect.DeepEqual(m.Shape(), []int{5}) {
		t.Errorf("Shape = %v", m.Shape())
	}
}

func TestRollingSumMinMax(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{4}, []float64{5, 1, 9, 3}, nil, "v")
	r, _ := da.Rolling("t", 2)
	s, _ := r.Sum()
	// [_ 6 10 12]
	if s.Data()[1] != 6 || s.Data()[2] != 10 || s.Data()[3] != 12 {
		t.Errorf("Sum mobile = %v", s.Data())
	}
	r2, _ := da.Rolling("t", 2)
	mn, _ := r2.Min()
	if mn.Data()[1] != 1 || mn.Data()[2] != 1 || mn.Data()[3] != 3 {
		t.Errorf("Min mobile = %v", mn.Data())
	}
}

func TestRolling2D(t *testing.T) {
	// dims [t, x], fenêtre sur t. Chaque colonne x est lissée indépendamment.
	da, _ := NewDataArray([]string{"t", "x"}, []int{3, 2},
		[]float64{1, 10, 2, 20, 3, 30}, nil, "v")
	r, _ := da.Rolling("t", 2)
	s, _ := r.Sum()
	got := s.Data()
	// ligne 0 : NaN NaN ; ligne 1 : [1+2, 10+20]=[3 30] ; ligne 2 : [2+3, 20+30]=[5 50]
	if !math.IsNaN(got[0]) || !math.IsNaN(got[1]) {
		t.Errorf("première ligne devrait être NaN : %v", got)
	}
	if got[2] != 3 || got[3] != 30 || got[4] != 5 || got[5] != 50 {
		t.Errorf("Sum mobile 2D = %v", got)
	}
}

func TestRollingErreurs(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{3}, []float64{1, 2, 3}, nil, "v")
	if _, err := da.Rolling("z", 2); err == nil {
		t.Error("erreur attendue : dimension inconnue")
	}
	if _, err := da.Rolling("t", 0); err == nil {
		t.Error("erreur attendue : fenêtre nulle")
	}
}
