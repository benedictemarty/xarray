package xarray

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// DataArray est un tableau N-dimensionnel étiqueté : une Variable enrichie de
// coordonnées associées à ses dimensions et d'un nom optionnel.
//
// Les coordonnées permettent l'indexation par label (Sel) en plus de
// l'indexation par position (Isel).
type DataArray struct {
	variable *Variable
	// coords associe un nom de coordonnée à une Variable 1D. Pour ce premier
	// incrément, seules les coordonnées dites « de dimension » sont gérées :
	// leur nom est celui d'une dimension et leur longueur correspond à la
	// taille de cette dimension.
	coords map[string]*Variable
	name   string
}

// NewDataArray construit un DataArray.
//
//   - dims/shape/data définissent la Variable sous-jacente (cf. NewVariable) ;
//   - coords associe optionnellement à certaines dimensions un vecteur d'étiquettes
//     de même longueur que la dimension ;
//   - name est le nom (optionnel) du tableau.
func NewDataArray(dims []string, shape []int, data []float64, coords map[string][]float64, name string) (*DataArray, error) {
	v, err := NewVariable(dims, shape, data)
	if err != nil {
		return nil, err
	}

	coordVars := make(map[string]*Variable, len(coords))
	for dim, labels := range coords {
		axis := v.dimIndex(dim)
		if axis == -1 {
			return nil, fmt.Errorf("xarray: coordonnée %q ne correspond à aucune dimension", dim)
		}
		if len(labels) != v.shape[axis] {
			return nil, fmt.Errorf("xarray: coordonnée %q de longueur %d incompatible avec la dimension de taille %d", dim, len(labels), v.shape[axis])
		}
		cv, err := NewVariable([]string{dim}, []int{len(labels)}, labels)
		if err != nil {
			return nil, err
		}
		coordVars[dim] = cv
	}

	return &DataArray{variable: v, coords: coordVars, name: name}, nil
}

// Name renvoie le nom du tableau.
func (da *DataArray) Name() string { return da.name }

// Rename renvoie une copie du tableau portant le nom fourni.
func (da *DataArray) Rename(name string) *DataArray {
	c := da.clone()
	c.name = name
	return c
}

// Dims renvoie les noms de dimensions.
func (da *DataArray) Dims() []string { return da.variable.Dims() }

// HasDim indique si le tableau possède la dimension dim.
func (da *DataArray) HasDim(dim string) bool { return da.variable.dimIndex(dim) != -1 }

// Shape renvoie la forme du tableau.
func (da *DataArray) Shape() []int { return da.variable.Shape() }

// Ndim renvoie le nombre de dimensions.
func (da *DataArray) Ndim() int { return da.variable.Ndim() }

// Size renvoie le nombre total d'éléments.
func (da *DataArray) Size() int { return da.variable.Size() }

// Data renvoie une copie des données plates (ordre C).
func (da *DataArray) Data() []float64 { return da.variable.Data() }

// Variable renvoie la Variable sous-jacente.
func (da *DataArray) Variable() *Variable { return da.variable }

// Coord renvoie les étiquettes de la coordonnée associée à dim, ou une erreur
// si aucune coordonnée n'est définie.
func (da *DataArray) Coord(dim string) ([]float64, error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée pour la dimension %q", dim)
	}
	return cv.Data(), nil
}

// clone effectue une copie profonde du DataArray.
func (da *DataArray) clone() *DataArray {
	coords := make(map[string]*Variable, len(da.coords))
	for k, cv := range da.coords {
		nv, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
		coords[k] = nv
	}
	nv, _ := NewVariable(da.variable.Dims(), da.variable.Shape(), da.variable.Data())
	return &DataArray{variable: nv, coords: coords, name: da.name}
}

// Isel sélectionne par position entière le long d'une dimension. La dimension
// est supprimée du résultat, ainsi que sa coordonnée éventuelle.
func (da *DataArray) Isel(dim string, index int) (*DataArray, error) {
	nv, err := da.variable.Isel(dim, index)
	if err != nil {
		return nil, err
	}
	coords := make(map[string]*Variable, len(da.coords))
	for k, cv := range da.coords {
		if k == dim {
			continue // la coordonnée de la dimension réduite disparaît
		}
		ncv, _ := NewVariable(cv.Dims(), cv.Shape(), cv.Data())
		coords[k] = ncv
	}
	return &DataArray{variable: nv, coords: coords, name: da.name}, nil
}

// Sel sélectionne par label le long d'une dimension : l'étiquette est recherchée
// dans la coordonnée de la dimension, puis Isel est appliqué à sa position.
func (da *DataArray) Sel(dim string, label float64) (*DataArray, error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: indexation par label impossible : aucune coordonnée pour la dimension %q", dim)
	}
	pos := -1
	for i, l := range cv.data {
		if l == label {
			pos = i
			break
		}
	}
	if pos == -1 {
		return nil, fmt.Errorf("xarray: étiquette %v absente de la coordonnée %q", label, dim)
	}
	return da.Isel(dim, pos)
}

// --- Réductions ---------------------------------------------------------

// Sum renvoie la somme de tous les éléments.
func (da *DataArray) Sum() float64 {
	var s float64
	for _, x := range da.variable.data {
		s += x
	}
	return s
}

// Mean renvoie la moyenne de tous les éléments. Renvoie NaN pour un tableau vide.
func (da *DataArray) Mean() float64 {
	n := len(da.variable.data)
	if n == 0 {
		return math.NaN()
	}
	return da.Sum() / float64(n)
}

// Min renvoie le minimum de tous les éléments. Renvoie NaN pour un tableau vide.
func (da *DataArray) Min() float64 {
	d := da.variable.data
	if len(d) == 0 {
		return math.NaN()
	}
	m := d[0]
	for _, x := range d[1:] {
		if x < m {
			m = x
		}
	}
	return m
}

// Max renvoie le maximum de tous les éléments. Renvoie NaN pour un tableau vide.
func (da *DataArray) Max() float64 {
	d := da.variable.data
	if len(d) == 0 {
		return math.NaN()
	}
	m := d[0]
	for _, x := range d[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

// String fournit une représentation lisible du DataArray.
func (da *DataArray) String() string {
	var b strings.Builder
	name := da.name
	if name == "" {
		name = "<sans nom>"
	}
	fmt.Fprintf(&b, "<xarray.DataArray %s (", name)
	parts := make([]string, da.variable.Ndim())
	for i, d := range da.variable.dims {
		parts[i] = fmt.Sprintf("%s: %d", d, da.variable.shape[i])
	}
	b.WriteString(strings.Join(parts, ", "))
	b.WriteString(")>\n")
	fmt.Fprintf(&b, "Données : %v\n", da.variable.data)

	if len(da.coords) > 0 {
		b.WriteString("Coordonnées :\n")
		names := make([]string, 0, len(da.coords))
		for k := range da.coords {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			fmt.Fprintf(&b, "  * %-8s %v\n", k, da.coords[k].data)
		}
	}
	return b.String()
}
