package ndarray

import "testing"

func grille(n int) *NDArray {
	d := make([]float64, n*n)
	for i := range d {
		d[i] = float64(i)
	}
	a, _ := New([]int{n, n}, d)
	return a
}

func BenchmarkAddSameShape(b *testing.B) {
	a := grille(1000) // 1 M éléments
	c := grille(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.Add(c); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAddSmall(b *testing.B) {
	a := grille(100)
	c := grille(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.Add(c); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAddInto1M(b *testing.B) {
	// In-place : destination réutilisée -> aucune allocation par opération.
	a := grille(1000)
	c := grille(1000)
	dst := Zeros(1000, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := AddInto(dst, a, c); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSumAxis(b *testing.B) {
	a := grille(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.SumAxis(0); err != nil {
			b.Fatal(err)
		}
	}
}
