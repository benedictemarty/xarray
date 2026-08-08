package xarray

import (
	"math"
	"testing"
)

// TestWebMercator valide la transformation EPSG:4326 ↔ EPSG:3857 contre des
// valeurs de référence produites par pyproj (Transformer, always_xy).
func TestWebMercator(t *testing.T) {
	ref := []struct{ lon, lat, ex, ey float64 }{
		{2.3522, 48.8566, 261845.706, 6250564.350},  // Paris
		{-122.4, 37.77, -13625505.673, 4546985.284}, // San Francisco
		{139.69, 35.68, 15550219.669, 4256678.732},  // Tokyo
		{-10, -80, -1113194.908, -15538711.096},     // haute latitude sud
	}
	for _, p := range ref {
		x, y, err := TransformXY("EPSG:4326", "EPSG:3857", p.lon, p.lat)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(x-p.ex) > 0.01 || math.Abs(y-p.ey) > 0.01 {
			t.Errorf("(%.4f,%.4f) -> (%.3f,%.3f), pyproj (%.3f,%.3f)", p.lon, p.lat, x, y, p.ex, p.ey)
		}
		// Aller-retour exact.
		lon, lat, _ := TransformXY("EPSG:3857", "EPSG:4326", x, y)
		if math.Abs(lon-p.lon) > 1e-9 || math.Abs(lat-p.lat) > 1e-9 {
			t.Errorf("aller-retour (%.4f,%.4f) -> (%.9f,%.9f)", p.lon, p.lat, lon, lat)
		}
	}
}

func TestTransformXYCases(t *testing.T) {
	// Identité (même CRS, alias reconnus).
	x, y, err := TransformXY("CRS84", "EPSG:4326", 3, 45)
	if err != nil || x != 3 || y != 45 {
		t.Errorf("identité = (%v,%v,%v)", x, y, err)
	}
	// Paire non prise en charge -> erreur explicite.
	if _, _, err := TransformXY("EPSG:4326", "EPSG:27572", 3, 45); err == nil {
		t.Error("erreur attendue : CRS non géré (Lambert II historique)")
	}
}
