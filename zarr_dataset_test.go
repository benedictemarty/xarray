package xarray

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestZarrDatasetConsolidatedMetadata vérifie l'écriture d'un .zmetadata
// consolidé listant les métadonnées de chaque array (permet consolidated=True
// côté zarr-python, sans RuntimeWarning).
func TestZarrDatasetConsolidatedMetadata(t *testing.T) {
	da, _ := NewDataArray([]string{"y", "x"}, []int{2, 2}, []float64{1, 2, 3, 4},
		map[string][]float64{"y": {0, 1}, "x": {0, 1}}, "v")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"v": da})
	dir := filepath.Join(t.TempDir(), "c.zarr")
	if err := WriteDatasetZarr(dir, ds, ZarrNone); err != nil {
		t.Fatalf("WriteDatasetZarr : %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".zmetadata"))
	if err != nil {
		t.Fatalf(".zmetadata absent : %v", err)
	}
	var m zconsolidated
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("json .zmetadata : %v", err)
	}
	if m.ZarrConsolidatedFormat != 1 {
		t.Errorf("zarr_consolidated_format = %d, attendu 1", m.ZarrConsolidatedFormat)
	}
	// Clés attendues : .zgroup racine + .zarray/.zattrs de v, x, y.
	for _, key := range []string{".zgroup", "v/.zarray", "v/.zattrs", "x/.zarray", "y/.zarray"} {
		if _, ok := m.Metadata[key]; !ok {
			t.Errorf("clé %q absente de .zmetadata (clés=%v)", key, keysOfRaw(m.Metadata))
		}
	}
}

func keysOfRaw(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestZarrDatasetAllerRetour(t *testing.T) {
	temp, _ := NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")
	pluie, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{100, 200},
		map[string][]float64{"temps": {2020, 2021}}, "pluie")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"temperature": temp, "pluie": pluie})

	dir := filepath.Join(t.TempDir(), "grp.zarr")
	if err := WriteDatasetZarr(dir, ds, ZarrZlib); err != nil {
		t.Fatalf("WriteDatasetZarr : %v", err)
	}
	// .zgroup présent.
	if _, err := os.Stat(filepath.Join(dir, ".zgroup")); err != nil {
		t.Errorf(".zgroup absent : %v", err)
	}

	got, err := ReadDatasetZarr(dir)
	if err != nil {
		t.Fatalf("ReadDatasetZarr : %v", err)
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
	// La coordonnée temps de temperature doit avoir été réattachée.
	if c, _ := tp.Coord("temps"); !reflect.DeepEqual(c, []float64{2020, 2021}) {
		t.Errorf("Coord(temps) sur temperature = %v", c)
	}
}

func TestZarrDatasetPasUnGroupe(t *testing.T) {
	dir := t.TempDir() // pas de .zgroup
	if _, err := ReadDatasetZarr(dir); err == nil {
		t.Error("erreur attendue : .zgroup absent")
	}
}
