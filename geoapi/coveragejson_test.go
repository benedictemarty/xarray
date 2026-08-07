package geoapi

import (
	"encoding/json"
	"testing"

	"github.com/bmarty/xarray"
)

func TestToCoverageJSON(t *testing.T) {
	// Grille 2×3 (latitude × longitude).
	da, _ := xarray.NewDataArray(
		[]string{"latitude", "longitude"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"latitude": {45, 44}, "longitude": {0, 1, 2}},
		"temperature",
	)
	b, err := ToCoverageJSON(da, "temperature", "longitude", "latitude")
	if err != nil {
		t.Fatalf("ToCoverageJSON : %v", err)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("JSON invalide : %v", err)
	}
	if doc["type"] != "Coverage" {
		t.Errorf("type = %v", doc["type"])
	}
	dom := doc["domain"].(map[string]interface{})
	if dom["domainType"] != "Grid" {
		t.Errorf("domainType = %v", dom["domainType"])
	}
	axes := dom["axes"].(map[string]interface{})
	x := axes["x"].(map[string]interface{})["values"].([]interface{})
	if len(x) != 3 {
		t.Errorf("axe x = %d valeurs, attendu 3", len(x))
	}
	rng := doc["ranges"].(map[string]interface{})["temperature"].(map[string]interface{})
	if rng["type"] != "NdArray" {
		t.Errorf("range type = %v", rng["type"])
	}
	vals := rng["values"].([]interface{})
	if len(vals) != 6 || vals[0].(float64) != 1 || vals[5].(float64) != 6 {
		t.Errorf("values = %v", vals)
	}
	shape := rng["shape"].([]interface{})
	if shape[0].(float64) != 2 || shape[1].(float64) != 3 {
		t.Errorf("shape = %v", shape)
	}
}

func TestToCoverageJSONErreurs(t *testing.T) {
	// 1D -> erreur.
	da1, _ := xarray.NewDataArray([]string{"x"}, []int{2}, []float64{1, 2},
		map[string][]float64{"x": {0, 1}}, "v")
	if _, err := ToCoverageJSON(da1, "v", "x", "y"); err == nil {
		t.Error("erreur attendue : pas 2D")
	}
	// Mauvais ordre de dimensions.
	da2, _ := xarray.NewDataArray([]string{"longitude", "latitude"}, []int{2, 2},
		[]float64{1, 2, 3, 4},
		map[string][]float64{"longitude": {0, 1}, "latitude": {45, 44}}, "v")
	if _, err := ToCoverageJSON(da2, "v", "longitude", "latitude"); err == nil {
		t.Error("erreur attendue : dimensions dans le mauvais ordre")
	}
}
