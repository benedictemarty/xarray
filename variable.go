// Package xarray fournit des tableaux N-dimensionnels étiquetés en Go,
// inspirés de la bibliothèque Python xarray.
//
// L'architecture reprend celle de xarray :
//
//   - Variable  : conteneur bas niveau (données plates + dimensions nommées) ;
//   - DataArray : Variable + coordonnées étiquetées + nom (indexation par label) ;
//   - Dataset   : collection de DataArrays partageant des dimensions et coordonnées.
//
// Pour ce premier incrément, les données sont stockées en float64 et
// disposées en mémoire selon l'ordre C (row-major).
package xarray

import (
	"fmt"
	"strings"
)

// Variable est un tableau N-dimensionnel dont les axes portent un nom.
// C'est la brique de base sur laquelle DataArray et Dataset sont construits.
//
// Les données sont stockées à plat en ordre C (row-major) : le dernier axe
// varie le plus vite.
type Variable struct {
	dims  []string
	shape []int
	data  []float64
	attrs map[string]string
}

// NewVariable construit une Variable à partir de noms de dimensions, d'une forme
// (shape) et de données plates en ordre C.
//
// Contraintes vérifiées :
//   - len(dims) == len(shape) ;
//   - les noms de dimensions sont non vides et uniques ;
//   - les tailles sont positives ou nulles ;
//   - len(data) == produit des tailles.
func NewVariable(dims []string, shape []int, data []float64) (*Variable, error) {
	if len(dims) != len(shape) {
		return nil, fmt.Errorf("xarray: %d dimension(s) mais %d taille(s) de forme", len(dims), len(shape))
	}
	seen := make(map[string]struct{}, len(dims))
	for _, d := range dims {
		if d == "" {
			return nil, fmt.Errorf("xarray: nom de dimension vide interdit")
		}
		if _, ok := seen[d]; ok {
			return nil, fmt.Errorf("xarray: dimension %q dupliquée", d)
		}
		seen[d] = struct{}{}
	}
	size := 1
	for i, s := range shape {
		if s < 0 {
			return nil, fmt.Errorf("xarray: taille négative %d pour la dimension %q", s, dims[i])
		}
		size *= s
	}
	if len(data) != size {
		return nil, fmt.Errorf("xarray: %d valeur(s) fournie(s) pour une forme de taille %d", len(data), size)
	}

	dimsCopy := append([]string(nil), dims...)
	shapeCopy := append([]int(nil), shape...)
	dataCopy := append([]float64(nil), data...)

	return &Variable{
		dims:  dimsCopy,
		shape: shapeCopy,
		data:  dataCopy,
		attrs: map[string]string{},
	}, nil
}

// Dims renvoie une copie des noms de dimensions.
func (v *Variable) Dims() []string { return append([]string(nil), v.dims...) }

// Shape renvoie une copie de la forme (taille de chaque dimension).
func (v *Variable) Shape() []int { return append([]int(nil), v.shape...) }

// Ndim renvoie le nombre de dimensions.
func (v *Variable) Ndim() int { return len(v.dims) }

// Size renvoie le nombre total d'éléments.
func (v *Variable) Size() int {
	size := 1
	for _, s := range v.shape {
		size *= s
	}
	return size
}

// Data renvoie une copie des données plates (ordre C).
func (v *Variable) Data() []float64 { return append([]float64(nil), v.data...) }

// Attrs renvoie une copie des attributs (métadonnées libres).
func (v *Variable) Attrs() map[string]string {
	out := make(map[string]string, len(v.attrs))
	for k, val := range v.attrs {
		out[k] = val
	}
	return out
}

// SetAttr définit un attribut (métadonnée) sur la variable.
func (v *Variable) SetAttr(key, value string) {
	if v.attrs == nil {
		v.attrs = map[string]string{}
	}
	v.attrs[key] = value
}

// dimIndex renvoie l'indice de l'axe portant le nom dim, ou -1 s'il est absent.
func (v *Variable) dimIndex(dim string) int {
	for i, d := range v.dims {
		if d == dim {
			return i
		}
	}
	return -1
}

// strides calcule les pas (strides) en ordre C pour la forme courante.
func (v *Variable) strides() []int {
	st := make([]int, len(v.shape))
	acc := 1
	for i := len(v.shape) - 1; i >= 0; i-- {
		st[i] = acc
		acc *= v.shape[i]
	}
	return st
}

// flatIndex convertit un multi-indice en indice plat (ordre C).
func (v *Variable) flatIndex(idx []int) (int, error) {
	if len(idx) != len(v.shape) {
		return 0, fmt.Errorf("xarray: %d indice(s) fourni(s) pour un tableau à %d dimension(s)", len(idx), len(v.shape))
	}
	st := v.strides()
	flat := 0
	for i, k := range idx {
		if k < 0 || k >= v.shape[i] {
			return 0, fmt.Errorf("xarray: indice %d hors bornes [0,%d) sur la dimension %q", k, v.shape[i], v.dims[i])
		}
		flat += k * st[i]
	}
	return flat, nil
}

// At renvoie la valeur au multi-indice positionnel donné.
func (v *Variable) At(idx ...int) (float64, error) {
	flat, err := v.flatIndex(idx)
	if err != nil {
		return 0, err
	}
	return v.data[flat], nil
}

// Isel (integer select) sélectionne une position entière sur une dimension et
// renvoie une nouvelle Variable dont cette dimension est supprimée.
//
// C'est l'équivalent bas niveau de DataArray.isel de xarray.
func (v *Variable) Isel(dim string, index int) (*Variable, error) {
	axis := v.dimIndex(dim)
	if axis == -1 {
		return nil, fmt.Errorf("xarray: dimension %q absente", dim)
	}
	if index < 0 || index >= v.shape[axis] {
		return nil, fmt.Errorf("xarray: indice %d hors bornes [0,%d) sur la dimension %q", index, v.shape[axis], dim)
	}

	newDims := make([]string, 0, len(v.dims)-1)
	newShape := make([]int, 0, len(v.shape)-1)
	for i := range v.dims {
		if i == axis {
			continue
		}
		newDims = append(newDims, v.dims[i])
		newShape = append(newShape, v.shape[i])
	}

	newSize := 1
	for _, s := range newShape {
		newSize *= s
	}
	newData := make([]float64, 0, newSize)

	st := v.strides()
	// Parcours de toutes les positions de la sous-tranche.
	counter := make([]int, len(newShape))
	for done := false; !done; {
		flat := index * st[axis]
		j := 0
		for i := range v.dims {
			if i == axis {
				continue
			}
			flat += counter[j] * st[i]
			j++
		}
		newData = append(newData, v.data[flat])

		// Incrément du compteur multi-dimensionnel (ordre C).
		if len(counter) == 0 {
			done = true
		}
		for k := len(counter) - 1; k >= 0; k-- {
			counter[k]++
			if counter[k] < newShape[k] {
				break
			}
			counter[k] = 0
			if k == 0 {
				done = true
			}
		}
	}

	return &Variable{
		dims:  newDims,
		shape: newShape,
		data:  newData,
		attrs: v.Attrs(),
	}, nil
}

// String fournit une représentation lisible de la variable.
func (v *Variable) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "<xarray.Variable (")
	parts := make([]string, len(v.dims))
	for i, d := range v.dims {
		parts[i] = fmt.Sprintf("%s: %d", d, v.shape[i])
	}
	b.WriteString(strings.Join(parts, ", "))
	b.WriteString(")>\n")
	fmt.Fprintf(&b, "%v", v.data)
	return b.String()
}
