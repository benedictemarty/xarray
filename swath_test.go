package xarray

import (
	"math"
	"testing"
)

// TestResampleSwathNearest vérifie le rééchantillonnage d'une fauchée (pixels
// géolocalisés par lon/lat propres) vers une grille régulière : binning correct,
// moyenne des collisions, cellules vides à NaN.
func TestResampleSwathNearest(t *testing.T) {
	// 4 pixels dont 2 tombent dans la même cellule (moyenne attendue).
	data := []float64{10, 20, 30, 40}
	lon := []float64{0.2, 0.8, 1.2, 5.0} // les deux premiers -> cellule (col 0)
	lat := []float64{9.5, 9.6, 9.5, 5.0}
	dstT := Affine{A: 1, C: 0, E: -1, F: 10} // 6×10 : lon 0..6, lat 10..0
	grid, counts, err := ResampleSwathNearest(data, lon, lat, dstT, 6, 10)
	if err != nil {
		t.Fatal(err)
	}
	// Cellule (row 0, col 0) : pixels 0 et 1 -> moyenne (10+20)/2 = 15.
	if counts[0] != 2 || math.Abs(grid[0]-15) > 1e-9 {
		t.Errorf("cellule[0,0] = %v (n=%d), attendu 15 (n=2)", grid[0], counts[0])
	}
	// Cellule (row 0, col 1) : pixel 2 (lon 1.2) -> 30.
	if grid[1] != 30 {
		t.Errorf("cellule[0,1] = %v, attendu 30", grid[1])
	}
	// Une cellule vide -> NaN.
	if !math.IsNaN(grid[3*6+3]) {
		t.Errorf("cellule vide attendue NaN, obtenu %v", grid[3*6+3])
	}
	// Longueurs incohérentes -> erreur.
	if _, _, err := ResampleSwathNearest(data, lon[:2], lat, dstT, 6, 10); err == nil {
		t.Error("erreur attendue : longueurs différentes")
	}
}

// TestResampleSwathNearestRadius : le rayon de recherche comble les cellules
// vides d'une grille plus fine que la fauchée.
func TestResampleSwathNearestRadius(t *testing.T) {
	n := 120
	data := make([]float64, n)
	lon := make([]float64, n)
	lat := make([]float64, n)
	for k := 0; k < n; k++ {
		f := float64(k) / float64(n-1)
		lon[k], lat[k], data[k] = f*10, 40+f*10, 100
	}
	dstT := Affine{A: 0.25, C: 0, E: -0.25, F: 50} // grille fine 40×40
	g0, err := ResampleSwathNearestRadius(data, lon, lat, dstT, 40, 40, 0)
	if err != nil {
		t.Fatal(err)
	}
	g2, _ := ResampleSwathNearestRadius(data, lon, lat, dstT, 40, 40, 2)
	filled := func(g []float64) int {
		n := 0
		for _, v := range g {
			if !math.IsNaN(v) {
				n++
			}
		}
		return n
	}
	if filled(g2) <= filled(g0) {
		t.Errorf("le rayon devrait combler des trous : r0=%d, r2=%d", filled(g0), filled(g2))
	}
	// La valeur reste celle du plus proche pixel (ici 100 partout où rempli).
	for _, v := range g2 {
		if !math.IsNaN(v) && v != 100 {
			t.Errorf("valeur = %v, attendu 100", v)
		}
	}
}

// TestSwathToDataArray : une fauchée rééchantillonnée devient un DataArray
// géoréférencé (coordonnées lon/lat régulières, CRS EPSG:4326).
func TestSwathToDataArray(t *testing.T) {
	data := []float64{5, 7}
	lon := []float64{0.5, 2.5}
	lat := []float64{9.5, 9.5}
	dstT := Affine{A: 1, C: 0, E: -1, F: 10}
	da, err := SwathToDataArray(data, lon, lat, dstT, 4, 4, "v", "latitude", "longitude")
	if err != nil {
		t.Fatal(err)
	}
	if xs, _ := da.Coord("longitude"); xs[0] != 0.5 {
		t.Errorf("coord longitude = %v", xs)
	}
	if da.Variable().Attrs()["crs"] != "EPSG:4326" {
		t.Errorf("CRS attendu EPSG:4326")
	}
	if da.Data()[0] != 5 { // pixel lon 0.5 -> cellule (0,0)
		t.Errorf("cellule[0,0] = %v, attendu 5", da.Data()[0])
	}
}
