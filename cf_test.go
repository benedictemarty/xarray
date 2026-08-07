package xarray

import (
	"bytes"
	"math"
	"testing"
	"time"
)

// makeCF construit un Dataset avec attributs de packing pour tester l'aller-retour
// netCDF et le décodage CF.
func makeCF(t *testing.T) *Dataset[float64] {
	t.Helper()
	coords := map[string][]float64{"x": {0, 1, 2}}
	da, err := NewDataArray([]string{"x"}, []int{3}, []float64{10, 20, 999}, coords, "t2m")
	if err != nil {
		t.Fatal(err)
	}
	da.Variable().SetAttr("units", "K")
	da.Variable().SetAttr("long_name", "Température")
	da.Variable().SetAttr("scale_factor", "0.1")
	da.Variable().SetAttr("add_offset", "273.15")
	da.Variable().SetAttr("_FillValue", "999")
	ds, err := NewDataset(map[string]*DataArray[float64]{"t2m": da})
	if err != nil {
		t.Fatal(err)
	}
	return ds
}

// TestNetCDFAttrsRoundTrip vérifie que les attributs survivent à l'aller-retour
// (écriture puis relecture), y compris les attributs CF numériques.
func TestNetCDFAttrsRoundTrip(t *testing.T) {
	ds := makeCF(t)
	var buf bytes.Buffer
	if err := ds.WriteNetCDF(&buf); err != nil {
		t.Fatal(err)
	}
	got, err := ReadDatasetNetCDF[float64](&buf)
	if err != nil {
		t.Fatal(err)
	}
	da, err := got.Get("t2m")
	if err != nil {
		t.Fatal(err)
	}
	attrs := da.Variable().Attrs()
	if attrs["units"] != "K" {
		t.Errorf("units = %q", attrs["units"])
	}
	if attrs["long_name"] != "Température" {
		t.Errorf("long_name = %q", attrs["long_name"])
	}
	if attrs["scale_factor"] != "0.1" {
		t.Errorf("scale_factor = %q, attendu 0.1", attrs["scale_factor"])
	}
	if attrs["_FillValue"] != "999" {
		t.Errorf("_FillValue = %q", attrs["_FillValue"])
	}
}

// TestDecodeCF vérifie le dépacking : valeur = brut*scale + offset, fill -> NaN.
func TestDecodeCF(t *testing.T) {
	dec, err := DecodeCF(makeCF(t))
	if err != nil {
		t.Fatal(err)
	}
	da, _ := dec.Get("t2m")
	d := da.Data()
	// 10*0.1+273.15 = 274.15 ; 20*0.1+273.15 = 275.15 ; 999 -> NaN
	if math.Abs(d[0]-274.15) > 1e-9 || math.Abs(d[1]-275.15) > 1e-9 {
		t.Errorf("valeurs décodées = %v", d)
	}
	if !math.IsNaN(d[2]) {
		t.Errorf("_FillValue non converti en NaN: %v", d[2])
	}
	// Les attributs consommés sont retirés, units conservé.
	attrs := da.Variable().Attrs()
	if _, ok := attrs["scale_factor"]; ok {
		t.Error("scale_factor aurait dû être retiré après décodage")
	}
	if attrs["units"] != "K" {
		t.Errorf("units perdu: %q", attrs["units"])
	}
}

// TestDecodeCFNoop : sans attribut de packing, les données sont inchangées.
func TestDecodeCFNoop(t *testing.T) {
	coords := map[string][]float64{"x": {0, 1}}
	da, _ := NewDataArray([]string{"x"}, []int{2}, []float64{3, 4}, coords, "v")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"v": da})
	dec, err := DecodeCF(ds)
	if err != nil {
		t.Fatal(err)
	}
	dd, _ := dec.Get("v")
	if dd.Data()[0] != 3 || dd.Data()[1] != 4 {
		t.Errorf("données modifiées à tort: %v", dd.Data())
	}
}

func TestDecodeCFTime(t *testing.T) {
	// "hours since 2020-01-01 00:00:00", valeurs 0 et 24 -> ref et ref+1 jour.
	out, err := DecodeCFTime([]float64{0, 24}, "hours since 2020-01-01 00:00:00")
	if err != nil {
		t.Fatal(err)
	}
	ref := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if out[0] != EpochSeconds(ref) {
		t.Errorf("t0 = %v, attendu %v", out[0], EpochSeconds(ref))
	}
	if out[1]-out[0] != 86400 {
		t.Errorf("écart = %v, attendu 86400 s", out[1]-out[0])
	}
	if _, err := DecodeCFTime([]float64{0}, "bogus"); err == nil {
		t.Error("attendu une erreur sur units invalides")
	}
}

func TestDecodeTimeDataset(t *testing.T) {
	coords := map[string][]float64{"time": {0, 1}, "x": {0, 1}}
	da, _ := NewDataArray([]string{"time", "x"}, []int{2, 2}, []float64{1, 2, 3, 4}, coords, "v")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"v": da})
	// Pose l'attribut units sur la coordonnée time via un aller-retour netCDF.
	ds.coords["time"].SetAttr("units", "days since 2020-01-01")
	dec, err := DecodeTime(ds, "time")
	if err != nil {
		t.Fatal(err)
	}
	tc, _ := dec.Coord("time")
	ref := EpochSeconds(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	if tc[0] != ref || tc[1]-tc[0] != 86400 {
		t.Errorf("coord time décodée = %v", tc)
	}
}

func TestRejectUnlimitedDim(t *testing.T) {
	// En-tête CDF-1 avec numrecs=1 -> doit être refusé, pas paniquer.
	buf := []byte{'C', 'D', 'F', 1, 0, 0, 0, 1}
	if _, err := ReadDatasetNetCDF[float64](bytes.NewReader(buf)); err == nil {
		t.Error("attendu une erreur pour numrecs != 0")
	}
}
