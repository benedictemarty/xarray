package xarray

// Masques et comptages de valeurs manquantes (NaN). Pour les types entiers,
// aucune valeur n'est NaN.

// IsNull renvoie un masque : 1 là où la valeur est NaN, 0 sinon.
func (da *DataArray[T]) IsNull() *DataArray[T] {
	var one T = 1
	var zero T
	return da.Apply(func(x T) T {
		if isNaNT(x) {
			return one
		}
		return zero
	})
}

// NotNull renvoie un masque : 1 là où la valeur est présente, 0 si NaN.
func (da *DataArray[T]) NotNull() *DataArray[T] {
	var one T = 1
	var zero T
	return da.Apply(func(x T) T {
		if isNaNT(x) {
			return zero
		}
		return one
	})
}

// Count renvoie le nombre de valeurs présentes (non-NaN).
func (da *DataArray[T]) Count() int {
	n := 0
	for _, x := range da.variable.data {
		if !isNaNT(x) {
			n++
		}
	}
	return n
}

// CountAxis renvoie, le long de dim, le nombre de valeurs présentes (non-NaN).
func (da *DataArray[T]) CountAxis(dim string) (*DataArray[float64], error) {
	return reduceAxisDA[T, float64](da, dim, func(s []T) float64 {
		n := 0
		for _, x := range s {
			if !isNaNT(x) {
				n++
			}
		}
		return float64(n)
	})
}
