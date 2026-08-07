package xarray

import (
	"reflect"
	"testing"
)

func TestLazySumAxis2D(t *testing.T) {
	// [[1 2 3],[4 5 6],[7 8 9],[10 11 12]] dims t,x (4×3)
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	da, _ := NewDataArray([]string{"t", "x"}, []int{4, 3}, data,
		map[string][]float64{"t": {0, 1, 2, 3}, "x": {10, 20, 30}}, "v")
	lz, _ := Chunk(da, 3) // 2 chunks (3 + 1)

	// Somme le long de t (axe 0) -> par colonne : [1+4+7+10, 2+5+8+11, 3+6+9+12]
	s0, err := lz.SumAxis("t")
	if err != nil {
		t.Fatalf("SumAxis(t) : %v", err)
	}
	if !reflect.DeepEqual(s0.Dims(), []string{"x"}) {
		t.Errorf("Dims = %v", s0.Dims())
	}
	if !reflect.DeepEqual(s0.Data(), []float64{22, 26, 30}) {
		t.Errorf("SumAxis(t) = %v, attendu [22 26 30]", s0.Data())
	}
	if c, _ := s0.Coord("x"); !reflect.DeepEqual(c, []float64{10, 20, 30}) {
		t.Errorf("coord x = %v", c)
	}

	// Somme le long de x (axe 1) -> par ligne : [6, 15, 24, 33]
	s1, _ := lz.SumAxis("x")
	if !reflect.DeepEqual(s1.Dims(), []string{"t"}) {
		t.Errorf("Dims = %v", s1.Dims())
	}
	if !reflect.DeepEqual(s1.Data(), []float64{6, 15, 24, 33}) {
		t.Errorf("SumAxis(x) = %v, attendu [6 15 24 33]", s1.Data())
	}
}

func TestLazyMeanMinMaxAxis(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6}
	da, _ := NewDataArray([]string{"t", "x"}, []int{2, 3}, data, nil, "v")
	lz, _ := Chunk(da, 1) // 2 chunks

	m, _ := lz.MeanAxis("t") // colonnes : [(1+4)/2, (2+5)/2, (3+6)/2] = [2.5 3.5 4.5]
	if !reflect.DeepEqual(m.Data(), []float64{2.5, 3.5, 4.5}) {
		t.Errorf("MeanAxis(t) = %v", m.Data())
	}
	mx, _ := lz.MaxAxis("x") // lignes : [3, 6]
	if !reflect.DeepEqual(mx.Data(), []float64{3, 6}) {
		t.Errorf("MaxAxis(x) = %v", mx.Data())
	}
	mn, _ := lz.MinAxis("t") // colonnes : [1, 2, 3]
	if !reflect.DeepEqual(mn.Data(), []float64{1, 2, 3}) {
		t.Errorf("MinAxis(t) = %v", mn.Data())
	}
}

func TestLazyReduceAxis1D(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{5}, []float64{1, 2, 3, 4, 5}, nil, "v")
	lz, _ := Chunk(da, 2)
	s, _ := lz.SumAxis("t") // scalaire : 15
	if s.Ndim() != 0 || s.Data()[0] != 15 {
		t.Errorf("SumAxis 1D = %v (ndim %d)", s.Data(), s.Ndim())
	}
}

func TestLazyReduceAxisCoherence(t *testing.T) {
	// Le résultat lazy doit coïncider avec la réduction directe.
	data := make([]float64, 30)
	for i := range data {
		data[i] = float64(i)
	}
	da, _ := NewDataArray([]string{"t", "x"}, []int{6, 5}, data, nil, "v")
	direct, _ := da.SumAxis("t")
	lz, _ := Chunk(da, 4)
	lazy, _ := lz.SumAxis("t")
	if !reflect.DeepEqual(lazy.Data(), direct.Data()) {
		t.Errorf("lazy %v != direct %v", lazy.Data(), direct.Data())
	}
}
