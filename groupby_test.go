package xarray

import (
	"reflect"
	"testing"
)

func TestGroupBy1D(t *testing.T) {
	// Coordonnée « t » avec valeurs répétées : groupes {1:[10 20], 2:[30 40 50]}
	da, _ := NewDataArray([]string{"t"}, []int{5}, []float64{10, 20, 30, 40, 50},
		map[string][]float64{"t": {1, 1, 2, 2, 2}}, "v")

	g, err := da.GroupBy("t")
	if err != nil {
		t.Fatalf("GroupBy : %v", err)
	}
	if g.Groups() != 2 {
		t.Errorf("Groups = %d, attendu 2", g.Groups())
	}
	if !reflect.DeepEqual(g.Labels(), []float64{1, 2}) {
		t.Errorf("Labels = %v", g.Labels())
	}

	s, err := g.Sum()
	if err != nil {
		t.Fatalf("Sum : %v", err)
	}
	// groupe 1 : 10+20=30 ; groupe 2 : 30+40+50=120
	if !reflect.DeepEqual(s.Data(), []float64{30, 120}) {
		t.Errorf("Sum = %v, attendu [30 120]", s.Data())
	}
	if !reflect.DeepEqual(s.Dims(), []string{"t"}) {
		t.Errorf("Dims = %v", s.Dims())
	}
	if c, _ := s.Coord("t"); !reflect.DeepEqual(c, []float64{1, 2}) {
		t.Errorf("Coord = %v, attendu [1 2]", c)
	}

	m, _ := da.GroupBy("t")
	mean, _ := m.Mean()
	// groupe 1 : 15 ; groupe 2 : 40
	if !reflect.DeepEqual(mean.Data(), []float64{15, 40}) {
		t.Errorf("Mean = %v, attendu [15 40]", mean.Data())
	}
}

func TestGroupBy2D(t *testing.T) {
	// dims [t, x], t=[0 0 1] (3), x taille 2.
	// data (3x2) : [[1 2],[3 4],[5 6]]
	da, _ := NewDataArray([]string{"t", "x"}, []int{3, 2},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"t": {0, 0, 1}, "x": {10, 20}}, "v")

	g, _ := da.GroupBy("t")
	s, err := g.Sum()
	if err != nil {
		t.Fatalf("Sum : %v", err)
	}
	// groupe t=0 : lignes 0,1 -> somme sur t -> [1+3, 2+4] = [4 6]
	// groupe t=1 : ligne 2 -> [5 6]
	if !reflect.DeepEqual(s.Dims(), []string{"t", "x"}) {
		t.Errorf("Dims = %v", s.Dims())
	}
	if !reflect.DeepEqual(s.Shape(), []int{2, 2}) {
		t.Errorf("Shape = %v", s.Shape())
	}
	if !reflect.DeepEqual(s.Data(), []float64{4, 6, 5, 6}) {
		t.Errorf("Data = %v, attendu [4 6 5 6]", s.Data())
	}
	// La coordonnée x (non groupée) est conservée.
	if c, _ := s.Coord("x"); !reflect.DeepEqual(c, []float64{10, 20}) {
		t.Errorf("Coord(x) = %v", c)
	}
	// La coordonnée t devient les groupes uniques.
	if c, _ := s.Coord("t"); !reflect.DeepEqual(c, []float64{0, 1}) {
		t.Errorf("Coord(t) = %v", c)
	}
}

func TestGroupByMinMax(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{4}, []float64{5, 1, 9, 3},
		map[string][]float64{"t": {0, 0, 1, 1}}, "v")
	g, _ := da.GroupBy("t")
	mn, _ := g.Min()
	if !reflect.DeepEqual(mn.Data(), []float64{1, 3}) {
		t.Errorf("Min = %v, attendu [1 3]", mn.Data())
	}
	g2, _ := da.GroupBy("t")
	mx, _ := g2.Max()
	if !reflect.DeepEqual(mx.Data(), []float64{5, 9}) {
		t.Errorf("Max = %v, attendu [5 9]", mx.Data())
	}
}

func TestGroupByInt(t *testing.T) {
	// Vérifie le regroupement avec un type entier.
	da, _ := NewDataArray([]string{"t"}, []int{4}, []int32{2, 4, 6, 8},
		map[string][]int32{"t": {1, 2, 1, 2}}, "v")
	g, _ := da.GroupBy("t")
	s, _ := g.Sum()
	// groupe 1 : positions 0,2 -> 2+6=8 ; groupe 2 : positions 1,3 -> 4+8=12
	if !reflect.DeepEqual(s.Data(), []int32{8, 12}) {
		t.Errorf("Sum = %v, attendu [8 12]", s.Data())
	}
}

func TestGroupBySansCoord(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{2}, []float64{1, 2}, nil, "v")
	if _, err := da.GroupBy("t"); err == nil {
		t.Error("erreur attendue : dimension sans coordonnée")
	}
}
