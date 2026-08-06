package xarray

// addFloat64 additionne x et y élément par élément dans dst.
//
// Point clé de performance : cette boucle « nue » (sans closure) est inlinée et
// très bien optimisée par le compilateur Go — elle atteint la bande passante
// mémoire. Elle est nettement plus rapide que le chemin générique de binaryOp,
// dont la closure func(T,T) T n'est PAS inlinée (un appel par élément), et même
// plus rapide que notre noyau AVX manuel (voir docs/BENCHMARKS.md § SIMD).
//
// Précondition : len(dst) <= len(x) et len(dst) <= len(y).
func addFloat64(dst, x, y []float64) {
	for i := 0; i < len(dst); i++ {
		dst[i] = x[i] + y[i]
	}
}

// Noyaux spécialisés float64 (sans closure) pour les autres opérations, dans le
// même esprit : boucles nues optimisées par le compilateur.

func subFloat64(dst, x, y []float64) {
	for i := 0; i < len(dst); i++ {
		dst[i] = x[i] - y[i]
	}
}

func mulFloat64(dst, x, y []float64) {
	for i := 0; i < len(dst); i++ {
		dst[i] = x[i] * y[i]
	}
}

func divFloat64(dst, x, y []float64) {
	for i := 0; i < len(dst); i++ {
		dst[i] = x[i] / y[i]
	}
}
