package xarray

import (
	"math"
	"testing"
)

// TestGeostationary valide la projection géostationnaire MTG contre des valeurs
// de référence pyproj (proj=geos, sweep=y, WGS84), avec aller-retour et gestion
// du limbe.
func TestGeostationary(t *testing.T) {
	g := MTGGeos()
	cases := []struct {
		lon, lat, ex, ey float64
	}{
		{0, 0, 0, 0},
		{2.3522, 48.8566, 162662.166, 4482303.767}, // Paris
		{-20, 10, -2117899.881, 1083600.230},       // Atlantique
		{30, -25, 2779841.289, -2568160.252},       // Afrique australe
	}
	for _, c := range cases {
		x, y, ok := g.Forward(c.lon, c.lat)
		if !ok {
			t.Fatalf("(%v,%v) invisible", c.lon, c.lat)
		}
		if math.Abs(x-c.ex) > 0.01 || math.Abs(y-c.ey) > 0.01 {
			t.Errorf("(%.4f,%.4f) -> (%.3f,%.3f), pyproj (%.3f,%.3f)", c.lon, c.lat, x, y, c.ex, c.ey)
		}
		lon, lat, ok := g.Inverse(x, y)
		if !ok || math.Abs(lon-c.lon) > 1e-6 || math.Abs(lat-c.lat) > 1e-6 {
			t.Errorf("aller-retour (%.6f,%.6f)", lon, lat)
		}
	}
	// Point à l'opposé du satellite : au-delà du limbe, non visible.
	if _, _, ok := g.Forward(180, 0); ok {
		t.Error("le point 180° devrait être invisible depuis le satellite à 0°")
	}
}
