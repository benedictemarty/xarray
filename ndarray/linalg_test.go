package ndarray

import (
	"reflect"
	"testing"
)

func TestMatmul(t *testing.T) {
	// [[1 2 3],[4 5 6]] (2×3) · [[7 8],[9 10],[11 12]] (3×2) = [[58 64],[139 154]]
	a, _ := New([]int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	b, _ := New([]int{3, 2}, []float64{7, 8, 9, 10, 11, 12})
	c, err := Matmul(a, b)
	if err != nil {
		t.Fatalf("Matmul : %v", err)
	}
	if !reflect.DeepEqual(c.Shape(), []int{2, 2}) {
		t.Errorf("Shape = %v", c.Shape())
	}
	if !reflect.DeepEqual(c.Data(), []float64{58, 64, 139, 154}) {
		t.Errorf("Matmul = %v, attendu [58 64 139 154]", c.Data())
	}
}

func TestMatmulIdentite(t *testing.T) {
	a, _ := New([]int{2, 2}, []float64{1, 2, 3, 4})
	id, _ := New([]int{2, 2}, []float64{1, 0, 0, 1})
	c, _ := Matmul(a, id)
	if !reflect.DeepEqual(c.Data(), a.Data()) {
		t.Errorf("a·I = %v, attendu %v", c.Data(), a.Data())
	}
}

func TestMatmulIncompatible(t *testing.T) {
	a, _ := New([]int{2, 3}, make([]float64, 6))
	b, _ := New([]int{2, 2}, make([]float64, 4))
	if _, err := Matmul(a, b); err == nil {
		t.Error("erreur attendue : dimensions internes incompatibles")
	}
}

func TestTranspose(t *testing.T) {
	a, _ := New([]int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	at, err := a.T()
	if err != nil {
		t.Fatalf("T : %v", err)
	}
	if !reflect.DeepEqual(at.Shape(), []int{3, 2}) {
		t.Errorf("Shape = %v", at.Shape())
	}
	// [[1 2 3],[4 5 6]]^T = [[1 4],[2 5],[3 6]]
	if !reflect.DeepEqual(at.Data(), []float64{1, 4, 2, 5, 3, 6}) {
		t.Errorf("T = %v", at.Data())
	}
}

func TestMatVec(t *testing.T) {
	// [[1 2],[3 4]] · [5 6] = [1*5+2*6, 3*5+4*6] = [17 39]
	a, _ := New([]int{2, 2}, []float64{1, 2, 3, 4})
	v, _ := New([]int{2}, []float64{5, 6})
	r, err := MatVec(a, v)
	if err != nil {
		t.Fatalf("MatVec : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{17, 39}) {
		t.Errorf("MatVec = %v, attendu [17 39]", r.Data())
	}
}

func BenchmarkMatmul256(b *testing.B) {
	n := 256
	x := grille(n)
	y := grille(n)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Matmul(x, y); err != nil {
			b.Fatal(err)
		}
	}
}
