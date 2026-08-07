package xarray

import (
	"reflect"
	"testing"
)

func TestDotMatmul(t *testing.T) {
	// a(x=2, k=3) · b(k=3, y=2) = c(x=2, y=2), équivalent d'un produit matriciel.
	a, _ := NewDataArray([]string{"x", "k"}, []int{2, 3}, []float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"x": {0, 1}}, "a")
	b, _ := NewDataArray([]string{"k", "y"}, []int{3, 2}, []float64{7, 8, 9, 10, 11, 12},
		map[string][]float64{"y": {100, 200}}, "b")
	c, err := Dot(a, b, "k")
	if err != nil {
		t.Fatalf("Dot : %v", err)
	}
	if !reflect.DeepEqual(c.Dims(), []string{"x", "y"}) {
		t.Errorf("Dims = %v, attendu [x y]", c.Dims())
	}
	// [[58 64],[139 154]]
	if !reflect.DeepEqual(c.Data(), []float64{58, 64, 139, 154}) {
		t.Errorf("Dot = %v, attendu [58 64 139 154]", c.Data())
	}
	// Coordonnées restantes conservées.
	if cx, _ := c.Coord("x"); !reflect.DeepEqual(cx, []float64{0, 1}) {
		t.Errorf("coord x = %v", cx)
	}
	if cy, _ := c.Coord("y"); !reflect.DeepEqual(cy, []float64{100, 200}) {
		t.Errorf("coord y = %v", cy)
	}
}

func TestDotProduitScalaire(t *testing.T) {
	// a(k) · b(k) -> scalaire (0 dimension).
	a, _ := NewDataArray([]string{"k"}, []int{3}, []float64{1, 2, 3}, nil, "a")
	b, _ := NewDataArray([]string{"k"}, []int{3}, []float64{4, 5, 6}, nil, "b")
	c, err := Dot(a, b, "k")
	if err != nil {
		t.Fatalf("Dot : %v", err)
	}
	// 1*4 + 2*5 + 3*6 = 32
	if c.Ndim() != 0 || c.Data()[0] != 32 {
		t.Errorf("Dot scalaire = %v (ndim %d), attendu 32", c.Data(), c.Ndim())
	}
}

func TestDotMatVec(t *testing.T) {
	// a(x=2, k=2) · v(k=2) -> (x=2).
	a, _ := NewDataArray([]string{"x", "k"}, []int{2, 2}, []float64{1, 2, 3, 4}, nil, "a")
	v, _ := NewDataArray([]string{"k"}, []int{2}, []float64{5, 6}, nil, "v")
	c, _ := Dot(a, v, "k")
	// [1*5+2*6, 3*5+4*6] = [17 39]
	if !reflect.DeepEqual(c.Data(), []float64{17, 39}) {
		t.Errorf("Dot matvec = %v, attendu [17 39]", c.Data())
	}
}

func TestDotErreurs(t *testing.T) {
	a, _ := NewDataArray([]string{"x", "k"}, []int{2, 3}, make([]float64, 6), nil, "a")
	b, _ := NewDataArray([]string{"k", "y"}, []int{2, 2}, make([]float64, 4), nil, "b")
	// k a des tailles différentes (3 vs 2).
	if _, err := Dot(a, b, "k"); err == nil {
		t.Error("erreur attendue : tailles de k incompatibles")
	}
	// dimension de contraction absente.
	if _, err := Dot(a, b, "z"); err == nil {
		t.Error("erreur attendue : dimension absente")
	}
}
