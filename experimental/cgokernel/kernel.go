// Package cgokernel est une DÉMONSTRATION isolée : appeler un noyau C
// vectorisé depuis Go via cgo, pour répondre à la question « peut-on réutiliser
// du C (comme celui de NumPy) ? ».
//
// Ce paquet est volontairement à l'écart du projet principal (qui reste 100 %
// Go pur). Il nécessite CGO_ENABLED=1 et un compilateur C.
package cgokernel

// #cgo CFLAGS: -O3 -mavx2
// #include <stddef.h>
//
// // Boucle simple : gcc -O3 -mavx2 l'auto-vectorise en AVX (comme NumPy).
// static void add_f64(double* restrict dst, const double* restrict a,
//                     const double* restrict b, size_t n) {
//     for (size_t i = 0; i < n; i++) dst[i] = a[i] + b[i];
// }
//
// static void call_add(double* dst, double* a, double* b, size_t n) {
//     add_f64(dst, a, b, n);
// }
import "C"
import "unsafe"

// AddF64 additionne a et b dans dst via le noyau C vectorisé.
// Précondition : len(dst) <= len(a) et len(dst) <= len(b), len(dst) > 0.
func AddF64(dst, a, b []float64) {
	n := len(dst)
	if n == 0 {
		return
	}
	C.call_add(
		(*C.double)(unsafe.Pointer(&dst[0])),
		(*C.double)(unsafe.Pointer(&a[0])),
		(*C.double)(unsafe.Pointer(&b[0])),
		C.size_t(n),
	)
}

// AddF64Go est la version Go pure, pour comparaison directe dans le même paquet.
func AddF64Go(dst, a, b []float64) {
	for i := 0; i < len(dst); i++ {
		dst[i] = a[i] + b[i]
	}
}
