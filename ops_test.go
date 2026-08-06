package xarray

import (
	"reflect"
	"testing"
)

func TestTranspose(t *testing.T) {
	v, _ := NewVariable([]string{"x", "y"}, []int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	tr, err := v.Transpose("y", "x")
	if err != nil {
		t.Fatalf("Transpose : %v", err)
	}
	if !reflect.DeepEqual(tr.Shape(), []int{3, 2}) {
		t.Errorf("Shape = %v, attendu [3 2]", tr.Shape())
	}
	// L'original [[1 2 3],[4 5 6]] transposé -> [[1 4],[2 5],[3 6]]
	if !reflect.DeepEqual(tr.Data(), []float64{1, 4, 2, 5, 3, 6}) {
		t.Errorf("Data = %v", tr.Data())
	}
	if _, err := v.Transpose("y", "z"); err == nil {
		t.Error("erreur attendue : dimension inconnue")
	}
}

func TestReduceAxis(t *testing.T) {
	// [[1 2 3],[4 5 6]] dims x,y
	da, _ := NewDataArray([]string{"x", "y"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"x": {0, 1}, "y": {10, 20, 30}}, "a")

	// Somme le long de x -> dim y : [1+4, 2+5, 3+6] = [5 7 9]
	sx, err := da.SumAxis("x")
	if err != nil {
		t.Fatalf("SumAxis : %v", err)
	}
	if !reflect.DeepEqual(sx.Dims(), []string{"y"}) {
		t.Errorf("Dims = %v", sx.Dims())
	}
	if !reflect.DeepEqual(sx.Data(), []float64{5, 7, 9}) {
		t.Errorf("Data = %v, attendu [5 7 9]", sx.Data())
	}
	// La coordonnée x doit disparaître, y subsister.
	if _, err := sx.Coord("y"); err != nil {
		t.Error("coordonnée y attendue")
	}

	// Moyenne le long de y -> dim x : [(1+2+3)/3, (4+5+6)/3] = [2 5]
	my, _ := da.MeanAxis("y")
	if !reflect.DeepEqual(my.Data(), []float64{2, 5}) {
		t.Errorf("MeanAxis(y) = %v, attendu [2 5]", my.Data())
	}
}

func TestArithmetiqueBroadcast(t *testing.T) {
	// a : dim x (2) = [1 2] ; b : dim y (3) = [10 20 30]
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2}, nil, "a")
	b, _ := NewDataArray([]string{"y"}, []int{3}, []float64{10, 20, 30}, nil, "b")

	// Broadcasting par nom -> dims [x y], shape [2 3]
	// [[1+10,1+20,1+30],[2+10,2+20,2+30]] = [11 21 31 12 22 32]
	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add : %v", err)
	}
	if !reflect.DeepEqual(sum.Dims(), []string{"x", "y"}) {
		t.Errorf("Dims = %v", sum.Dims())
	}
	if !reflect.DeepEqual(sum.Data(), []float64{11, 21, 31, 12, 22, 32}) {
		t.Errorf("Data = %v", sum.Data())
	}
}

func TestArithmetiqueMemeDim(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3}, nil, "a")
	b, _ := NewDataArray([]string{"x"}, []int{3}, []float64{10, 10, 10}, nil, "b")
	m, _ := a.Mul(b)
	if !reflect.DeepEqual(m.Data(), []float64{10, 20, 30}) {
		t.Errorf("Mul = %v", m.Data())
	}
	s, _ := b.Sub(a)
	if !reflect.DeepEqual(s.Data(), []float64{9, 8, 7}) {
		t.Errorf("Sub = %v", s.Data())
	}
}

func TestArithmetiqueTaillesIncompatibles(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2}, nil, "a")
	b, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3}, nil, "b")
	if _, err := a.Add(b); err == nil {
		t.Error("erreur attendue : dimension x de tailles différentes sans coordonnées")
	}
}

func TestAlignement(t *testing.T) {
	// a sur x = [0 1 2], b sur x = [1 2 3] -> intersection [1 2]
	a, _ := NewDataArray([]string{"x"}, []int{3}, []float64{10, 20, 30},
		map[string][]float64{"x": {0, 1, 2}}, "a")
	b, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"x": {1, 2, 3}}, "b")

	// a aligné : positions x=1,2 -> [20 30] ; b aligné : x=1,2 -> [1 2]
	// somme -> [21 32]
	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add aligné : %v", err)
	}
	if !reflect.DeepEqual(sum.Data(), []float64{21, 32}) {
		t.Errorf("Data = %v, attendu [21 32]", sum.Data())
	}
	labels, _ := sum.Coord("x")
	if !reflect.DeepEqual(labels, []float64{1, 2}) {
		t.Errorf("coord x = %v, attendu [1 2]", labels)
	}
}

func TestAlignementSansCommun(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2},
		map[string][]float64{"x": {0, 1}}, "a")
	b, _ := NewDataArray([]string{"x"}, []int{2}, []float64{3, 4},
		map[string][]float64{"x": {5, 6}}, "b")
	if _, err := a.Add(b); err == nil {
		t.Error("erreur attendue : aucune étiquette commune")
	}
}

func TestScalaires(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"x": {0, 1, 2}}, "a")
	r := a.AddScalar(10).MulScalar(2)
	if !reflect.DeepEqual(r.Data(), []float64{22, 24, 26}) {
		t.Errorf("Data = %v, attendu [22 24 26]", r.Data())
	}
	// Coordonnées conservées.
	if labels, _ := r.Coord("x"); !reflect.DeepEqual(labels, []float64{0, 1, 2}) {
		t.Errorf("coord x = %v", labels)
	}
}

func TestReduceAxis3D(t *testing.T) {
	// forme 2x2x2, données 0..7
	da, _ := NewDataArray([]string{"a", "b", "c"}, []int{2, 2, 2},
		[]float64{0, 1, 2, 3, 4, 5, 6, 7}, nil, "cube")
	// Somme le long de b -> dims [a c]
	// a=0: b=0->(0,1) b=1->(2,3) => c0=0+2=2, c1=1+3=4
	// a=1: b=0->(4,5) b=1->(6,7) => c0=4+6=10, c1=5+7=12
	s, err := da.SumAxis("b")
	if err != nil {
		t.Fatalf("SumAxis : %v", err)
	}
	if !reflect.DeepEqual(s.Dims(), []string{"a", "c"}) {
		t.Errorf("Dims = %v", s.Dims())
	}
	if !reflect.DeepEqual(s.Data(), []float64{2, 4, 10, 12}) {
		t.Errorf("Data = %v, attendu [2 4 10 12]", s.Data())
	}
}
