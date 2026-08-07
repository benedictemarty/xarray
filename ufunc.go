package xarray

import "math"

// Fonctions élément par élément (« universal functions »), préservant dimensions
// et coordonnées. Pour les fonctions transcendantes (Sqrt/Exp/Log/Pow), le calcul
// passe par float64 ; pour un type entier, le résultat est tronqué vers T.

// Apply applique une fonction arbitraire à chaque élément.
func (da *DataArray[T]) Apply(fn func(T) T) *DataArray[T] {
	nv := da.variable.mapScalar(fn)
	return &DataArray[T]{variable: nv, coords: da.cloneCoords(), name: da.name}
}

// Abs renvoie la valeur absolue élément par élément.
func (da *DataArray[T]) Abs() *DataArray[T] {
	return da.Apply(func(x T) T {
		if x < 0 {
			return -x
		}
		return x
	})
}

// Clip borne chaque élément à l'intervalle [lo, hi].
func (da *DataArray[T]) Clip(lo, hi T) *DataArray[T] {
	if lo > hi {
		lo, hi = hi, lo
	}
	return da.Apply(func(x T) T {
		if x < lo {
			return lo
		}
		if x > hi {
			return hi
		}
		return x
	})
}

// Sqrt renvoie la racine carrée élément par élément.
func (da *DataArray[T]) Sqrt() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Sqrt(float64(x))) })
}

// Exp renvoie l'exponentielle élément par élément.
func (da *DataArray[T]) Exp() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Exp(float64(x))) })
}

// Log renvoie le logarithme naturel élément par élément.
func (da *DataArray[T]) Log() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Log(float64(x))) })
}

// Pow élève chaque élément à la puissance p.
func (da *DataArray[T]) Pow(p float64) *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Pow(float64(x), p)) })
}
