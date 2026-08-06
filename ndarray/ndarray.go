// Package ndarray fournit un tableau dense N-dimensionnel de float64, façon
// « mini-NumPy », destiné à servir de moteur de calcul.
//
// Contrairement à xarray.Variable (étiqueté, générique), NDArray est spécialisé
// float64 et sans closure sur le chemin chaud : les opérations élément par
// élément sont des boucles directes que le compilateur Go optimise bien. Le
// broadcasting est **positionnel** (aligné à droite, dimensions de taille 1
// étirées), comme NumPy — et non par nom comme xarray.
//
// Portée : sous-ensemble volontairement restreint (arithmétique, réductions,
// broadcasting). Ce n'est PAS un portage complet de NumPy (pas d'algèbre
// linéaire, pas de dtypes multiples, pas de vues/slicing avancé).
package ndarray

import (
	"fmt"
	"math"
	"strings"
)

// NDArray est un tableau dense de float64 en ordre C (row-major).
type NDArray struct {
	data    []float64
	shape   []int
	strides []int
}

func cStrides(shape []int) []int {
	st := make([]int, len(shape))
	acc := 1
	for i := len(shape) - 1; i >= 0; i-- {
		st[i] = acc
		acc *= shape[i]
	}
	return st
}

func product(shape []int) int {
	p := 1
	for _, s := range shape {
		p *= s
	}
	return p
}

// New construit un NDArray à partir d'une forme et de données plates (ordre C).
func New(shape []int, data []float64) (*NDArray, error) {
	for _, s := range shape {
		if s < 0 {
			return nil, fmt.Errorf("ndarray: taille négative %d", s)
		}
	}
	if product(shape) != len(data) {
		return nil, fmt.Errorf("ndarray: %d valeur(s) pour une forme de taille %d", len(data), product(shape))
	}
	return &NDArray{
		data:    append([]float64(nil), data...),
		shape:   append([]int(nil), shape...),
		strides: cStrides(shape),
	}, nil
}

// Zeros crée un tableau de zéros de la forme donnée.
func Zeros(shape ...int) *NDArray {
	sh := append([]int(nil), shape...)
	return &NDArray{data: make([]float64, product(sh)), shape: sh, strides: cStrides(sh)}
}

// Arange crée un tableau 1D [0, 1, …, n-1].
func Arange(n int) *NDArray {
	d := make([]float64, n)
	for i := range d {
		d[i] = float64(i)
	}
	a, _ := New([]int{n}, d)
	return a
}

// Shape renvoie une copie de la forme.
func (a *NDArray) Shape() []int { return append([]int(nil), a.shape...) }

// Ndim renvoie le nombre de dimensions.
func (a *NDArray) Ndim() int { return len(a.shape) }

// Size renvoie le nombre total d'éléments.
func (a *NDArray) Size() int { return len(a.data) }

// Data renvoie une copie des données plates.
func (a *NDArray) Data() []float64 { return append([]float64(nil), a.data...) }

// At renvoie la valeur au multi-indice donné.
func (a *NDArray) At(idx ...int) (float64, error) {
	if len(idx) != len(a.shape) {
		return 0, fmt.Errorf("ndarray: %d indice(s) pour %d dimension(s)", len(idx), len(a.shape))
	}
	flat := 0
	for i, k := range idx {
		if k < 0 || k >= a.shape[i] {
			return 0, fmt.Errorf("ndarray: indice %d hors bornes [0,%d)", k, a.shape[i])
		}
		flat += k * a.strides[i]
	}
	return a.data[flat], nil
}

func sameShape(a, b *NDArray) bool {
	if len(a.shape) != len(b.shape) {
		return false
	}
	for i := range a.shape {
		if a.shape[i] != b.shape[i] {
			return false
		}
	}
	return true
}

// String fournit une représentation compacte.
func (a *NDArray) String() string {
	parts := make([]string, len(a.shape))
	for i, s := range a.shape {
		parts[i] = fmt.Sprintf("%d", s)
	}
	return fmt.Sprintf("NDArray(shape=[%s]) %v", strings.Join(parts, "×"), a.data)
}

// Sum renvoie la somme de tous les éléments.
func (a *NDArray) Sum() float64 {
	var s float64
	for _, x := range a.data {
		s += x
	}
	return s
}

// Mean renvoie la moyenne (NaN si vide).
func (a *NDArray) Mean() float64 {
	if len(a.data) == 0 {
		return math.NaN()
	}
	return a.Sum() / float64(len(a.data))
}
