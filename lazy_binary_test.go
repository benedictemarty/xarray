package xarray

import (
	"path/filepath"
	"testing"
)

func TestLazyBinaryMemoire(t *testing.T) {
	a, _ := NewDataArray([]string{"t"}, []int{6}, []float64{1, 2, 3, 4, 5, 6}, nil, "a")
	b, _ := NewDataArray([]string{"t"}, []int{6}, []float64{10, 10, 10, 10, 10, 10}, nil, "b")
	la, _ := Chunk(a, 4)
	lb, _ := Chunk(b, 4)

	sum, err := la.Add(lb)
	if err != nil {
		t.Fatalf("Add : %v", err)
	}
	res, _ := sum.Compute()
	if !floatsEqual(res.Data(), []float64{11, 12, 13, 14, 15, 16}) {
		t.Errorf("a+b = %v", res.Data())
	}

	// Graphe composé : (a*2 - b) chunk par chunk.
	expr, _ := la.MulScalar(2).Sub(lb)
	r2, _ := expr.Compute()
	// [2-10, 4-10, 6-10, 8-10, 10-10, 12-10] = [-8 -6 -4 -2 0 2]
	if !floatsEqual(r2.Data(), []float64{-8, -6, -4, -2, 0, 2}) {
		t.Errorf("a*2-b = %v", r2.Data())
	}

	// Réduction sur un tableau combiné.
	s, _ := sum.Sum()
	if s != 11+12+13+14+15+16 {
		t.Errorf("Sum(a+b) = %v", s)
	}
}

func TestLazyBinaryIncompatible(t *testing.T) {
	a, _ := NewDataArray([]string{"t"}, []int{4}, []float64{1, 2, 3, 4}, nil, "a")
	b, _ := NewDataArray([]string{"t"}, []int{6}, []float64{1, 2, 3, 4, 5, 6}, nil, "b")
	la, _ := Chunk(a, 2)
	lb, _ := Chunk(b, 2)
	if _, err := la.Add(lb); err == nil {
		t.Error("erreur attendue : formes incompatibles")
	}
}

func TestLazyBinaryOutOfCore(t *testing.T) {
	// Deux tableaux sur disque combinés sans être chargés entièrement.
	n := 50
	da := make([]float64, n)
	db := make([]float64, n)
	for i := range da {
		da[i] = float64(i)
		db[i] = float64(i * 2)
	}
	dir := t.TempDir()
	pa := filepath.Join(dir, "a.f64")
	pb := filepath.Join(dir, "b.f64")
	WriteRawF64(pa, da)
	WriteRawF64(pb, db)

	la, _ := ChunkFile(pa, []string{"t"}, []int{n}, nil, 8)
	lb, _ := ChunkFile(pb, []string{"t"}, []int{n}, nil, 8)
	prod, err := la.Mul(lb)
	if err != nil {
		t.Fatalf("Mul : %v", err)
	}
	// somme des i*(2i) = 2*somme(i^2)
	var expected float64
	for i := 0; i < n; i++ {
		expected += float64(i) * float64(i*2)
	}
	s, _ := prod.Sum()
	if s != expected {
		t.Errorf("Sum(a*b) out-of-core = %v, attendu %v", s, expected)
	}
}
