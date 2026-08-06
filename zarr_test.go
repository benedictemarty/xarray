package xarray

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestZarrAllerRetour(t *testing.T) {
	da, _ := NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")

	dir := filepath.Join(t.TempDir(), "arr.zarr")
	if err := WriteDataArrayZarr(dir, da, []int{2, 2}, ZarrNone); err != nil {
		t.Fatalf("WriteDataArrayZarr : %v", err)
	}
	// Le .zarray et le .zattrs doivent exister.
	if _, err := os.Stat(filepath.Join(dir, ".zarray")); err != nil {
		t.Errorf(".zarray absent : %v", err)
	}

	got, err := ReadDataArrayZarr(dir)
	if err != nil {
		t.Fatalf("ReadDataArrayZarr : %v", err)
	}
	if !reflect.DeepEqual(got.Dims(), []string{"temps", "lieu"}) {
		t.Errorf("Dims = %v", got.Dims())
	}
	if !reflect.DeepEqual(got.Data(), []float64{1, 2, 3, 4, 5, 6}) {
		t.Errorf("Data = %v", got.Data())
	}
	if got.Name() != "temperature" {
		t.Errorf("Name = %q", got.Name())
	}
	if c, _ := got.Coord("lieu"); !reflect.DeepEqual(c, []float64{10, 20, 30}) {
		t.Errorf("Coord(lieu) = %v", c)
	}
}

func TestZarrChunksNonAlignes(t *testing.T) {
	// Forme 5×4 avec chunks 2×3 : chunks de bord partiels (complétés par fill).
	data := make([]float64, 20)
	for i := range data {
		data[i] = float64(i + 1)
	}
	da, _ := NewDataArray([]string{"x", "y"}, []int{5, 4}, data, nil, "v")

	dir := filepath.Join(t.TempDir(), "nb.zarr")
	if err := WriteDataArrayZarr(dir, da, []int{2, 3}, ZarrNone); err != nil {
		t.Fatalf("Write : %v", err)
	}
	got, err := ReadDataArrayZarr(dir)
	if err != nil {
		t.Fatalf("Read : %v", err)
	}
	if !reflect.DeepEqual(got.Shape(), []int{5, 4}) {
		t.Errorf("Shape = %v", got.Shape())
	}
	if !reflect.DeepEqual(got.Data(), data) {
		t.Errorf("Data = %v", got.Data())
	}
}

func TestZarrCompressionZlib(t *testing.T) {
	data := make([]float64, 100)
	for i := range data {
		data[i] = float64(i) * 0.5
	}
	da, _ := NewDataArray([]string{"x", "y"}, []int{10, 10}, data, nil, "z")

	dir := filepath.Join(t.TempDir(), "z.zarr")
	if err := WriteDataArrayZarr(dir, da, []int{4, 4}, ZarrZlib); err != nil {
		t.Fatalf("Write zlib : %v", err)
	}
	got, err := ReadDataArrayZarr(dir)
	if err != nil {
		t.Fatalf("Read zlib : %v", err)
	}
	if !reflect.DeepEqual(got.Data(), data) {
		t.Errorf("Data zlib != original")
	}
}

func TestZarrChunksInvalides(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3}, nil, "v")
	dir := filepath.Join(t.TempDir(), "bad.zarr")
	if err := WriteDataArrayZarr(dir, da, []int{2, 2}, ZarrNone); err == nil {
		t.Error("erreur attendue : nombre de chunks incohérent")
	}
	if err := WriteDataArrayZarr(dir, da, []int{0}, ZarrNone); err == nil {
		t.Error("erreur attendue : taille de chunk nulle")
	}
}

func TestZarrDtypeNonSupporte(t *testing.T) {
	dir := t.TempDir()
	writeJSONFile(filepath.Join(dir, ".zarray"), zarrayMeta{
		ZarrFormat: 2, Shape: []int{2}, Chunks: []int{2}, Dtype: "<i4", Order: "C",
	})
	writeJSONFile(filepath.Join(dir, ".zattrs"), zattrsMeta{Dims: []string{"x"}})
	if _, err := ReadDataArrayZarr(dir); err == nil {
		t.Error("erreur attendue : dtype non supporté")
	}
}
