package xarray

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

// TestEquivalenceAvecXarray vérifie que xarray-go produit exactement les mêmes
// valeurs que xarray (Python) pour un jeu d'opérations. Le fichier de référence
// bench/expected.json est produit par bench/crosscheck.py.
//
// Le test est ignoré si le fichier de référence est absent (xarray non exécuté).
func TestEquivalenceAvecXarray(t *testing.T) {
	const ref = "bench/expected.json"
	raw, err := os.ReadFile(ref)
	if err != nil {
		t.Skipf("référence %s absente (lancer d'abord: python3 bench/crosscheck.py) : %v", ref, err)
	}
	var expected map[string][]float64
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("lecture de %s : %v", ref, err)
	}

	got := map[string][]float64{}

	// add
	a, _ := NewDataArray([]string{"x", "y"}, []int{2, 3}, []float64{0, 1, 2, 3, 4, 5},
		map[string][]float64{"x": {0, 1}, "y": {10, 20, 30}}, "")
	b, _ := NewDataArray([]string{"x", "y"}, []int{2, 3}, []float64{6, 7, 8, 9, 10, 11},
		map[string][]float64{"x": {0, 1}, "y": {10, 20, 30}}, "")
	radd, _ := a.Add(b)
	got["add"] = radd.Data()

	// broadcast
	x, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3}, nil, "")
	y, _ := NewDataArray([]string{"y"}, []int{2}, []float64{10, 20}, nil, "")
	rb, _ := x.Add(y)
	got["broadcast"] = rb.Data()

	// sum_axis_x / mean_axis_x
	sx, _ := a.SumAxis("x")
	got["sum_axis_x"] = sx.Data()
	mx, _ := a.MeanAxis("x")
	got["mean_axis_x"] = mx.Data()

	// outer_join
	a2, _ := NewDataArray([]string{"k"}, []int{3}, []float64{10, 20, 30},
		map[string][]float64{"k": {0, 1, 2}}, "")
	b2, _ := NewDataArray([]string{"k"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"k": {1, 2, 3}}, "")
	oj, _ := a2.AddJoin(b2, JoinOuter, 0)
	got["outer_join"] = oj.Data()

	// groupby_sum
	g, _ := NewDataArray([]string{"t"}, []int{5}, []float64{10, 20, 30, 40, 50},
		map[string][]float64{"t": {1, 1, 2, 2, 2}}, "")
	gb, _ := g.GroupBy("t")
	gs, _ := gb.Sum()
	got["groupby_sum"] = gs.Data()

	const tol = 1e-9
	for op, exp := range expected {
		g, ok := got[op]
		if !ok {
			t.Errorf("opération %q calculée par Python mais absente côté Go", op)
			continue
		}
		if len(g) != len(exp) {
			t.Errorf("%s : longueur %d (Go) vs %d (Python)", op, len(g), len(exp))
			continue
		}
		for i := range exp {
			if math.Abs(g[i]-exp[i]) > tol {
				t.Errorf("%s[%d] : %v (Go) != %v (Python)", op, i, g[i], exp[i])
			}
		}
	}
}
