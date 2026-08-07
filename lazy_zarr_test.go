package xarray

import (
	"path/filepath"
	"testing"
)

func TestLazyZarr2D(t *testing.T) {
	// Écrit un store Zarr 2D chunké, puis le traite en lazy (out-of-core).
	R, C := 20, 4
	data := make([]float64, R*C)
	var expected float64
	for i := range data {
		data[i] = float64(i)
		expected += float64(i)
	}
	da, _ := NewDataArray([]string{"t", "x"}, []int{R, C}, data,
		map[string][]float64{"x": {10, 20, 30, 40}}, "v")

	dir := filepath.Join(t.TempDir(), "big.zarr")
	// Chunks 7×3 (non alignés) + compression zlib, pour exercer readBlock.
	if err := WriteDataArrayZarr(dir, da, []int{7, 3}, ZarrZlib); err != nil {
		t.Fatalf("WriteDataArrayZarr : %v", err)
	}

	lz, err := ChunkZarr(dir, 5) // blocs lazy de 5 lignes
	if err != nil {
		t.Fatalf("ChunkZarr : %v", err)
	}
	if lz.NumChunks() != 4 {
		t.Errorf("NumChunks = %d, attendu 4", lz.NumChunks())
	}

	// Réduction en streaming : ne charge que ~1 bloc à la fois.
	s, err := lz.Sum()
	if err != nil {
		t.Fatalf("Sum : %v", err)
	}
	if s != expected {
		t.Errorf("Sum = %v, attendu %v", s, expected)
	}

	// Compute doit reconstruire exactement le tableau (transformé).
	res, err := lz.AddScalar(1).Compute()
	if err != nil {
		t.Fatalf("Compute : %v", err)
	}
	if !floatsEqual(res.Data(), addScalar(data, 1)) {
		t.Errorf("Compute out-of-core Zarr incorrect")
	}
	// Coordonnées lues depuis .zattrs.
	if c, _ := res.Coord("x"); !floatsEqual(c, []float64{10, 20, 30, 40}) {
		t.Errorf("coord x = %v", c)
	}
}

func TestLazyZarr1D(t *testing.T) {
	n := 13
	data := make([]float64, n)
	for i := range data {
		data[i] = float64(i * i)
	}
	da, _ := NewDataArray([]string{"t"}, []int{n}, data, nil, "v")
	dir := filepath.Join(t.TempDir(), "v.zarr")
	if err := WriteDataArrayZarr(dir, da, []int{4}, ZarrNone); err != nil {
		t.Fatalf("Write : %v", err)
	}
	lz, _ := ChunkZarr(dir, 5)
	res, _ := lz.Compute()
	if !floatsEqual(res.Data(), data) {
		t.Errorf("Compute 1D = %v, attendu %v", res.Data(), data)
	}
}

func floatsEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func addScalar(s []float64, v float64) []float64 {
	out := make([]float64, len(s))
	for i, x := range s {
		out[i] = x + v
	}
	return out
}
