//go:build amd64 && !noasm

package xarray

// Noyau AVX (voir simd_amd64.s), conservé comme expérience documentée. La mesure
// montre qu'il ne bat PAS la boucle Go de addFloat64 sur cette opération
// memory-bound (docs/BENCHMARKS.md § SIMD) ; il n'est donc pas utilisé par Add.

//go:noescape
func addFloat64AVX(dst, x, y *float64, n int)

//go:noescape
func cpuHasAVX() bool

var avxAvailable = cpuHasAVX()

func avxActif() bool { return avxAvailable }

// addFloat64Vec additionne via le noyau AVX quand disponible (reste scalaire).
// Utilisé uniquement pour la comparaison de performance (benchmarks/tests).
func addFloat64Vec(dst, x, y []float64) {
	n := len(dst)
	if avxAvailable && n >= 4 {
		m := n &^ 3
		addFloat64AVX(&dst[0], &x[0], &y[0], m)
		for i := m; i < n; i++ {
			dst[i] = x[i] + y[i]
		}
		return
	}
	for i := 0; i < n; i++ {
		dst[i] = x[i] + y[i]
	}
}
