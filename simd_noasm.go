//go:build !amd64 || noasm

package xarray

// Sur les plateformes sans noyau assembleur, addFloat64Vec se replie sur la
// boucle scalaire (identique à addFloat64).

func avxActif() bool { return false }

func addFloat64Vec(dst, x, y []float64) {
	for i := 0; i < len(dst); i++ {
		dst[i] = x[i] + y[i]
	}
}
