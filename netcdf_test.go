package xarray

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNetCDFDataArrayAllerRetour(t *testing.T) {
	da, _ := NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")

	var buf bytes.Buffer
	if err := da.WriteNetCDF(&buf); err != nil {
		t.Fatalf("WriteNetCDF : %v", err)
	}
	// Signature CDF-1 attendue.
	if !bytes.HasPrefix(buf.Bytes(), []byte{'C', 'D', 'F', 1}) {
		t.Fatal("signature CDF-1 absente")
	}

	got, err := ReadDataArrayNetCDF[float64](&buf)
	if err != nil {
		t.Fatalf("ReadDataArrayNetCDF : %v", err)
	}
	if got.Name() != "temperature" {
		t.Errorf("Name = %q", got.Name())
	}
	if !reflect.DeepEqual(got.Dims(), []string{"temps", "lieu"}) {
		t.Errorf("Dims = %v", got.Dims())
	}
	if !reflect.DeepEqual(got.Data(), []float64{1, 2, 3, 4, 5, 6}) {
		t.Errorf("Data = %v", got.Data())
	}
	c, _ := got.Coord("lieu")
	if !reflect.DeepEqual(c, []float64{10, 20, 30}) {
		t.Errorf("Coord(lieu) = %v", c)
	}
}

func TestNetCDFDatasetAllerRetour(t *testing.T) {
	temp, _ := NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")
	pluie, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{100, 200},
		map[string][]float64{"temps": {2020, 2021}}, "pluie")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"temperature": temp, "pluie": pluie})

	var buf bytes.Buffer
	if err := ds.WriteNetCDF(&buf); err != nil {
		t.Fatalf("WriteNetCDF : %v", err)
	}
	got, err := ReadDatasetNetCDF[float64](&buf)
	if err != nil {
		t.Fatalf("ReadDatasetNetCDF : %v", err)
	}
	if !reflect.DeepEqual(got.VarNames(), []string{"pluie", "temperature"}) {
		t.Errorf("VarNames = %v", got.VarNames())
	}
	tp, _ := got.Get("temperature")
	if !reflect.DeepEqual(tp.Data(), []float64{1, 2, 3, 4, 5, 6}) {
		t.Errorf("temperature = %v", tp.Data())
	}
	if c, _ := got.Coord("lieu"); !reflect.DeepEqual(c, []float64{10, 20, 30}) {
		t.Errorf("Coord(lieu) = %v", c)
	}
	pl, _ := got.Get("pluie")
	if !reflect.DeepEqual(pl.Data(), []float64{100, 200}) {
		t.Errorf("pluie = %v", pl.Data())
	}
}

func TestNetCDFFloat32(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{4}, []float32{1.5, 2.5, 3.5, 4.5},
		map[string][]float32{"x": {0, 1, 2, 3}}, "v")
	var buf bytes.Buffer
	if err := da.WriteNetCDF(&buf); err != nil {
		t.Fatalf("WriteNetCDF float32 : %v", err)
	}
	got, err := ReadDataArrayNetCDF[float32](&buf)
	if err != nil {
		t.Fatalf("ReadDataArrayNetCDF float32 : %v", err)
	}
	if !reflect.DeepEqual(got.Data(), []float32{1.5, 2.5, 3.5, 4.5}) {
		t.Errorf("Data = %v", got.Data())
	}
}

func TestNetCDFInt32(t *testing.T) {
	da, _ := NewDataArray([]string{"x", "y"}, []int{2, 2}, []int32{10, 20, 30, 40},
		map[string][]int32{"x": {0, 1}, "y": {5, 6}}, "compteur")
	var buf bytes.Buffer
	if err := da.WriteNetCDF(&buf); err != nil {
		t.Fatalf("WriteNetCDF int32 : %v", err)
	}
	got, err := ReadDataArrayNetCDF[int32](&buf)
	if err != nil {
		t.Fatalf("ReadDataArrayNetCDF int32 : %v", err)
	}
	if !reflect.DeepEqual(got.Data(), []int32{10, 20, 30, 40}) {
		t.Errorf("Data = %v", got.Data())
	}
	if c, _ := got.Coord("y"); !reflect.DeepEqual(c, []int32{5, 6}) {
		t.Errorf("Coord(y) = %v", c)
	}
}

func TestNetCDFTypeNonSupporte(t *testing.T) {
	// int (64 bits) n'a pas d'équivalent en CDF-1 : erreur attendue.
	da, _ := NewDataArray([]string{"x"}, []int{2}, []int{1, 2}, nil, "v")
	var buf bytes.Buffer
	if err := da.WriteNetCDF(&buf); err == nil {
		t.Error("erreur attendue : type int non supporté par l'export netCDF")
	}
}

func TestNetCDFSignatureInvalide(t *testing.T) {
	if _, err := ReadDatasetNetCDF[float64](bytes.NewBufferString("XXXX")); err == nil {
		t.Error("erreur attendue : signature invalide")
	}
}
