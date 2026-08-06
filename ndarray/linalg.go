package ndarray

import "fmt"

// Matmul calcule le produit matriciel de deux tableaux 2D : a (m×k) · b (k×n)
// donne (m×n).
//
// L'implémentation utilise l'ordre de boucle « ikj », qui parcourt les lignes de
// b et de c de façon contiguë (cache-friendly) — l'ordre le plus rapide en Go
// pur. Elle ne rivalise PAS avec un BLAS optimisé (NumPy) : voir le benchmark et
// docs/NDARRAY.md.
func Matmul(a, b *NDArray) (*NDArray, error) {
	if a.Ndim() != 2 || b.Ndim() != 2 {
		return nil, fmt.Errorf("ndarray: Matmul attend deux tableaux 2D (%dD × %dD)", a.Ndim(), b.Ndim())
	}
	m, k := a.shape[0], a.shape[1]
	k2, n := b.shape[0], b.shape[1]
	if k != k2 {
		return nil, fmt.Errorf("ndarray: dimensions internes incompatibles (%d vs %d)", k, k2)
	}
	c := Zeros(m, n)
	A, B, C := a.data, b.data, c.data
	for i := 0; i < m; i++ {
		cRow := C[i*n : i*n+n]
		for p := 0; p < k; p++ {
			aip := A[i*k+p]
			if aip == 0 {
				continue
			}
			bRow := B[p*n : p*n+n]
			for j := 0; j < n; j++ {
				cRow[j] += aip * bRow[j]
			}
		}
	}
	return c, nil
}

// T renvoie la transposée d'un tableau 2D.
func (a *NDArray) T() (*NDArray, error) {
	if a.Ndim() != 2 {
		return nil, fmt.Errorf("ndarray: T attend un tableau 2D (%dD)", a.Ndim())
	}
	m, n := a.shape[0], a.shape[1]
	out := Zeros(n, m)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			out.data[j*m+i] = a.data[i*n+j]
		}
	}
	return out, nil
}

// MatVec calcule le produit matrice-vecteur : a (m×k) · v (k) donne (m).
func MatVec(a, v *NDArray) (*NDArray, error) {
	if a.Ndim() != 2 || v.Ndim() != 1 {
		return nil, fmt.Errorf("ndarray: MatVec attend (2D, 1D), reçu (%dD, %dD)", a.Ndim(), v.Ndim())
	}
	m, k := a.shape[0], a.shape[1]
	if k != v.shape[0] {
		return nil, fmt.Errorf("ndarray: dimensions incompatibles (%d vs %d)", k, v.shape[0])
	}
	out := Zeros(m)
	for i := 0; i < m; i++ {
		var s float64
		aRow := a.data[i*k : i*k+k]
		for p := 0; p < k; p++ {
			s += aRow[p] * v.data[p]
		}
		out.data[i] = s
	}
	return out, nil
}
