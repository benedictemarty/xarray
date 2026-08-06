package ndarray

import (
	"reflect"
	"testing"
)

func TestNewEtAccès(t *testing.T) {
	a, err := New([]int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	if err != nil {
		t.Fatalf("New : %v", err)
	}
	if a.Ndim() != 2 || a.Size() != 6 {
		t.Errorf("Ndim/Size = %d/%d", a.Ndim(), a.Size())
	}
	if v, _ := a.At(1, 2); v != 6 {
		t.Errorf("At(1,2) = %v", v)
	}
	if _, err := New([]int{2, 2}, []float64{1}); err == nil {
		t.Error("erreur attendue : taille incohérente")
	}
}

func TestAddSameShape(t *testing.T) {
	a, _ := New([]int{2, 2}, []float64{1, 2, 3, 4})
	b, _ := New([]int{2, 2}, []float64{10, 20, 30, 40})
	c, _ := a.Add(b)
	if !reflect.DeepEqual(c.Data(), []float64{11, 22, 33, 44}) {
		t.Errorf("Add = %v", c.Data())
	}
	d, _ := a.Mul(b)
	if !reflect.DeepEqual(d.Data(), []float64{10, 40, 90, 160}) {
		t.Errorf("Mul = %v", d.Data())
	}
}

func TestBroadcastPositionnel(t *testing.T) {
	// (2,3) + (3,) -> broadcasting aligné à droite (la ligne est diffusée).
	a, _ := New([]int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	b, _ := New([]int{3}, []float64{10, 20, 30})
	c, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add broadcast : %v", err)
	}
	if !reflect.DeepEqual(c.Shape(), []int{2, 3}) {
		t.Errorf("Shape = %v", c.Shape())
	}
	// [[1+10,2+20,3+30],[4+10,5+20,6+30]]
	if !reflect.DeepEqual(c.Data(), []float64{11, 22, 33, 14, 25, 36}) {
		t.Errorf("Data = %v", c.Data())
	}
}

func TestBroadcastColonne(t *testing.T) {
	// (2,1) + (1,3) -> (2,3), les deux diffusés.
	a, _ := New([]int{2, 1}, []float64{1, 2})
	b, _ := New([]int{1, 3}, []float64{10, 20, 30})
	c, _ := a.Add(b)
	if !reflect.DeepEqual(c.Shape(), []int{2, 3}) {
		t.Errorf("Shape = %v", c.Shape())
	}
	// [[11,21,31],[12,22,32]]
	if !reflect.DeepEqual(c.Data(), []float64{11, 21, 31, 12, 22, 32}) {
		t.Errorf("Data = %v", c.Data())
	}
}

func TestBroadcastIncompatible(t *testing.T) {
	a, _ := New([]int{2, 3}, make([]float64, 6))
	b, _ := New([]int{2, 2}, make([]float64, 4))
	if _, err := a.Add(b); err == nil {
		t.Error("erreur attendue : formes non diffusables")
	}
}

func TestReductions(t *testing.T) {
	// [[1 2 3],[4 5 6]]
	a, _ := New([]int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	if a.Sum() != 21 {
		t.Errorf("Sum = %v", a.Sum())
	}
	if a.Mean() != 3.5 {
		t.Errorf("Mean = %v", a.Mean())
	}
	// Somme axe 0 -> [1+4,2+5,3+6] = [5 7 9]
	s0, _ := a.SumAxis(0)
	if !reflect.DeepEqual(s0.Data(), []float64{5, 7, 9}) {
		t.Errorf("SumAxis(0) = %v", s0.Data())
	}
	// Moyenne axe 1 -> [(1+2+3)/3,(4+5+6)/3] = [2 5]
	m1, _ := a.MeanAxis(1)
	if !reflect.DeepEqual(m1.Data(), []float64{2, 5}) {
		t.Errorf("MeanAxis(1) = %v", m1.Data())
	}
}

func TestScalaires(t *testing.T) {
	a := Arange(4) // [0 1 2 3]
	r := a.AddScalar(10).MulScalar(2)
	if !reflect.DeepEqual(r.Data(), []float64{20, 22, 24, 26}) {
		t.Errorf("scalaires = %v", r.Data())
	}
}
