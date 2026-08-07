package geoapi

import (
	"reflect"
	"testing"

	"github.com/benedictemarty/xarray"
)

func grilleGeo(t *testing.T) *xarray.DataArray[float64] {
	t.Helper()
	// latitude [45 44 43], longitude [0 1 2 3], données 0..11
	data := make([]float64, 12)
	for i := range data {
		data[i] = float64(i)
	}
	da, _ := xarray.NewDataArray(
		[]string{"latitude", "longitude"}, []int{3, 4}, data,
		map[string][]float64{"latitude": {45, 44, 43}, "longitude": {0, 1, 2, 3}}, "v")
	return da
}

func TestSubsetBBox(t *testing.T) {
	da := grilleGeo(t)
	// bbox : longitudes [1,2], latitudes [44,45] -> lat 45,44 ; lon 1,2
	sub, err := SubsetBBox(da, "longitude", "latitude", BBox{MinX: 1, MinY: 44, MaxX: 2, MaxY: 45})
	if err != nil {
		t.Fatalf("SubsetBBox : %v", err)
	}
	if !reflect.DeepEqual(sub.Shape(), []int{2, 2}) {
		t.Errorf("Shape = %v, attendu [2 2]", sub.Shape())
	}
	// lignes lat 45,44 (indices 0,1), colonnes lon 1,2 (indices 1,2)
	// data original ligne0=[0 1 2 3], ligne1=[4 5 6 7] -> [1 2 ; 5 6]
	if !reflect.DeepEqual(sub.Data(), []float64{1, 2, 5, 6}) {
		t.Errorf("Data = %v, attendu [1 2 5 6]", sub.Data())
	}
	if la, _ := sub.Coord("latitude"); !reflect.DeepEqual(la, []float64{45, 44}) {
		t.Errorf("latitude = %v", la)
	}
}

// TestSubsetBBoxGeoref verrouille la chaîne complète : un raster géoréférencé via
// une affine (axe y décroissant, comme les rasters réels) est découpé par une
// emprise exprimée en coordonnées monde.
func TestSubsetBBoxGeoref(t *testing.T) {
	// Raster 3×4 sans coords ; affine : origine (-10, 50), pixel 0.25, y décroissant.
	data := make([]float64, 12)
	for i := range data {
		data[i] = float64(i)
	}
	da, _ := xarray.NewDataArray([]string{"y", "x"}, []int{3, 4}, data, nil, "B04")
	gr := xarray.GeoRef{Transform: xarray.Affine{A: 0.25, B: 0, C: -10, D: 0, E: -0.25, F: 50}}
	g, err := da.Georeference(gr, "x", "y")
	if err != nil {
		t.Fatalf("Georeference : %v", err)
	}
	// Centres : x=[-9.875,-9.625,-9.375,-9.125], y=[49.875,49.625,49.375].
	// bbox monde x∈[-9.8,-9.3], y∈[49.4,49.7] -> colonnes 1,2 ; ligne 1 -> [5 6].
	sub, err := SubsetBBox(g, "x", "y", BBox{MinX: -9.8, MinY: 49.4, MaxX: -9.3, MaxY: 49.7})
	if err != nil {
		t.Fatalf("SubsetBBox : %v", err)
	}
	if !reflect.DeepEqual(sub.Shape(), []int{1, 2}) {
		t.Errorf("Shape = %v, attendu [1 2]", sub.Shape())
	}
	if !reflect.DeepEqual(sub.Data(), []float64{5, 6}) {
		t.Errorf("Data = %v, attendu [5 6]", sub.Data())
	}
}

func TestPosition(t *testing.T) {
	da := grilleGeo(t)
	// point (lon=0.9, lat=44.1) -> plus proche lon 1 (idx1), lat 44 (idx1)
	// data[lat=44][lon=1] = ligne1[1] = 5
	v, err := Position(da, "longitude", "latitude", 0.9, 44.1)
	if err != nil {
		t.Fatalf("Position : %v", err)
	}
	if v != 5 {
		t.Errorf("Position = %v, attendu 5", v)
	}
	// coin : (lon=3, lat=43) -> data[lat=43][lon=3] = ligne2[3] = 11
	v2, _ := Position(da, "longitude", "latitude", 3, 43)
	if v2 != 11 {
		t.Errorf("Position coin = %v, attendu 11", v2)
	}
}
