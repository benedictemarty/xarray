package xarray

import (
	"reflect"
	"testing"
)

func TestResampleMean(t *testing.T) {
	// Coordonnée t = [0 1 2 3 4 5], pas 2 -> bins [0,2)->{0,1}, [2,4)->{2,3}, [4,6)->{4,5}
	da, _ := NewDataArray([]string{"t"}, []int{6}, []float64{10, 20, 30, 40, 50, 60},
		map[string][]float64{"t": {0, 1, 2, 3, 4, 5}}, "v")
	r, err := da.Resample("t", 2)
	if err != nil {
		t.Fatalf("Resample : %v", err)
	}
	if r.Groups() != 3 {
		t.Errorf("Groups = %d, attendu 3", r.Groups())
	}
	m, err := r.Mean()
	if err != nil {
		t.Fatalf("Mean : %v", err)
	}
	// moyennes : (10+20)/2=15, (30+40)/2=35, (50+60)/2=55
	if !reflect.DeepEqual(m.Data(), []float64{15, 35, 55}) {
		t.Errorf("Mean = %v, attendu [15 35 55]", m.Data())
	}
	// nouvelle coordonnée = bornes gauches des bins
	if c, _ := m.Coord("t"); !reflect.DeepEqual(c, []float64{0, 2, 4}) {
		t.Errorf("coord = %v, attendu [0 2 4]", c)
	}
}

func TestResampleSum2D(t *testing.T) {
	// dims [t, x], resample sur t. t=[0 1 2 3], pas 2 -> 2 bins {0,1},{2,3}
	da, _ := NewDataArray([]string{"t", "x"}, []int{4, 2},
		[]float64{1, 2, 3, 4, 5, 6, 7, 8},
		map[string][]float64{"t": {0, 1, 2, 3}, "x": {10, 20}}, "v")
	r, _ := da.Resample("t", 2)
	s, _ := r.Sum()
	if !reflect.DeepEqual(s.Shape(), []int{2, 2}) {
		t.Errorf("Shape = %v", s.Shape())
	}
	// bin0 (t=0,1) : [1+3, 2+4]=[4 6] ; bin1 (t=2,3) : [5+7, 6+8]=[12 14]
	if !reflect.DeepEqual(s.Data(), []float64{4, 6, 12, 14}) {
		t.Errorf("Sum = %v, attendu [4 6 12 14]", s.Data())
	}
	if c, _ := s.Coord("x"); !reflect.DeepEqual(c, []float64{10, 20}) {
		t.Errorf("coord x = %v", c)
	}
}

func TestResampleBinsIrreguliers(t *testing.T) {
	// Étiquettes non uniformément réparties : t=[0 1 5 6], pas 2
	// bins : 0->{0,1}, 2->{}, 4->{5? floor((5-0)/2)=2 -> bin 2 (borne 4)}, ...
	// floor(0/2)=0, floor(1/2)=0, floor(5/2)=2, floor(6/2)=3
	da, _ := NewDataArray([]string{"t"}, []int{4}, []float64{1, 1, 1, 1},
		map[string][]float64{"t": {0, 1, 5, 6}}, "v")
	r, _ := da.Resample("t", 2)
	s, _ := r.Sum()
	// bins non vides : 0 (borne 0, {0,1}->2), 2 (borne 4, {5}->1), 3 (borne 6, {6}->1)
	if !reflect.DeepEqual(s.Data(), []float64{2, 1, 1}) {
		t.Errorf("Sum = %v, attendu [2 1 1]", s.Data())
	}
	if c, _ := s.Coord("t"); !reflect.DeepEqual(c, []float64{0, 4, 6}) {
		t.Errorf("coord = %v, attendu [0 4 6]", c)
	}
}

func TestResampleErreurs(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{2}, []float64{1, 2}, nil, "v")
	if _, err := da.Resample("t", 2); err == nil {
		t.Error("erreur attendue : pas de coordonnée")
	}
	da2, _ := NewDataArray([]string{"t"}, []int{2}, []float64{1, 2},
		map[string][]float64{"t": {0, 1}}, "v")
	if _, err := da2.Resample("t", 0); err == nil {
		t.Error("erreur attendue : pas nul")
	}
}
