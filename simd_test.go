package xarray

import (
	"reflect"
	"testing"
)

func TestAddFloat64SIMD(t *testing.T) {
	// Taille volontairement non multiple de 4 pour exercer le reste scalaire.
	n := 4099
	x := make([]float64, n)
	y := make([]float64, n)
	attendu := make([]float64, n)
	for i := range x {
		x[i] = float64(i) * 1.5
		y[i] = float64(i) - 3.25
		attendu[i] = x[i] + y[i]
	}
	dst := make([]float64, n)
	addFloat64(dst, x, y)
	if !reflect.DeepEqual(dst, attendu) {
		t.Fatalf("addFloat64 (boucle) incorrect")
	}
	// Le noyau vectoriel doit produire le même résultat.
	dstVec := make([]float64, n)
	addFloat64Vec(dstVec, x, y)
	if !reflect.DeepEqual(dstVec, attendu) {
		t.Fatalf("addFloat64Vec incorrect (AVX actif : %v)", avxActif())
	}
}

func TestAddFloat64Petit(t *testing.T) {
	// Moins de 4 éléments : chemin scalaire même si AVX présent.
	dst := make([]float64, 3)
	addFloat64(dst, []float64{1, 2, 3}, []float64{10, 20, 30})
	if !reflect.DeepEqual(dst, []float64{11, 22, 33}) {
		t.Errorf("addFloat64 petit = %v", dst)
	}
}

func TestAddViaSIMDEtGenerique(t *testing.T) {
	// Add sur float64 mêmes formes passe par le noyau SIMD ; le résultat doit
	// être identique au chemin générique.
	a, _ := NewDataArray([]string{"x"}, []int{5}, []float64{1, 2, 3, 4, 5},
		map[string][]float64{"x": {0, 1, 2, 3, 4}}, "a")
	b, _ := NewDataArray([]string{"x"}, []int{5}, []float64{10, 20, 30, 40, 50},
		map[string][]float64{"x": {0, 1, 2, 3, 4}}, "b")
	viaSIMD, _ := a.Add(b)
	viaGen, _ := a.binary(b, func(x, y float64) float64 { return x + y })
	if !reflect.DeepEqual(viaSIMD.Data(), viaGen.Data()) {
		t.Errorf("SIMD %v != générique %v", viaSIMD.Data(), viaGen.Data())
	}
	if !reflect.DeepEqual(viaSIMD.Data(), []float64{11, 22, 33, 44, 55}) {
		t.Errorf("résultat = %v", viaSIMD.Data())
	}
}

func BenchmarkAddFloat64Kernel(b *testing.B) {
	// Noyau isolé, taille tenant en cache (8192 float64 = 64 Ko) : met en
	// évidence le gain SIMD compute-bound, sans l'overhead d'alignement/copie.
	n := 8192
	x := make([]float64, n)
	y := make([]float64, n)
	dst := make([]float64, n)
	for i := range x {
		x[i] = float64(i)
		y[i] = float64(i)
	}
	b.SetBytes(int64(n * 8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addFloat64(dst, x, y)
	}
}

func BenchmarkAddFloat64KernelVec(b *testing.B) {
	// Même noyau via le chemin vectoriel (AVX si dispo) — pour comparaison.
	n := 8192
	x := make([]float64, n)
	y := make([]float64, n)
	dst := make([]float64, n)
	for i := range x {
		x[i] = float64(i)
		y[i] = float64(i)
	}
	b.SetBytes(int64(n * 8))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addFloat64Vec(dst, x, y)
	}
}

func BenchmarkAddLargeSameShape(b *testing.B) {
	// 1000x1000 mêmes coordonnées : chemin SIMD élément par élément.
	n := 1000
	data := make([]float64, n*n)
	coord := make([]float64, n)
	for i := range coord {
		coord[i] = float64(i)
	}
	da, _ := NewDataArray([]string{"x", "y"}, []int{n, n}, data,
		map[string][]float64{"x": coord, "y": coord}, "v")
	db, _ := NewDataArray([]string{"x", "y"}, []int{n, n}, data,
		map[string][]float64{"x": coord, "y": coord}, "v")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := da.Add(db); err != nil {
			b.Fatal(err)
		}
	}
}
