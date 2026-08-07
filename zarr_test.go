package xarray

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestZarrReadIntDtypes lit un store Zarr réel (fixture testdata/) dont les
// coordonnées sont en int64 (`<i8`) et une variable en int32 (`<i4`) — cas
// courant de zarr-python. Ils doivent être convertis en float64.
func TestZarrReadIntDtypes(t *testing.T) {
	ds, err := ReadDatasetZarr("testdata/zarr_int_dtypes")
	if err != nil {
		t.Fatalf("ReadDatasetZarr (dtypes entiers) : %v", err)
	}
	if yv, _ := ds.Coord("y"); !reflect.DeepEqual(yv, []float64{45, 44, 43}) {
		t.Errorf("coord y (i8) = %v, attendu [45 44 43]", yv)
	}
	if xv, _ := ds.Coord("x"); !reflect.DeepEqual(xv, []float64{0, 1, 2, 3}) {
		t.Errorf("coord x (i8) = %v, attendu [0 1 2 3]", xv)
	}
	cnt, err := ds.Get("count")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cnt.Data(), []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}) {
		t.Errorf("count (i4) = %v", cnt.Data())
	}
}

// TestZarrReadBloscZstd lit un store Blosc dont le codec est zstd (via
// github.com/klauspost/compress). v[i] = i % 50 sur 2000 valeurs.
func TestZarrReadBloscZstd(t *testing.T) {
	da, err := ReadDataArrayZarr("testdata/zarr_blosc_zstd")
	if err != nil {
		t.Fatalf("ReadDataArrayZarr (zstd) : %v", err)
	}
	d := da.Data()
	if len(d) != 2000 {
		t.Fatalf("taille = %d, attendu 2000", len(d))
	}
	for i, x := range d {
		if x != float64(i%50) {
			t.Fatalf("v[%d] = %v, attendu %d (zstd incorrect)", i, x, i%50)
		}
	}
}

// TestBitUnshuffleRemainder vérifie que le bitshuffle avec nelem non multiple de
// 8 est refusé par une erreur explicite (cas non pris en charge), plutôt que de
// produire des données fausses.
func TestBitUnshuffleRemainder(t *testing.T) {
	// 100 éléments d'un octet chacun -> nelem = 100, non multiple de 8.
	if _, err := bitUnshuffle(make([]byte, 100), 1); err == nil {
		t.Error("erreur attendue : bitshuffle nelem non multiple de 8")
	}
	// 64 éléments -> multiple de 8, doit passer.
	if _, err := bitUnshuffle(make([]byte, 64), 1); err != nil {
		t.Errorf("nelem multiple de 8 refusé à tort : %v", err)
	}
}

// TestZarrReadBloscBitshuffle lit un store Blosc avec filtre bitshuffle
// (shuffle=2), produit par zarr-python. v[i] = i % 16 sur 16×16.
func TestZarrReadBloscBitshuffle(t *testing.T) {
	da, err := ReadDataArrayZarr("testdata/zarr_blosc_bitshuffle")
	if err != nil {
		t.Fatalf("ReadDataArrayZarr (bitshuffle) : %v", err)
	}
	d := da.Data()
	if len(d) != 256 {
		t.Fatalf("taille = %d, attendu 256", len(d))
	}
	for i, x := range d {
		if x != float64(i%16) {
			t.Fatalf("v[%d] = %v, attendu %d (bitshuffle incorrect)", i, x, i%16)
		}
	}
}

// TestZarrReadBloscMultiblock lit un store Blosc dont le chunk s'étale sur
// plusieurs blocs (dont un dernier bloc partiel, non découpé) : cas où le
// dé-filtrage doit se faire par bloc. v[i] = i % 97 sur 200000 valeurs.
func TestZarrReadBloscMultiblock(t *testing.T) {
	da, err := ReadDataArrayZarr("testdata/zarr_blosc_multiblock")
	if err != nil {
		t.Fatalf("ReadDataArrayZarr (multiblock) : %v", err)
	}
	d := da.Data()
	if len(d) != 200000 {
		t.Fatalf("taille = %d, attendu 200000", len(d))
	}
	for i, x := range d {
		if x != float64(i%97) {
			t.Fatalf("v[%d] = %v, attendu %d (décodage multi-blocs incorrect)", i, x, i%97)
		}
	}
}

// TestZarrReadBloscLZ4 lit un store Zarr réel produit par zarr-python avec le
// compresseur par défaut (Blosc/LZ4 + byte-shuffle). Fixture dans testdata/ :
// v[i] = i % 16 sur une grille 16×16. Valide le décodeur Blosc pur Go
// (conteneur, découpage en sous-flux, LZ4, sous-flux non compressé, unshuffle).
func TestZarrReadBloscLZ4(t *testing.T) {
	ds, err := ReadDatasetZarr("testdata/zarr_blosc_lz4")
	if err != nil {
		t.Fatalf("ReadDatasetZarr (Blosc) : %v", err)
	}
	v, err := ds.Get("v")
	if err != nil {
		t.Fatal(err)
	}
	d := v.Data()
	if len(d) != 256 {
		t.Fatalf("taille = %d, attendu 256", len(d))
	}
	for i, x := range d {
		if x != float64(i%16) {
			t.Fatalf("v[%d] = %v, attendu %d (décodage Blosc/LZ4 incorrect)", i, x, i%16)
		}
	}
}

// TestZarrFillValueNull vérifie que .zarray écrit fill_value: null et non 0.
// Un fill_value numérique est interprété par xarray/zarr-python comme _FillValue
// et masque les valeurs égales (les 0 légitimes deviendraient NaN à la lecture).
func TestZarrFillValueNull(t *testing.T) {
	// Données contenant explicitement des zéros.
	da, _ := NewDataArray([]string{"y", "x"}, []int{2, 2},
		[]float64{0, 1, 0, 2},
		map[string][]float64{"y": {0, 1}, "x": {0, 1}}, "v")
	dir := filepath.Join(t.TempDir(), "f.zarr")
	if err := WriteDataArrayZarr(dir, da, []int{2, 2}, ZarrNone); err != nil {
		t.Fatalf("WriteDataArrayZarr : %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, ".zarray"))
	if err != nil {
		t.Fatalf("lecture .zarray : %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("json .zarray : %v", err)
	}
	if v, present := m["fill_value"]; !present || v != nil {
		t.Errorf("fill_value = %v (présent=%v), attendu null", v, present)
	}
	// L'aller-retour interne doit préserver les zéros.
	got, err := ReadDataArrayZarr(dir)
	if err != nil {
		t.Fatalf("ReadDataArrayZarr : %v", err)
	}
	if !reflect.DeepEqual(got.Data(), []float64{0, 1, 0, 2}) {
		t.Errorf("Data = %v, attendu [0 1 0 2]", got.Data())
	}
}

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
		ZarrFormat: 2, Shape: []int{2}, Chunks: []int{2}, Dtype: "<c16", Order: "C",
	})
	writeJSONFile(filepath.Join(dir, ".zattrs"), zattrsMeta{Dims: []string{"x"}})
	if _, err := ReadDataArrayZarr(dir); err == nil {
		t.Error("erreur attendue : dtype non supporté (complexe)")
	}
}
