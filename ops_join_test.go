package xarray

import (
	"reflect"
	"testing"
)

func deuxTableauxDecales(t *testing.T) (*DataArray[float64], *DataArray[float64]) {
	t.Helper()
	// a sur x=[0 1 2] = [10 20 30] ; b sur x=[1 2 3] = [1 2 3]
	a, _ := NewDataArray([]string{"x"}, []int{3}, []float64{10, 20, 30},
		map[string][]float64{"x": {0, 1, 2}}, "a")
	b, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"x": {1, 2, 3}}, "b")
	return a, b
}

func TestJoinInner(t *testing.T) {
	a, b := deuxTableauxDecales(t)
	// Intersection [1 2] : a=[20 30], b=[1 2] -> [21 32]
	r, err := a.AddJoin(b, JoinInner, 0)
	if err != nil {
		t.Fatalf("AddJoin inner : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{21, 32}) {
		t.Errorf("inner = %v, attendu [21 32]", r.Data())
	}
	if c, _ := r.Coord("x"); !reflect.DeepEqual(c, []float64{1, 2}) {
		t.Errorf("coord = %v, attendu [1 2]", c)
	}
}

func TestJoinOuter(t *testing.T) {
	a, b := deuxTableauxDecales(t)
	// Union [0 1 2 3], fill=0 :
	// a -> [10 20 30 0], b -> [0 1 2 3] -> somme [10 21 32 3]
	r, err := a.AddJoin(b, JoinOuter, 0)
	if err != nil {
		t.Fatalf("AddJoin outer : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{10, 21, 32, 3}) {
		t.Errorf("outer = %v, attendu [10 21 32 3]", r.Data())
	}
	if c, _ := r.Coord("x"); !reflect.DeepEqual(c, []float64{0, 1, 2, 3}) {
		t.Errorf("coord = %v, attendu [0 1 2 3]", c)
	}
}

func TestJoinLeft(t *testing.T) {
	a, b := deuxTableauxDecales(t)
	// Étiquettes de a [0 1 2], fill=0 : a=[10 20 30], b=[0 1 2] -> [10 21 32]
	r, err := a.AddJoin(b, JoinLeft, 0)
	if err != nil {
		t.Fatalf("AddJoin left : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{10, 21, 32}) {
		t.Errorf("left = %v, attendu [10 21 32]", r.Data())
	}
}

func TestJoinRight(t *testing.T) {
	a, b := deuxTableauxDecales(t)
	// Étiquettes de b [1 2 3], fill=0 : a=[20 30 0], b=[1 2 3] -> [21 32 3]
	r, err := a.AddJoin(b, JoinRight, 0)
	if err != nil {
		t.Fatalf("AddJoin right : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{21, 32, 3}) {
		t.Errorf("right = %v, attendu [21 32 3]", r.Data())
	}
	if c, _ := r.Coord("x"); !reflect.DeepEqual(c, []float64{1, 2, 3}) {
		t.Errorf("coord = %v, attendu [1 2 3]", c)
	}
}

func TestJoinOuterFillPersonnalise(t *testing.T) {
	a, b := deuxTableauxDecales(t)
	// fill = 100 : a -> [10 20 30 100], b -> [100 1 2 3] -> [110 21 32 103]
	r, _ := a.AddJoin(b, JoinOuter, 100)
	if !reflect.DeepEqual(r.Data(), []float64{110, 21, 32, 103}) {
		t.Errorf("outer fill=100 = %v", r.Data())
	}
}

func TestJoinType2D(t *testing.T) {
	// Jointure sur une dimension d'un tableau 2D : l'autre dimension est diffusée.
	a, _ := NewDataArray([]string{"x", "y"}, []int{2, 2},
		[]float64{1, 2, 3, 4},
		map[string][]float64{"x": {0, 1}, "y": {0, 1}}, "a")
	b, _ := NewDataArray([]string{"x", "y"}, []int{2, 2},
		[]float64{10, 20, 30, 40},
		map[string][]float64{"x": {1, 2}, "y": {0, 1}}, "b")
	// Outer sur x -> [0 1 2], inner-identique sur y [0 1].
	// a réindexé sur x : ligne0=[1 2], ligne1=[3 4], ligne2=[0 0]
	// b réindexé sur x : ligne0=[0 0], ligne1=[10 20], ligne2=[30 40]
	// somme : [1 2 13 24 30 40]
	r, err := a.AddJoin(b, JoinOuter, 0)
	if err != nil {
		t.Fatalf("AddJoin 2D : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{1, 2, 13, 24, 30, 40}) {
		t.Errorf("2D outer = %v", r.Data())
	}
}

func TestJoinTypeString(t *testing.T) {
	if JoinOuter.String() != "outer" || JoinInner.String() != "inner" {
		t.Errorf("String de JoinType incorrect")
	}
}
