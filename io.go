package xarray

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// --- JSON -------------------------------------------------------------------

// daJSON est la forme sérialisable d'un DataArray.
type daJSON[T Number] struct {
	Name   string         `json:"name,omitempty"`
	Dims   []string       `json:"dims"`
	Shape  []int          `json:"shape"`
	Data   []T            `json:"data"`
	Coords map[string][]T `json:"coords,omitempty"`
}

func (da *DataArray[T]) toJSON() daJSON[T] {
	coords := make(map[string][]T, len(da.coords))
	for k, cv := range da.coords {
		coords[k] = cv.Data()
	}
	if len(coords) == 0 {
		coords = nil
	}
	return daJSON[T]{
		Name:   da.name,
		Dims:   da.variable.Dims(),
		Shape:  da.variable.Shape(),
		Data:   da.variable.Data(),
		Coords: coords,
	}
}

// MarshalJSON sérialise le DataArray.
func (da *DataArray[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(da.toJSON())
}

func dataArrayFromJSON[T Number](j daJSON[T]) (*DataArray[T], error) {
	return NewDataArray(j.Dims, j.Shape, j.Data, j.Coords, j.Name)
}

// UnmarshalJSON désérialise un DataArray (avec validation via NewDataArray).
func (da *DataArray[T]) UnmarshalJSON(b []byte) error {
	var j daJSON[T]
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	built, err := dataArrayFromJSON(j)
	if err != nil {
		return err
	}
	*da = *built
	return nil
}

// WriteJSON écrit le DataArray au format JSON.
func (da *DataArray[T]) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(da.toJSON())
}

// ReadDataArrayJSON lit un DataArray depuis un flux JSON.
func ReadDataArrayJSON[T Number](r io.Reader) (*DataArray[T], error) {
	var j daJSON[T]
	if err := json.NewDecoder(r).Decode(&j); err != nil {
		return nil, err
	}
	return dataArrayFromJSON(j)
}

// dsJSON est la forme sérialisable d'un Dataset.
type dsJSON[T Number] struct {
	Dims   map[string]int       `json:"dims"`
	Coords map[string][]T       `json:"coords,omitempty"`
	Vars   map[string]daJSON[T] `json:"data_vars"`
}

func (ds *Dataset[T]) toJSON() dsJSON[T] {
	coords := make(map[string][]T, len(ds.coords))
	for k, cv := range ds.coords {
		coords[k] = cv.Data()
	}
	if len(coords) == 0 {
		coords = nil
	}
	vars := make(map[string]daJSON[T], len(ds.vars))
	for name, da := range ds.vars {
		vars[name] = da.toJSON()
	}
	return dsJSON[T]{Dims: ds.Dims(), Coords: coords, Vars: vars}
}

// WriteJSON écrit le Dataset au format JSON.
func (ds *Dataset[T]) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ds.toJSON())
}

// ReadDatasetJSON lit un Dataset depuis un flux JSON. Les coordonnées partagées
// sont réinjectées dans chaque variable selon ses dimensions.
func ReadDatasetJSON[T Number](r io.Reader) (*Dataset[T], error) {
	var j dsJSON[T]
	if err := json.NewDecoder(r).Decode(&j); err != nil {
		return nil, err
	}
	vars := make(map[string]*DataArray[T], len(j.Vars))
	for name, vj := range j.Vars {
		coords := map[string][]T{}
		for k, v := range vj.Coords {
			coords[k] = v
		}
		for _, dim := range vj.Dims {
			if _, ok := coords[dim]; !ok {
				if shared, ok := j.Coords[dim]; ok {
					coords[dim] = shared
				}
			}
		}
		da, err := NewDataArray(vj.Dims, vj.Shape, vj.Data, coords, vj.Name)
		if err != nil {
			return nil, fmt.Errorf("xarray: variable %q : %w", name, err)
		}
		vars[name] = da
	}
	return NewDataset(vars)
}

// --- CSV (format « tidy » : une ligne par cellule) --------------------------

// WriteCSV écrit le DataArray au format tidy : une colonne par dimension puis
// la valeur ; l'en-tête donne les noms des dimensions et de la variable.
func (da *DataArray[T]) WriteCSV(w io.Writer) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	valueName := da.name
	if valueName == "" {
		valueName = "value"
	}
	header := append(da.variable.Dims(), valueName)
	if err := cw.Write(header); err != nil {
		return err
	}

	dims := da.variable.dims
	shape := da.variable.shape
	strides := da.variable.strides()

	labels := make([][]T, len(dims))
	for i, d := range dims {
		if cv, ok := da.coords[d]; ok {
			labels[i] = cv.data
		} else {
			idx := make([]T, shape[i])
			for k := range idx {
				idx[k] = T(k)
			}
			labels[i] = idx
		}
	}

	counter := make([]int, len(dims))
	total := da.variable.Size()
	for n := 0; n < total; n++ {
		flat := 0
		rec := make([]string, len(dims)+1)
		for i := range dims {
			flat += counter[i] * strides[i]
			rec[i] = formatNum(labels[i][counter[i]])
		}
		rec[len(dims)] = formatNum(da.variable.data[flat])
		if err := cw.Write(rec); err != nil {
			return err
		}
		for k := len(counter) - 1; k >= 0; k-- {
			counter[k]++
			if counter[k] < shape[k] {
				break
			}
			counter[k] = 0
		}
	}
	return cw.Error()
}

// ReadDataArrayCSV lit un DataArray depuis un flux CSV au format tidy. La
// dernière colonne est la valeur ; les précédentes sont les dimensions.
func ReadDataArrayCSV[T Number](r io.Reader) (*DataArray[T], error) {
	cr := csv.NewReader(r)
	records, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 1 {
		return nil, fmt.Errorf("xarray: CSV vide")
	}
	header := records[0]
	if len(header) < 2 {
		return nil, fmt.Errorf("xarray: en-tête CSV invalide (au moins une dimension et une valeur attendues)")
	}
	ndim := len(header) - 1
	dims := header[:ndim]
	name := header[ndim]
	if name == "value" {
		name = ""
	}

	coordLabels := make([]map[T]int, ndim)
	coordOrder := make([][]T, ndim)
	for i := range coordLabels {
		coordLabels[i] = map[T]int{}
	}

	type cell struct {
		pos []int
		val T
	}
	cells := make([]cell, 0, len(records)-1)

	for _, rec := range records[1:] {
		if len(rec) != ndim+1 {
			return nil, fmt.Errorf("xarray: ligne CSV de %d champ(s), %d attendu(s)", len(rec), ndim+1)
		}
		pos := make([]int, ndim)
		for i := 0; i < ndim; i++ {
			lbl, err := parseNum[T](rec[i])
			if err != nil {
				return nil, fmt.Errorf("xarray: étiquette %q non numérique (dimension %q)", rec[i], dims[i])
			}
			if _, ok := coordLabels[i][lbl]; !ok {
				coordLabels[i][lbl] = len(coordOrder[i])
				coordOrder[i] = append(coordOrder[i], lbl)
			}
			pos[i] = coordLabels[i][lbl]
		}
		val, err := parseNum[T](rec[ndim])
		if err != nil {
			return nil, fmt.Errorf("xarray: valeur %q non numérique", rec[ndim])
		}
		cells = append(cells, cell{pos: pos, val: val})
	}

	shape := make([]int, ndim)
	for i := range shape {
		shape[i] = len(coordOrder[i])
	}
	size := product(shape)
	data := make([]T, size)
	filled := make([]bool, size)

	strides := make([]int, ndim)
	acc := 1
	for i := ndim - 1; i >= 0; i-- {
		strides[i] = acc
		acc *= shape[i]
	}
	for _, c := range cells {
		flat := 0
		for i, p := range c.pos {
			flat += p * strides[i]
		}
		data[flat] = c.val
		filled[flat] = true
	}
	for i, f := range filled {
		if !f {
			return nil, fmt.Errorf("xarray: cellule manquante à l'indice plat %d (grille incomplète)", i)
		}
	}

	coords := make(map[string][]T, ndim)
	for i, d := range dims {
		coords[d] = coordOrder[i]
	}
	return NewDataArray(dims, shape, data, coords, name)
}

func formatNum[T Number](x T) string {
	return strconv.FormatFloat(float64(x), 'g', -1, 64)
}

func parseNum[T Number](s string) (T, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return T(f), nil
}
