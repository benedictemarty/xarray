package xarray

import "fmt"

// WhereFunc renvoie une copie où chaque élément ne satisfaisant pas keep est
// remplacé par other. C'est le masquage conditionnel de base (équivalent de
// `da.where(cond, other)` de xarray, la condition étant un prédicat élément par
// élément).
func (da *DataArray[T]) WhereFunc(keep func(T) bool, other T) *DataArray[T] {
	out := da.clone()
	for i, x := range out.variable.data {
		if !keep(x) {
			out.variable.data[i] = other
		}
	}
	return out
}

// Where renvoie une copie où les éléments dont le masque vaut 0 sont remplacés
// par other (le masque est un DataArray de même forme, non-zéro = conserver).
func (da *DataArray[T]) Where(mask *DataArray[T], other T) (*DataArray[T], error) {
	if !sameDimsShape(da.variable, mask.variable) {
		return nil, fmt.Errorf("xarray: le masque %v doit avoir la même forme que %v", mask.Shape(), da.Shape())
	}
	out := da.clone()
	md := mask.variable.data
	for i := range out.variable.data {
		if md[i] == 0 {
			out.variable.data[i] = other
		}
	}
	return out, nil
}

// InterpolateNA remplit les valeurs manquantes (NaN) le long de dim par
// interpolation linéaire entre les valeurs valides encadrantes. L'interpolation
// se fait selon la coordonnée de dim si elle existe, sinon selon la position.
// Les NaN de bord (avant la première ou après la dernière valeur valide) sont
// conservés.
func (da *DataArray[T]) InterpolateNA(dim string) (*DataArray[T], error) {
	axis := da.variable.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	out := da.clone()
	shape := out.variable.shape
	stAxis := out.variable.strides()[axis]
	dimLen := shape[axis]
	data := out.variable.data

	// Positions d'interpolation : coordonnée si disponible, sinon indices.
	pos := make([]float64, dimLen)
	if cv, ok := out.coords[dim]; ok {
		for k := 0; k < dimLen; k++ {
			pos[k] = float64(cv.data[k])
		}
	} else {
		for k := 0; k < dimLen; k++ {
			pos[k] = float64(k)
		}
	}

	out.forEachLine(axis, func(base int) {
		lastValid := -1
		for k := 0; k < dimLen; k++ {
			if isNaNT(data[base+k*stAxis]) {
				continue
			}
			if lastValid >= 0 && lastValid < k-1 {
				v0 := float64(data[base+lastValid*stAxis])
				v1 := float64(data[base+k*stAxis])
				p0, p1 := pos[lastValid], pos[k]
				for m := lastValid + 1; m < k; m++ {
					frac := (pos[m] - p0) / (p1 - p0)
					data[base+m*stAxis] = T(v0 + (v1-v0)*frac)
				}
			}
			lastValid = k
		}
	})
	return out, nil
}
