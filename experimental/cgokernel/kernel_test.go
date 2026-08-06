package cgokernel

import (
	"math"
	"testing"
)

func TestAddF64Correct(t *testing.T) {
	n := 1000
	a := make([]float64, n)
	b := make([]float64, n)
	dst := make([]float64, n)
	for i := range a {
		a[i] = float64(i) * 1.5
		b[i] = float64(i) - 2.0
	}
	AddF64(dst, a, b)
	for i := range dst {
		if math.Abs(dst[i]-(a[i]+b[i])) > 1e-12 {
			t.Fatalf("cgo AddF64 incorrect à %d", i)
		}
	}
}

func donnees(n int) (a, b, dst []float64) {
	a = make([]float64, n)
	b = make([]float64, n)
	dst = make([]float64, n)
	for i := range a {
		a[i] = float64(i)
		b[i] = float64(i)
	}
	return
}

func BenchmarkAddCgo1M(b *testing.B) {
	a, y, dst := donnees(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddF64(dst, a, y)
	}
}

func BenchmarkAddGo1M(b *testing.B) {
	a, y, dst := donnees(1_000_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddF64Go(dst, a, y)
	}
}

func BenchmarkAddCgoSmall(b *testing.B) {
	a, y, dst := donnees(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddF64(dst, a, y)
	}
}

func BenchmarkAddGoSmall(b *testing.B) {
	a, y, dst := donnees(10_000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddF64Go(dst, a, y)
	}
}
