package xarray

// binOp identifie une opération arithmétique binaire, pour spécialiser les
// noyaux float64 sans passer par une closure func(T,T) T (non inlinée).
type binOp int

const (
	opAdd binOp = iota
	opSub
	opMul
	opDiv
)

// broadcastFloat64 effectue un broadcasting par nom spécialisé float64 : la
// sélection de l'opération se fait par un switch (parfaitement prédit car
// constant sur toute la boucle), et non par un appel de closure par élément.
func broadcastFloat64(a, b *Variable[float64], op binOp) (*Variable[float64], error) {
	resDims, resShape, aSt, bSt, err := broadcastLayout(a, b)
	if err != nil {
		return nil, err
	}
	out := &Variable[float64]{
		dims:  resDims,
		shape: resShape,
		data:  make([]float64, product(resShape)),
		attrs: map[string]string{},
	}
	resStrides := out.strides()
	parallelFill(len(out.data), func(lo, hi int) {
		fillBinaryRangeF64(out.data, a.data, b.data, op, aSt, bSt, resShape, resStrides, lo, hi)
	})
	return out, nil
}

// fillBinaryRangeF64 remplit out[lo:hi] par itération incrémentale (comme
// fillBinaryRange) mais avec une opération float64 sélectionnée par switch.
func fillBinaryRangeF64(out, adata, bdata []float64, op binOp, aSt, bSt, resShape, resStrides []int, lo, hi int) {
	n := len(resShape)
	counter := make([]int, n)
	flatA, flatB := 0, 0
	rem := lo
	for i := 0; i < n; i++ {
		counter[i] = rem / resStrides[i]
		rem %= resStrides[i]
		flatA += counter[i] * aSt[i]
		flatB += counter[i] * bSt[i]
	}

	for flat := lo; flat < hi; flat++ {
		xa, xb := adata[flatA], bdata[flatB]
		switch op {
		case opAdd:
			out[flat] = xa + xb
		case opSub:
			out[flat] = xa - xb
		case opMul:
			out[flat] = xa * xb
		case opDiv:
			out[flat] = xa / xb
		}
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
}
