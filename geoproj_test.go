package xarray

import (
	"math"
	"testing"
)

// TestUTM valide la projection UTM (Transverse Mercator WGS84) contre des valeurs
// de référence pyproj, dans plusieurs zones et hémisphères, avec aller-retour.
func TestUTM(t *testing.T) {
	cases := []struct {
		crs      string
		lon, lat float64
		ex, ey   float64
	}{
		{"EPSG:32631", 2.3522, 48.8566, 452482.533, 5411717.177}, // zone 31N (Paris)
		{"EPSG:32632", 6.0, 45.0, 263553.974, 4987329.505},       // zone 32N
		{"EPSG:32721", -58.0, -34.0, 407650.397, 6237393.340},    // zone 21S (Buenos Aires)
	}
	for _, c := range cases {
		x, y, err := TransformXY("EPSG:4326", c.crs, c.lon, c.lat)
		if err != nil {
			t.Fatalf("%s : %v", c.crs, err)
		}
		if math.Abs(x-c.ex) > 0.01 || math.Abs(y-c.ey) > 0.01 {
			t.Errorf("%s -> (%.3f,%.3f), pyproj (%.3f,%.3f)", c.crs, x, y, c.ex, c.ey)
		}
		lon, lat, _ := TransformXY(c.crs, "EPSG:4326", x, y)
		if math.Abs(lon-c.lon) > 1e-7 || math.Abs(lat-c.lat) > 1e-7 {
			t.Errorf("%s aller-retour (%.7f,%.7f)", c.crs, lon, lat)
		}
	}
}

// TestTransformChain : transformation entre deux CRS projetés via le pivot
// géographique (UTM31N -> Web Mercator), cohérente avec la valeur directe.
func TestTransformChain(t *testing.T) {
	// Paris en UTM31N -> 3857 doit égaler Paris (4326) -> 3857.
	x, y, err := TransformXY("EPSG:32631", "EPSG:3857", 452482.533, 5411717.177)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(x-261845.706) > 0.5 || math.Abs(y-6250564.350) > 0.5 {
		t.Errorf("UTM31N->3857 = (%.3f,%.3f), attendu ~ (261845.7, 6250564.35)", x, y)
	}
}

func TestProjectionForUnsupported(t *testing.T) {
	if _, _, err := TransformXY("EPSG:4326", "EPSG:2154", 3, 45); err == nil {
		t.Error("erreur attendue : Lambert-93 non géré")
	}
	// Zone UTM invalide.
	if _, _, err := TransformXY("EPSG:4326", "EPSG:32699", 3, 45); err == nil {
		t.Error("erreur attendue : zone UTM 99 invalide")
	}
}
