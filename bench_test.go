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

func BenchmarkMulSameShape(b *testing.B) {
	a := grille2D(1000) // 1 M, mêmes coordonnées -> chemin rapide float64
	c := grille2D(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.Mul(c); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSubSameShape(b *testing.B) {
	a := grille2D(1000)
	c := grille2D(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.Sub(c); err != nil {
			b.Fatal(err)
		}
	}
}

// Référence : chemin générique (closure) pour mesurer l'apport du noyau direct.
func BenchmarkMulGeneric1M(b *testing.B) {
	a := grille2D(1000)
	c := grille2D(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a.binary(c, func(x, y float64) float64 { return x * y }); err != nil {
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

func BenchmarkBroadcastLarge(b *testing.B) {
	// 1000 x 1000 = 1 M éléments : au-delà du seuil de parallélisation.
	x, _ := NewDataArray([]string{"x"}, []int{1000}, make([]float64, 1000), nil, "x")
	y, _ := NewDataArray([]string{"y"}, []int{1000}, make([]float64, 1000), nil, "y")
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

func BenchmarkGroupBySum(b *testing.B) {
	// dim t (1000) avec 10 groupes répétés, dim x (10).
	n, groupes := 1000, 10
	data := make([]float64, n*10)
	for i := range data {
		data[i] = float64(i)
	}
	ts := make([]float64, n)
	for i := 0; i < n; i++ {
		ts[i] = float64(i % groupes)
	}
	xs := make([]float64, 10)
	for i := range xs {
		xs[i] = float64(i)
	}
	da, _ := NewDataArray([]string{"t", "x"}, []int{n, 10}, data,
		map[string][]float64{"t": ts, "x": xs}, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g, _ := da.GroupBy("t")
		if _, err := g.Sum(); err != nil {
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
