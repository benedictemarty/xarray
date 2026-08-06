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

func applyOp(op binOp, x, y float64) float64 {
	switch op {
	case opAdd:
		return x + y
	case opSub:
		return x - y
	case opMul:
		return x * y
	default:
		return x / y
	}
}

// broadcastFloat64 effectue un broadcasting par nom spécialisé float64.
//
// Optimisation clé (itération strided) : au lieu d'un compteur multi-dimensionnel
// mis à jour à chaque élément, on itère **par lignes** (le dernier axe) et on
// exécute une **boucle interne contiguë**. Quand un opérande a un stride nul sur
// l'axe interne (cas typique du broadcasting), sa valeur est un scalaire hissé
// hors de la boucle interne — qui devient alors « scalaire op vecteur contigu »
// ou « vecteur op vecteur », que le compilateur optimise bien.
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
	n := len(resShape)
	if n == 0 { // résultat scalaire (0 dimension)
		out.data[0] = applyOp(op, a.data[0], b.data[0])
		return out, nil
	}

	L := resShape[n-1]
	aInner, bInner := aSt[n-1], bSt[n-1]
	extShape := resShape[:n-1]
	extASt, extBSt := aSt[:n-1], bSt[:n-1]
	extStrides := cStridesInt(extShape)
	numLines := product(extShape) // = len(out.data)/L (ou 1 si L==0)
	if L == 0 {
		return out, nil
	}

	parallelLines(numLines, len(out.data), func(loLine, hiLine int) {
		fillLinesF64(out.data, a.data, b.data, op, L, aInner, bInner, extShape, extASt, extBSt, extStrides, loLine, hiLine)
	})
	return out, nil
}

// cStridesInt calcule les strides en ordre C d'une forme (axes externes).
func cStridesInt(shape []int) []int {
	st := make([]int, len(shape))
	acc := 1
	for i := len(shape) - 1; i >= 0; i-- {
		st[i] = acc
		acc *= shape[i]
	}
	return st
}

// fillLinesF64 traite les lignes [loLine, hiLine) : pour chacune, une boucle
// interne contiguë sur l'axe le plus interne.
func fillLinesF64(out, adata, bdata []float64, op binOp, L, aInner, bInner int, extShape, extASt, extBSt, extStrides []int, loLine, hiLine int) {
	m := len(extShape)
	counter := make([]int, m)
	aBase, bBase := 0, 0
	rem := loLine
	for i := 0; i < m; i++ {
		counter[i] = rem / extStrides[i]
		rem %= extStrides[i]
		aBase += counter[i] * extASt[i]
		bBase += counter[i] * extBSt[i]
	}

	for line := loLine; line < hiLine; line++ {
		o := line * L
		switch {
		case aInner == 0 && bInner == 1:
			av := adata[aBase] // scalaire hissé
			dst := out[o : o+L]
			src := bdata[bBase : bBase+L]
			fillScalarVec(dst, av, src, op)
		case aInner == 1 && bInner == 0:
			bv := bdata[bBase]
			dst := out[o : o+L]
			src := adata[aBase : aBase+L]
			fillVecScalar(dst, src, bv, op)
		case aInner == 1 && bInner == 1:
			dst := out[o : o+L]
			fillVecVec(dst, adata[aBase:aBase+L], bdata[bBase:bBase+L], op)
		default: // aInner == 0 && bInner == 0 : valeur constante sur la ligne
			v := applyOp(op, adata[aBase], bdata[bBase])
			dst := out[o : o+L]
			for j := range dst {
				dst[j] = v
			}
		}
		// Avance sur les axes externes.
		for k := m - 1; k >= 0; k-- {
			counter[k]++
			aBase += extASt[k]
			bBase += extBSt[k]
			if counter[k] < extShape[k] {
				break
			}
			counter[k] = 0
			aBase -= extASt[k] * extShape[k]
			bBase -= extBSt[k] * extShape[k]
		}
	}
}

// Boucles internes contiguës spécialisées par opération (switch hors boucle),
// pour laisser le compilateur optimiser au mieux.

func fillScalarVec(dst []float64, s float64, src []float64, op binOp) {
	switch op {
	case opAdd:
		for j := range dst {
			dst[j] = s + src[j]
		}
	case opSub:
		for j := range dst {
			dst[j] = s - src[j]
		}
	case opMul:
		for j := range dst {
			dst[j] = s * src[j]
		}
	default:
		for j := range dst {
			dst[j] = s / src[j]
		}
	}
}

func fillVecScalar(dst, src []float64, s float64, op binOp) {
	switch op {
	case opAdd:
		for j := range dst {
			dst[j] = src[j] + s
		}
	case opSub:
		for j := range dst {
			dst[j] = src[j] - s
		}
	case opMul:
		for j := range dst {
			dst[j] = src[j] * s
		}
	default:
		for j := range dst {
			dst[j] = src[j] / s
		}
	}
}

func fillVecVec(dst, x, y []float64, op binOp) {
	switch op {
	case opAdd:
		for j := range dst {
			dst[j] = x[j] + y[j]
		}
	case opSub:
		for j := range dst {
			dst[j] = x[j] - y[j]
		}
	case opMul:
		for j := range dst {
			dst[j] = x[j] * y[j]
		}
	default:
		for j := range dst {
			dst[j] = x[j] / y[j]
		}
	}
}
