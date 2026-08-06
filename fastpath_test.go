package xarray

import (
	"reflect"
	"testing"
)

// Vérifie que le chemin rapide float64 (sans closure) donne exactement le même
// résultat que le chemin générique, pour les quatre opérations.
func TestFastPathEquivautGenerique(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{4}, []float64{6, 8, 10, 12},
		map[string][]float64{"x": {0, 1, 2, 3}}, "a")
	b, _ := NewDataArray([]string{"x"}, []int{4}, []float64{2, 4, 5, 3},
		map[string][]float64{"x": {0, 1, 2, 3}}, "b")

	cas := []struct {
		nom    string
		rapide func() (*DataArray[float64], error)
		gen    func(x, y float64) float64
	}{
		{"Add", func() (*DataArray[float64], error) { return a.Add(b) }, func(x, y float64) float64 { return x + y }},
		{"Sub", func() (*DataArray[float64], error) { return a.Sub(b) }, func(x, y float64) float64 { return x - y }},
		{"Mul", func() (*DataArray[float64], error) { return a.Mul(b) }, func(x, y float64) float64 { return x * y }},
		{"Div", func() (*DataArray[float64], error) { return a.Div(b) }, func(x, y float64) float64 { return x / y }},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			rapide, err := c.rapide()
			if err != nil {
				t.Fatal(err)
			}
			gen, _ := a.binary(b, c.gen)
			if !reflect.DeepEqual(rapide.Data(), gen.Data()) {
				t.Errorf("%s : rapide %v != générique %v", c.nom, rapide.Data(), gen.Data())
			}
			// Les coordonnées doivent être conservées.
			if crd, _ := rapide.Coord("x"); !reflect.DeepEqual(crd, []float64{0, 1, 2, 3}) {
				t.Errorf("%s : coord = %v", c.nom, crd)
			}
		})
	}
}

// Le chemin rapide ne doit pas s'appliquer au broadcasting (formes différentes)
// mais retomber correctement sur le générique.
func TestFastPathBroadcastRepli(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2}, nil, "a")
	b, _ := NewDataArray([]string{"y"}, []int{3}, []float64{10, 20, 30}, nil, "b")
	r, err := a.Mul(b)
	if err != nil {
		t.Fatalf("Mul broadcast : %v", err)
	}
	// [[1*10,1*20,1*30],[2*10,2*20,2*30]]
	if !reflect.DeepEqual(r.Data(), []float64{10, 20, 30, 20, 40, 60}) {
		t.Errorf("Mul broadcast = %v", r.Data())
	}
}
