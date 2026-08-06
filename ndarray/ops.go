package ndarray

import "fmt"

// --- Arithmétique élément par élément ---------------------------------------
//
// Chemin rapide (même forme) : boucle directe spécialisée, sans closure, que le
// compilateur Go optimise au mieux. Sinon : broadcasting positionnel façon NumPy.

// Add renvoie a + b.
func (a *NDArray) Add(b *NDArray) (*NDArray, error) {
	if sameShape(a, b) {
		out := &NDArray{data: make([]float64, len(a.data)), shape: a.Shape(), strides: cStrides(a.shape)}
		x, y, d := a.data, b.data, out.data
		for i := range d {
			d[i] = x[i] + y[i]
		}
		return out, nil
	}
	return broadcastBinary(a, b, func(x, y float64) float64 { return x + y })
}

// Sub renvoie a - b.
func (a *NDArray) Sub(b *NDArray) (*NDArray, error) {
	if sameShape(a, b) {
		out := &NDArray{data: make([]float64, len(a.data)), shape: a.Shape(), strides: cStrides(a.shape)}
		x, y, d := a.data, b.data, out.data
		for i := range d {
			d[i] = x[i] - y[i]
		}
		return out, nil
	}
	return broadcastBinary(a, b, func(x, y float64) float64 { return x - y })
}

// Mul renvoie a * b.
func (a *NDArray) Mul(b *NDArray) (*NDArray, error) {
	if sameShape(a, b) {
		out := &NDArray{data: make([]float64, len(a.data)), shape: a.Shape(), strides: cStrides(a.shape)}
		x, y, d := a.data, b.data, out.data
		for i := range d {
			d[i] = x[i] * y[i]
		}
		return out, nil
	}
	return broadcastBinary(a, b, func(x, y float64) float64 { return x * y })
}

// Div renvoie a / b.
func (a *NDArray) Div(b *NDArray) (*NDArray, error) {
	if sameShape(a, b) {
		out := &NDArray{data: make([]float64, len(a.data)), shape: a.Shape(), strides: cStrides(a.shape)}
		x, y, d := a.data, b.data, out.data
		for i := range d {
			d[i] = x[i] / y[i]
		}
		return out, nil
	}
	return broadcastBinary(a, b, func(x, y float64) float64 { return x / y })
}

// broadcastPrep calcule la forme résultante et les strides effectifs (0 quand la
// dimension est diffusée) pour un broadcasting positionnel aligné à droite.
func broadcastPrep(a, b *NDArray) (resShape, aSt, bSt []int, err error) {
	na, nb := len(a.shape), len(b.shape)
	n := na
	if nb > n {
		n = nb
	}
	resShape = make([]int, n)
	aSt = make([]int, n)
	bSt = make([]int, n)
	for i := 0; i < n; i++ {
		ai, bi := na-1-i, nb-1-i
		ra, sa := 1, 0
		if ai >= 0 {
			ra, sa = a.shape[ai], a.strides[ai]
		}
		rb, sb := 1, 0
		if bi >= 0 {
			rb, sb = b.shape[bi], b.strides[bi]
		}
		if ra != rb && ra != 1 && rb != 1 {
			return nil, nil, nil, fmt.Errorf("ndarray: formes non diffusables sur l'axe %d (%d vs %d)", i, ra, rb)
		}
		dim := ra
		if rb > dim {
			dim = rb
		}
		if ra == 1 {
			sa = 0
		}
		if rb == 1 {
			sb = 0
		}
		j := n - 1 - i
		resShape[j], aSt[j], bSt[j] = dim, sa, sb
	}
	return resShape, aSt, bSt, nil
}

// broadcastBinary applique op avec broadcasting positionnel. Itération
// incrémentale (flatA/flatB maintenus par pas) pour éviter le recalcul d'indice.
func broadcastBinary(a, b *NDArray, op func(x, y float64) float64) (*NDArray, error) {
	resShape, aSt, bSt, err := broadcastPrep(a, b)
	if err != nil {
		return nil, err
	}
	out := &NDArray{data: make([]float64, product(resShape)), shape: resShape, strides: cStrides(resShape)}
	n := len(resShape)
	counter := make([]int, n)
	flatA, flatB := 0, 0
	for i := range out.data {
		out.data[i] = op(a.data[flatA], b.data[flatB])
		for k := n - 1; k >= 0; k-- {
			counter[k]++
			flatA += aSt[k]
			flatB += bSt[k]
			if counter[k] < resShape[k] {
				break
			}
			counter[k] = 0
			flatA -= aSt[k] * resShape[k]
			flatB -= bSt[k] * resShape[k]
		}
	}
	return out, nil
}

// --- Opérations scalaires ---------------------------------------------------

// AddScalar renvoie a + s.
func (a *NDArray) AddScalar(s float64) *NDArray {
	out := &NDArray{data: make([]float64, len(a.data)), shape: a.Shape(), strides: cStrides(a.shape)}
	for i, x := range a.data {
		out.data[i] = x + s
	}
	return out
}

// MulScalar renvoie a * s.
func (a *NDArray) MulScalar(s float64) *NDArray {
	out := &NDArray{data: make([]float64, len(a.data)), shape: a.Shape(), strides: cStrides(a.shape)}
	for i, x := range a.data {
		out.data[i] = x * s
	}
	return out
}

// --- Réductions par axe -----------------------------------------------------

func (a *NDArray) reduceAxis(axis int, reducer func([]float64) float64) (*NDArray, error) {
	if axis < 0 || axis >= len(a.shape) {
		return nil, fmt.Errorf("ndarray: axe %d hors bornes [0,%d)", axis, len(a.shape))
	}
	newShape := make([]int, 0, len(a.shape)-1)
	for i, s := range a.shape {
		if i == axis {
			continue
		}
		newShape = append(newShape, s)
	}
	out := &NDArray{data: make([]float64, product(newShape)), shape: newShape, strides: cStrides(newShape)}
	nAxis := a.shape[axis]
	stAxis := a.strides[axis]

	counter := make([]int, len(newShape))
	buf := make([]float64, nAxis)
	for flatOut := range out.data {
		base := 0
		j := 0
		for i := range a.shape {
			if i == axis {
				continue
			}
			base += counter[j] * a.strides[i]
			j++
		}
		for k := 0; k < nAxis; k++ {
			buf[k] = a.data[base+k*stAxis]
		}
		out.data[flatOut] = reducer(buf)
		for k := len(newShape) - 1; k >= 0; k-- {
			counter[k]++
			if counter[k] < newShape[k] {
				break
			}
			counter[k] = 0
		}
	}
	return out, nil
}

// SumAxis réduit l'axe donné par somme.
func (a *NDArray) SumAxis(axis int) (*NDArray, error) {
	return a.reduceAxis(axis, func(s []float64) float64 {
		var t float64
		for _, x := range s {
			t += x
		}
		return t
	})
}

// MeanAxis réduit l'axe donné par moyenne.
func (a *NDArray) MeanAxis(axis int) (*NDArray, error) {
	return a.reduceAxis(axis, func(s []float64) float64 {
		if len(s) == 0 {
			return 0
		}
		var t float64
		for _, x := range s {
			t += x
		}
		return t / float64(len(s))
	})
}
