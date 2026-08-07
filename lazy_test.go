package xarray

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestLazyComputeMemoire(t *testing.T) {
	// 4×2, chunk de 2 lignes -> 2 chunks.
	da, _ := NewDataArray([]string{"t", "x"}, []int{4, 2},
		[]float64{1, 2, 3, 4, 5, 6, 7, 8},
		map[string][]float64{"t": {0, 1, 2, 3}, "x": {10, 20}}, "v")
	lz, err := Chunk(da, 2)
	if err != nil {
		t.Fatalf("Chunk : %v", err)
	}
	if lz.NumChunks() != 2 {
		t.Errorf("NumChunks = %d, attendu 2", lz.NumChunks())
	}
	// (x*2)+1 différé
	res, err := lz.MulScalar(2).AddScalar(1).Compute()
	if err != nil {
		t.Fatalf("Compute : %v", err)
	}
	attendu := []float64{3, 5, 7, 9, 11, 13, 15, 17}
	if !reflect.DeepEqual(res.Data(), attendu) {
		t.Errorf("Compute = %v, attendu %v", res.Data(), attendu)
	}
	// Coordonnées préservées.
	if c, _ := res.Coord("x"); !reflect.DeepEqual(c, []float64{10, 20}) {
		t.Errorf("coord x = %v", c)
	}
}

func TestLazyReductions(t *testing.T) {
	da, _ := NewDataArray([]string{"t"}, []int{6}, []float64{1, 2, 3, 4, 5, 6}, nil, "v")
	lz, _ := Chunk(da, 4) // 2 chunks (4 + 2)
	s, _ := lz.Sum()
	if s != 21 {
		t.Errorf("Sum = %v, attendu 21", s)
	}
	m, _ := lz.Mean()
	if m != 3.5 {
		t.Errorf("Mean = %v, attendu 3.5", m)
	}
	mn, _ := lz.Min()
	mx, _ := lz.Max()
	if mn != 1 || mx != 6 {
		t.Errorf("Min/Max = %v/%v", mn, mx)
	}
	// Réduction après transformation différée.
	s2, _ := lz.AddScalar(10).Sum()
	if s2 != 21+60 {
		t.Errorf("Sum(+10) = %v, attendu 81", s2)
	}
}

func TestLazyManyChunks(t *testing.T) {
	// Beaucoup de chunks pour exercer le parallélisme.
	n := 1000
	data := make([]float64, n)
	for i := range data {
		data[i] = float64(i)
	}
	da, _ := NewDataArray([]string{"t"}, []int{n}, data, nil, "v")
	lz, _ := Chunk(da, 7) // chunks non divisibles
	res, _ := lz.MulScalar(2).Compute()
	for i := 0; i < n; i++ {
		if res.Data()[i] != float64(i)*2 {
			t.Fatalf("Compute parallèle incorrect à %d : %v", i, res.Data()[i])
		}
	}
	s, _ := lz.Sum()
	if s != float64(n*(n-1)/2) {
		t.Errorf("Sum = %v", s)
	}
}

func TestLazyFileOutOfCore(t *testing.T) {
	// Écrit un tableau sur disque, puis le traite bloc par bloc (hors-mémoire).
	n := 100
	data := make([]float64, n*3)
	for i := range data {
		data[i] = float64(i)
	}
	path := filepath.Join(t.TempDir(), "big.f64")
	if err := WriteRawF64(path, data); err != nil {
		t.Fatalf("WriteRawF64 : %v", err)
	}
	lz, err := ChunkFile(path, []string{"t", "x"}, []int{n, 3}, nil, 10)
	if err != nil {
		t.Fatalf("ChunkFile : %v", err)
	}
	if lz.NumChunks() != 10 {
		t.Errorf("NumChunks = %d, attendu 10", lz.NumChunks())
	}
	// Somme de 0..299 = 299*300/2 = 44850
	s, err := lz.Sum()
	if err != nil {
		t.Fatalf("Sum : %v", err)
	}
	if s != 44850 {
		t.Errorf("Sum out-of-core = %v, attendu 44850", s)
	}
	// Compute matérialise le résultat transformé.
	res, _ := lz.AddScalar(1).Compute()
	if res.Data()[0] != 1 || res.Data()[299] != 300 {
		t.Errorf("Compute out-of-core : bornes = %v %v", res.Data()[0], res.Data()[299])
	}
}
