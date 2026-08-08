package xarray

import (
	"math"
	"reflect"
	"testing"
)

// TestReprojectIdentity : reprojeter vers le même CRS avec la même grille
// redonne exactement la source (plus proche voisin sur les centres de pixels).
func TestReprojectIdentity(t *testing.T) {
	sw, sh := 6, 4
	src := make([]float64, sw*sh)
	for i := range src {
		src[i] = float64(i)
	}
	tr := Affine{A: 1, C: 0, E: -1, F: 10}
	out, err := ReprojectNearest(src, sw, sh, tr, "EPSG:4326", tr, sw, sh, "EPSG:4326")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, src) {
		t.Errorf("identité = %v", out)
	}
}

// TestReprojectWebMercator : reprojection 4326 -> 3857, comparée au résultat de
// référence (pyproj + plus proche voisin), grille figée dans le test.
func TestReprojectWebMercator(t *testing.T) {
	sw, sh := 6, 4
	src := make([]float64, sw*sh)
	for i := range src {
		src[i] = float64(i)
	}
	srcT := Affine{A: 1, C: 0, E: -1, F: 10}
	x0, y0, _ := TransformXY("EPSG:4326", "EPSG:3857", 0.0, 10.0)
	x1, y1, _ := TransformXY("EPSG:4326", "EPSG:3857", 6.0, 6.0)
	dw, dh := 5, 5
	dstT := Affine{A: (x1 - x0) / float64(dw), C: x0, E: (y1 - y0) / float64(dh), F: y0}
	out, err := ReprojectNearest(src, sw, sh, srcT, "EPSG:4326", dstT, dw, dh, "EPSG:3857")
	if err != nil {
		t.Fatal(err)
	}
	// Référence validée contre pyproj (voir CHANGELOG Sprint 77).
	want := []float64{
		0, 1, 3, 4, 5,
		6, 7, 9, 10, 11,
		6, 7, 9, 10, 11, // ligne dupliquée : compression Web Mercator
		12, 13, 15, 16, 17,
		18, 19, 21, 22, 23,
	}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("reprojection = %v\nattendu       %v", out, want)
	}
}

// TestReprojectOutOfBounds : les pixels cibles hors de la source valent NaN.
func TestReprojectOutOfBounds(t *testing.T) {
	src := []float64{1, 2, 3, 4}
	tr := Affine{A: 1, C: 0, E: -1, F: 2}
	// Grille cible décalée de 100° : entièrement hors source.
	dstT := Affine{A: 1, C: 100, E: -1, F: 2}
	out, err := ReprojectNearest(src, 2, 2, tr, "EPSG:4326", dstT, 2, 2, "EPSG:4326")
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range out {
		if !math.IsNaN(v) {
			t.Errorf("attendu NaN hors emprise, obtenu %v", v)
		}
	}
}

func TestReprojectUnsupportedCRS(t *testing.T) {
	src := []float64{1, 2, 3, 4}
	tr := Affine{A: 1, E: -1, F: 2}
	if _, err := ReprojectNearest(src, 2, 2, tr, "EPSG:4326", tr, 2, 2, "EPSG:2154"); err == nil {
		t.Error("erreur attendue : CRS non pris en charge")
	}
}
