package xarray

import "math"

// Fonctions universelles supplémentaires (arrondis, signe, trigonométrie) et
// opérations binaires élément par élément.

// Round arrondit chaque élément à l'entier le plus proche.
func (da *DataArray[T]) Round() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Round(float64(x))) })
}

// Floor renvoie le plancher de chaque élément.
func (da *DataArray[T]) Floor() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Floor(float64(x))) })
}

// Ceil renvoie le plafond de chaque élément.
func (da *DataArray[T]) Ceil() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Ceil(float64(x))) })
}

// Sign renvoie -1, 0 ou 1 selon le signe de chaque élément.
func (da *DataArray[T]) Sign() *DataArray[T] {
	var one T = 1
	var zero T
	return da.Apply(func(x T) T {
		switch {
		case x > 0:
			return one
		case x < 0:
			return -one // valeur runtime (compatible avec les types non signés)
		default:
			return zero
		}
	})
}

// Sin renvoie le sinus élément par élément.
func (da *DataArray[T]) Sin() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Sin(float64(x))) })
}

// Cos renvoie le cosinus élément par élément.
func (da *DataArray[T]) Cos() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Cos(float64(x))) })
}

// Tanh renvoie la tangente hyperbolique élément par élément.
func (da *DataArray[T]) Tanh() *DataArray[T] {
	return da.Apply(func(x T) T { return T(math.Tanh(float64(x))) })
}

// Maximum renvoie le maximum élément par élément entre da et other (avec
// alignement et broadcasting).
func (da *DataArray[T]) Maximum(other *DataArray[T]) (*DataArray[T], error) {
	return da.binary(other, func(x, y T) T {
		if x > y {
			return x
		}
		return y
	})
}

// Minimum renvoie le minimum élément par élément entre da et other.
func (da *DataArray[T]) Minimum(other *DataArray[T]) (*DataArray[T], error) {
	return da.binary(other, func(x, y T) T {
		if x < y {
			return x
		}
		return y
	})
}
