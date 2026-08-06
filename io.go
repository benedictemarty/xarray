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
type daJSON struct {
	Name   string               `json:"name,omitempty"`
	Dims   []string             `json:"dims"`
	Shape  []int                `json:"shape"`
	Data   []float64            `json:"data"`
	Coords map[string][]float64 `json:"coords,omitempty"`
}

func (da *DataArray) toJSON() daJSON {
	coords := make(map[string][]float64, len(da.coords))
	for k, cv := range da.coords {
		coords[k] = cv.Data()
	}
	if len(coords) == 0 {
		coords = nil
	}
	return daJSON{
		Name:   da.name,
		Dims:   da.variable.Dims(),
		Shape:  da.variable.Shape(),
		Data:   da.variable.Data(),
		Coords: coords,
	}
}

// MarshalJSON sérialise le DataArray.
func (da *DataArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(da.toJSON())
}

func dataArrayFromJSON(j daJSON) (*DataArray, error) {
	return NewDataArray(j.Dims, j.Shape, j.Data, j.Coords, j.Name)
}

// UnmarshalJSON désérialise un DataArray (avec validation via NewDataArray).
func (da *DataArray) UnmarshalJSON(b []byte) error {
	var j daJSON
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
func (da *DataArray) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(da.toJSON())
}

// ReadDataArrayJSON lit un DataArray depuis un flux JSON.
func ReadDataArrayJSON(r io.Reader) (*DataArray, error) {
	var j daJSON
	if err := json.NewDecoder(r).Decode(&j); err != nil {
		return nil, err
	}
	return dataArrayFromJSON(j)
}

// dsJSON est la forme sérialisable d'un Dataset.
type dsJSON struct {
	Dims   map[string]int       `json:"dims"`
	Coords map[string][]float64 `json:"coords,omitempty"`
	Vars   map[string]daJSON    `json:"data_vars"`
}

func (ds *Dataset) toJSON() dsJSON {
	coords := make(map[string][]float64, len(ds.coords))
	for k, cv := range ds.coords {
		coords[k] = cv.Data()
	}
	if len(coords) == 0 {
		coords = nil
	}
	vars := make(map[string]daJSON, len(ds.vars))
	for name, da := range ds.vars {
		vars[name] = da.toJSON()
	}
	return dsJSON{Dims: ds.Dims(), Coords: coords, Vars: vars}
}

// WriteJSON écrit le Dataset au format JSON.
func (ds *Dataset) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ds.toJSON())
}

// ReadDatasetJSON lit un Dataset depuis un flux JSON. Les coordonnées partagées
// du dataset sont réinjectées dans chaque variable selon ses dimensions, afin de
// garantir la cohérence quelle que soit la variable qui les portait.
func ReadDatasetJSON(r io.Reader) (*Dataset, error) {
	var j dsJSON
	if err := json.NewDecoder(r).Decode(&j); err != nil {
		return nil, err
	}
	vars := make(map[string]*DataArray, len(j.Vars))
	for name, vj := range j.Vars {
		coords := map[string][]float64{}
		for k, v := range vj.Coords {
			coords[k] = v
		}
		// Réinjecte les coordonnées partagées manquantes.
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
//
// Chaque ligne contient une étiquette par dimension puis la valeur. L'en-tête
// donne les noms des dimensions suivis du nom de la variable (ou "value").
// Ce format est général (N dimensions) et sans ambiguïté.

// WriteCSV écrit le DataArray au format tidy.
func (da *DataArray) WriteCSV(w io.Writer) error {
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

	// Étiquettes par dimension : coordonnée si présente, sinon indices 0..n-1.
	labels := make([][]float64, len(dims))
	for i, d := range dims {
		if cv, ok := da.coords[d]; ok {
			labels[i] = cv.data
		} else {
			idx := make([]float64, shape[i])
			for k := range idx {
				idx[k] = float64(k)
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
			rec[i] = formatFloat(labels[i][counter[i]])
		}
		rec[len(dims)] = formatFloat(da.variable.data[flat])
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
// dernière colonne est la valeur ; les précédentes sont les dimensions. Les
// coordonnées sont reconstruites dans l'ordre d'apparition des étiquettes.
func ReadDataArrayCSV(r io.Reader) (*DataArray, error) {
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

	// Coordonnées reconstruites dans l'ordre d'apparition.
	coordLabels := make([]map[float64]int, ndim)
	coordOrder := make([][]float64, ndim)
	for i := range coordLabels {
		coordLabels[i] = map[float64]int{}
	}

	type cell struct {
		pos []int
		val float64
	}
	cells := make([]cell, 0, len(records)-1)

	for _, rec := range records[1:] {
		if len(rec) != ndim+1 {
			return nil, fmt.Errorf("xarray: ligne CSV de %d champ(s), %d attendu(s)", len(rec), ndim+1)
		}
		pos := make([]int, ndim)
		for i := 0; i < ndim; i++ {
			lbl, err := strconv.ParseFloat(rec[i], 64)
			if err != nil {
				return nil, fmt.Errorf("xarray: étiquette %q non numérique (dimension %q)", rec[i], dims[i])
			}
			if _, ok := coordLabels[i][lbl]; !ok {
				coordLabels[i][lbl] = len(coordOrder[i])
				coordOrder[i] = append(coordOrder[i], lbl)
			}
			pos[i] = coordLabels[i][lbl]
		}
		val, err := strconv.ParseFloat(rec[ndim], 64)
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
	data := make([]float64, size)
	filled := make([]bool, size)

	// Strides en ordre C.
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

	coords := make(map[string][]float64, ndim)
	for i, d := range dims {
		coords[d] = coordOrder[i]
	}
	return NewDataArray(dims, shape, data, coords, name)
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
