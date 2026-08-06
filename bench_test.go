package xarray

import (
	"bytes"
	"testing"
)

// grille2D construit un DataArray[float64] de taille n x n avec coordonnées.
func grille2D(n int) *DataArray[float64] {
	data := make([]float64, n*n)
	for i := range data {
		data[i] = float64(i)
	}
	xs := make([]float64, n)
	ys := make([]float64, n)
	for i := 0; i < n; i++ {
		xs[i] = float64(i)
		ys[i] = float64(i)
	}
	da, _ := NewDataArray([]string{"x", "y"}, []int{n, n}, data,
		map[string][]float64{"x": xs, "y": ys}, "grille")
	return da
}

func BenchmarkAdd(b *testing.B) {
	a := grille2D(100)
	c := grille2D(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.Add(c); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSumAxis(b *testing.B) {
	a := grille2D(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.SumAxis("x"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMeanAxis(b *testing.B) {
	a := grille2D(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.MeanAxis("x"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBroadcast(b *testing.B) {
	// x (200) broadcast avec y (200) -> 200x200
	x, _ := NewDataArray([]string{"x"}, []int{200}, make([]float64, 200), nil, "x")
	y, _ := NewDataArray([]string{"y"}, []int{200}, make([]float64, 200), nil, "y")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := x.Add(y); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClone(b *testing.B) {
	a := grille2D(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.clone()
	}
}

func BenchmarkWriteCSV(b *testing.B) {
	a := grille2D(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := a.WriteCSV(&buf); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOuterJoin(b *testing.B) {
	a := grille2D(100)
	c := grille2D(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.AddJoin(c, JoinOuter, 0); err != nil {
			b.Fatal(err)
		}
	}
}
