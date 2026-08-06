package ndarray

import "fmt"

// Opérations « in-place » : le résultat est écrit dans un tableau dst fourni par
// l'appelant, ce qui évite l'allocation (et la zéro-initialisation) du résultat
// à chaque opération. C'est le principal levier de performance identifié par les
// benchmarks : le coût de Add sur de gros tableaux est dominé par l'allocation,
// pas par le calcul (voir experimental/cgokernel/README.md).
//
// En réutilisant un même dst dans une boucle (ou via un pool), on atteint
// pratiquement le débit de NumPy sur les opérations memory-bound.

func checkSame3(dst, a, b *NDArray) error {
	if !sameShape(a, b) {
		return fmt.Errorf("ndarray: formes différentes entre opérandes %v et %v", a.shape, b.shape)
	}
	if !sameShape(dst, a) {
		return fmt.Errorf("ndarray: forme de destination %v incompatible avec %v", dst.shape, a.shape)
	}
	return nil
}

// AddInto écrit a + b dans dst (mêmes formes). Aucune allocation.
func AddInto(dst, a, b *NDArray) error {
	if err := checkSame3(dst, a, b); err != nil {
		return err
	}
	d, x, y := dst.data, a.data, b.data
	for i := range d {
		d[i] = x[i] + y[i]
	}
	return nil
}

// SubInto écrit a - b dans dst.
func SubInto(dst, a, b *NDArray) error {
	if err := checkSame3(dst, a, b); err != nil {
		return err
	}
	d, x, y := dst.data, a.data, b.data
	for i := range d {
		d[i] = x[i] - y[i]
	}
	return nil
}

// MulInto écrit a * b dans dst.
func MulInto(dst, a, b *NDArray) error {
	if err := checkSame3(dst, a, b); err != nil {
		return err
	}
	d, x, y := dst.data, a.data, b.data
	for i := range d {
		d[i] = x[i] * y[i]
	}
	return nil
}

// DivInto écrit a / b dans dst.
func DivInto(dst, a, b *NDArray) error {
	if err := checkSame3(dst, a, b); err != nil {
		return err
	}
	d, x, y := dst.data, a.data, b.data
	for i := range d {
		d[i] = x[i] / y[i]
	}
	return nil
}

// AddInPlace écrit a + b directement dans a (a est modifié). b doit avoir la
// même forme. Utile pour accumuler sans aucune allocation.
func (a *NDArray) AddInPlace(b *NDArray) error {
	if !sameShape(a, b) {
		return fmt.Errorf("ndarray: formes différentes %v et %v", a.shape, b.shape)
	}
	for i := range a.data {
		a.data[i] += b.data[i]
	}
	return nil
}

// EmptyLike renvoie un tableau non initialisé (contenu quelconque) de la même
// forme que a, réutilisable comme destination. Contrairement à Zeros, il ne
// garantit pas un contenu nul — à n'utiliser que si l'on écrit intégralement le
// résultat ensuite (ex. via *Into). (En Go, make zéro-initialise tout de même ;
// cette fonction existe surtout pour l'intention et un éventuel pooling.)
func EmptyLike(a *NDArray) *NDArray {
	return &NDArray{
		data:    make([]float64, len(a.data)),
		shape:   a.Shape(),
		strides: cStrides(a.shape),
	}
}
