package xarray

import (
	"math"
	"reflect"
	"testing"
)

// TestAffineApplyInverse valide la géotransformation contre les valeurs de
// référence de la bibliothèque `affine` (GDAL/rasterio) : origine (-10, 50),
// pixel 0.25°, y décroissant.
func TestAffineApplyInverse(t *testing.T) {
	tr := Affine{A: 0.25, B: 0, C: -10, D: 0, E: -0.25, F: 50}
	x, y := tr.Apply(3.5, 2.5) // centre du pixel (col=3, row=2)
	if math.Abs(x+9.125) > 1e-9 || math.Abs(y-49.375) > 1e-9 {
		t.Errorf("Apply = (%v, %v), attendu (-9.125, 49.375)", x, y)
	}
	inv, err := tr.Inverse()
	if err != nil {
		t.Fatal(err)
	}
	col, row := inv.Apply(x, y)
	if math.Abs(col-3.5) > 1e-9 || math.Abs(row-2.5) > 1e-9 {
		t.Errorf("Inverse.Apply = (%v, %v), attendu (3.5, 2.5)", col, row)
	}
	// Transformation dégénérée -> erreur.
	if _, err := (Affine{}).Inverse(); err == nil {
		t.Error("erreur attendue : transformation non inversible")
	}
}

func TestAffineGDALRoundTrip(t *testing.T) {
	tr := Affine{A: 0.25, B: 0, C: -10, D: 0, E: -0.25, F: 50}
	gt := tr.GDAL()
	if gt != [6]float64{-10, 0.25, 0, 50, 0, -0.25} {
		t.Errorf("GDAL = %v", gt)
	}
	if FromGDAL(gt) != tr {
		t.Errorf("FromGDAL(GDAL) != identité")
	}
}

func TestGeoreference(t *testing.T) {
	tr := Affine{A: 0.25, B: 0, C: -10, D: 0, E: -0.25, F: 50}
	da, _ := NewDataArray([]string{"y", "x"}, []int{2, 4},
		[]float64{0, 1, 2, 3, 4, 5, 6, 7}, nil, "B04")
	g, err := da.Georeference(GeoRef{Transform: tr, CRS: "EPSG:4326"}, "x", "y")
	if err != nil {
		t.Fatalf("Georeference : %v", err)
	}
	xs, _ := g.Coord("x")
	ys, _ := g.Coord("y")
	// Centres de pixels : x0 = -10 + 0.5·0.25 = -9.875 ; y0 = 50 - 0.5·0.25 = 49.875.
	if !reflect.DeepEqual(xs, []float64{-9.875, -9.625, -9.375, -9.125}) {
		t.Errorf("coord x = %v", xs)
	}
	if !reflect.DeepEqual(ys, []float64{49.875, 49.625}) {
		t.Errorf("coord y = %v", ys)
	}
	if g.Variable().Attrs()["crs"] != "EPSG:4326" {
		t.Errorf("CRS non stocké : %v", g.Variable().Attrs())
	}
	// Mauvaises dimensions -> erreur.
	bad, _ := NewDataArray([]string{"x", "y"}, []int{4, 2}, make([]float64, 8), nil, "b")
	if _, err := bad.Georeference(GeoRef{Transform: tr}, "x", "y"); err == nil {
		t.Error("erreur attendue : dimensions inattendues")
	}
}

func TestParseGDALGeoTransform(t *testing.T) {
	tr, err := ParseGDALGeoTransform("-10.0 0.25 0.0 50.0 0.0 -0.25")
	if err != nil {
		t.Fatal(err)
	}
	if tr != (Affine{A: 0.25, B: 0, C: -10, D: 0, E: -0.25, F: 50}) {
		t.Errorf("affine = %+v", tr)
	}
	if _, err := ParseGDALGeoTransform("1 2 3"); err == nil {
		t.Error("erreur attendue : nombre de champs incorrect")
	}
}

// TestGeoRefFromCF lit un Zarr géoréférencé (convention rioxarray/CF :
// spatial_ref + GeoTransform + grid_mapping) et vérifie l'extraction automatique
// du CRS et de l'affine, puis leur application aux coordonnées.
func TestGeoRefFromCF(t *testing.T) {
	ds, err := ReadDatasetZarr("testdata/zarr_georef")
	if err != nil {
		t.Fatalf("ReadDatasetZarr : %v", err)
	}
	gr, ok := ds.GeoRefFromCF("B04")
	if !ok {
		t.Fatal("géoréférencement CF non extrait")
	}
	if gr.CRS != "EPSG:4326" {
		t.Errorf("CRS = %q", gr.CRS)
	}
	if gr.Transform != (Affine{A: 0.25, B: 0, C: -10, D: 0, E: -0.25, F: 50}) {
		t.Errorf("affine = %+v", gr.Transform)
	}
	b, _ := ds.Get("B04")
	g, err := b.Georeference(gr, "x", "y")
	if err != nil {
		t.Fatalf("Georeference : %v", err)
	}
	if xs, _ := g.Coord("x"); !reflect.DeepEqual(xs, []float64{-9.875, -9.625, -9.375, -9.125}) {
		t.Errorf("coord x = %v", xs)
	}
	// Variable sans grid_mapping -> non trouvé.
	if _, ok := ds.GeoRefFromCF("spatial_ref"); ok {
		t.Error("grid_mapping inexistant devrait donner ok=false")
	}
}

func TestGeoCoordsRotationRejected(t *testing.T) {
	// Rotation (B != 0) -> pas d'axes 1D séparables.
	if _, _, err := GeoCoords(Affine{A: 1, B: 0.1, D: 0, E: 1}, 4, 4); err == nil {
		t.Error("erreur attendue : rotation non représentable en axes 1D")
	}
}
