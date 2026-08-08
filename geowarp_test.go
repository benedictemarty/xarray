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

// TestSampling vérifie les noyaux d'échantillonnage (centre de pixel i à i+0.5).
func TestSampling(t *testing.T) {
	// gradient 6×4 : src[r][c] = 10r + c.
	w, h := 6, 4
	src := make([]float64, w*h)
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			src[r*w+c] = float64(10*r + c)
		}
	}
	// Plus proche voisin : (col=2.7,row=1.2) -> pixel (2,1) = 12.
	if v := sampleNearest(src, w, h, 2.7, 1.2); v != 12 {
		t.Errorf("nearest = %v, attendu 12", v)
	}
	// Bilinéaire au centre exact d'un pixel -> valeur du pixel.
	if v := sampleBilinear(src, w, h, 1.5, 1.5); v != 11 {
		t.Errorf("bilinéaire centre = %v, attendu 11", v)
	}
	// Bilinéaire à mi-chemin des 4 pixels {1,2,11,12} -> moyenne 6.5.
	if v := sampleBilinear(src, w, h, 2.0, 1.0); v != 6.5 {
		t.Errorf("bilinéaire milieu = %v, attendu 6.5", v)
	}
	// Bilinéaire exact sur un champ linéaire : f(col,row)=10(row-0.5)+(col-0.5).
	if v := sampleBilinear(src, w, h, 3.2, 2.4); math.Abs(v-(10*(2.4-0.5)+(3.2-0.5))) > 1e-12 {
		t.Errorf("bilinéaire linéaire = %v", v)
	}
	// Hors bornes -> NaN.
	if !math.IsNaN(sampleBilinear(src, w, h, 0.2, 0.2)) {
		t.Error("bord attendu NaN (bilinéaire)")
	}
	if !math.IsNaN(sampleNearest(src, w, h, -1, 0)) {
		t.Error("hors bornes attendu NaN (nearest)")
	}
}

func TestReprojectUnsupportedCRS(t *testing.T) {
	src := []float64{1, 2, 3, 4}
	tr := Affine{A: 1, E: -1, F: 2}
	if _, err := ReprojectNearest(src, 2, 2, tr, "EPSG:4326", tr, 2, 2, "EPSG:27572"); err == nil {
		t.Error("erreur attendue : CRS non pris en charge")
	}
}
