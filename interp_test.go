package xarray

import (
	"math"
	"reflect"
	"testing"
)

func TestBracketF(t *testing.T) {
	asc := []float64{0, 1, 2, 3}
	if i0, i1, w, ok := bracketF(asc, 1.5); !ok || i0 != 1 || i1 != 2 || math.Abs(w-0.5) > 1e-9 {
		t.Errorf("croissant = %d,%d,%v,%v", i0, i1, w, ok)
	}
	desc := []float64{3, 2, 1, 0} // décroissant (cas latitude)
	if i0, i1, w, ok := bracketF(desc, 2.25); !ok || i0 != 0 || i1 != 1 || math.Abs(w-0.75) > 1e-9 {
		t.Errorf("décroissant = %d,%d,%v,%v", i0, i1, w, ok)
	}
	if _, _, _, ok := bracketF(asc, 9); ok {
		t.Error("hors axe devrait donner ok=false")
	}
}

// TestInterpBilinear : interpolation au point sur une grille 2×2, y décroissant.
func TestInterpBilinear(t *testing.T) {
	// v(lat, lon) : (1,0)=0 (1,1)=10 (0,0)=20 (0,1)=30.
	da, _ := NewDataArray([]string{"latitude", "longitude"}, []int{2, 2},
		[]float64{0, 10, 20, 30},
		map[string][]float64{"latitude": {1, 0}, "longitude": {0, 1}}, "v")

	// Centre (0.5, 0.5) : moyenne des 4 = 15.
	r, err := InterpBilinear(da, "longitude", "latitude", 0.5, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if r.Data()[0] != 15 {
		t.Errorf("centre = %v, attendu 15", r.Data()[0])
	}
	// Champ linéaire : le long de lon à lat=1, v = 10*lon. En (0.3, 1) -> 3.
	if r, _ := InterpBilinear(da, "longitude", "latitude", 0.3, 1); math.Abs(r.Data()[0]-3) > 1e-9 {
		t.Errorf("(0.3,1) = %v, attendu 3", r.Data()[0])
	}
	// Coin exact (0,1) : lat=1, lon=0 -> 0.
	if r, _ := InterpBilinear(da, "longitude", "latitude", 0, 1); r.Data()[0] != 0 {
		t.Errorf("coin = %v, attendu 0", r.Data()[0])
	}
	// Hors grille -> erreur.
	if _, err := InterpBilinear(da, "longitude", "latitude", 5, 5); err == nil {
		t.Error("erreur attendue : hors grille")
	}
}

// TestInterpBilinearKeepsDims : les dimensions supplémentaires (temps) sont
// conservées, seules x/y sont réduites.
func TestInterpBilinearKeepsDims(t *testing.T) {
	// dims [time, lat, lon], time=2.
	data := []float64{
		0, 10, 20, 30, // t0
		100, 110, 120, 130, // t1
	}
	da, _ := NewDataArray([]string{"time", "latitude", "longitude"}, []int{2, 2, 2}, data,
		map[string][]float64{"time": {0, 1}, "latitude": {1, 0}, "longitude": {0, 1}}, "v")
	r, err := InterpBilinear(da, "longitude", "latitude", 0.5, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r.Dims(), []string{"time"}) {
		t.Errorf("dims restantes = %v, attendu [time]", r.Dims())
	}
	// t0 -> 15, t1 -> 115.
	if !reflect.DeepEqual(r.Data(), []float64{15, 115}) {
		t.Errorf("data = %v, attendu [15 115]", r.Data())
	}
}
