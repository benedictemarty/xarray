// Package ml fournit quelques algorithmes de machine learning classiques
// construits sur le moteur ndarray.
//
// Portée : ML « classique » et pédagogique (régression linéaire, k-means,
// standardisation). Ce n'est PAS un framework d'apprentissage profond (pas
// d'autograd, pas de GPU) — pour cela, Python (PyTorch/TF) reste la référence,
// et Go sert surtout à l'inférence en production.
package ml

import (
	"fmt"
	"math"

	nd "github.com/benedictemarty/xarray/ndarray"
)

// Standardize centre-réduit chaque colonne d'une matrice de caractéristiques
// (m échantillons × n variables) : (x - moyenne) / écart-type. Renvoie la
// matrice transformée ainsi que les moyennes et écarts-types par colonne (pour
// appliquer la même transformation à d'autres données).
func Standardize(x *nd.NDArray) (out, mean, std *nd.NDArray, err error) {
	sh := x.Shape()
	if len(sh) != 2 {
		return nil, nil, nil, fmt.Errorf("ml: Standardize attend une matrice 2D (%dD)", len(sh))
	}
	m, n := sh[0], sh[1]
	data := x.Data()

	meanD := make([]float64, n)
	for j := 0; j < n; j++ {
		var s float64
		for i := 0; i < m; i++ {
			s += data[i*n+j]
		}
		meanD[j] = s / float64(m)
	}
	stdD := make([]float64, n)
	for j := 0; j < n; j++ {
		var s float64
		for i := 0; i < m; i++ {
			d := data[i*n+j] - meanD[j]
			s += d * d
		}
		stdD[j] = math.Sqrt(s / float64(m))
	}
	o := make([]float64, m*n)
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			s := stdD[j]
			if s == 0 {
				s = 1 // colonne constante : évite la division par zéro
			}
			o[i*n+j] = (data[i*n+j] - meanD[j]) / s
		}
	}
	out, _ = nd.New([]int{m, n}, o)
	mean, _ = nd.New([]int{n}, meanD)
	std, _ = nd.New([]int{n}, stdD)
	return out, mean, std, nil
}
