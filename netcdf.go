package xarray

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"
)

// Support d'un sous-ensemble du format netCDF « classique » (CDF-1).
//
// Périmètre assumé :
//   - dimensions de taille fixe (pas de dimension d'enregistrement illimitée) ;
//   - variables numériques de type NC_BYTE/NC_SHORT/NC_INT/NC_FLOAT/NC_DOUBLE ;
//   - coordonnées de dimension (variables 1D nommées comme leur dimension) ;
//   - pas d'attributs globaux ni de variable (listes ABSENT).
//
// Ce n'est PAS une implémentation complète de netCDF (ni NetCDF-4/HDF5, ni
// CDF-5, ni enregistrements illimités). Elle vise l'échange auto-cohérent
// (aller-retour) de tableaux étiquetés.

const (
	ncByte   int32 = 1
	ncShort  int32 = 3
	ncInt    int32 = 4
	ncFloat  int32 = 5
	ncDouble int32 = 6

	tagDimension int32 = 10
	tagVariable  int32 = 11
)

var ncEndian = binary.BigEndian

// ncTypeInfo déduit le type netCDF et la taille (octets) à partir du type Go T.
// Renvoie une erreur pour les types sans équivalent en CDF-1 (int/uint 64 bits…).
func ncTypeInfo[T Number]() (int32, int, error) {
	var z T
	switch any(z).(type) {
	case float64:
		return ncDouble, 8, nil
	case float32:
		return ncFloat, 4, nil
	case int32:
		return ncInt, 4, nil
	case int16:
		return ncShort, 2, nil
	case int8:
		return ncByte, 1, nil
	default:
		return 0, 0, fmt.Errorf("xarray: type %T non pris en charge par l'export netCDF (utilisez float64/float32/int32/int16/int8)", z)
	}
}

func pad4(n int) int { return (4 - n%4) % 4 }

// serializeNC encode des données numériques en big-endian selon ncType, avec
// remplissage jusqu'à un multiple de 4 octets.
func serializeNC[T Number](data []T, ncType int32) []byte {
	buf := new(bytes.Buffer)
	for _, x := range data {
		switch ncType {
		case ncDouble:
			binary.Write(buf, ncEndian, float64(x))
		case ncFloat:
			binary.Write(buf, ncEndian, float32(x))
		case ncInt:
			binary.Write(buf, ncEndian, int32(x))
		case ncShort:
			binary.Write(buf, ncEndian, int16(x))
		case ncByte:
			binary.Write(buf, ncEndian, int8(x))
		}
	}
	for i := 0; i < pad4(buf.Len()); i++ {
		buf.WriteByte(0)
	}
	return buf.Bytes()
}

// ncVarDesc décrit une variable prête à être écrite.
type ncVarDesc struct {
	name   string
	dimIDs []int32
	data   []byte // déjà encodé et paddé
}

func writeNCString(w io.Writer, s string) {
	binary.Write(w, ncEndian, int32(len(s)))
	w.Write([]byte(s))
	w.Write(make([]byte, pad4(len(s))))
}

// writeNCHeader écrit l'en-tête complet ; begins fournit l'offset de données de
// chaque variable (dans l'ordre de vars). Le type netCDF est uniforme (ncType).
func writeNCHeader(w io.Writer, dimNames []string, dimLens []int32, vars []ncVarDesc, begins []int32, ncType int32) {
	w.Write([]byte{'C', 'D', 'F', 1})
	binary.Write(w, ncEndian, int32(0)) // numrecs

	// dim_list
	if len(dimNames) == 0 {
		binary.Write(w, ncEndian, int32(0))
		binary.Write(w, ncEndian, int32(0))
	} else {
		binary.Write(w, ncEndian, tagDimension)
		binary.Write(w, ncEndian, int32(len(dimNames)))
		for i, name := range dimNames {
			writeNCString(w, name)
			binary.Write(w, ncEndian, dimLens[i])
		}
	}

	// gatt_list : ABSENT
	binary.Write(w, ncEndian, int32(0))
	binary.Write(w, ncEndian, int32(0))

	// var_list
	if len(vars) == 0 {
		binary.Write(w, ncEndian, int32(0))
		binary.Write(w, ncEndian, int32(0))
	} else {
		binary.Write(w, ncEndian, tagVariable)
		binary.Write(w, ncEndian, int32(len(vars)))
		for i, v := range vars {
			writeNCString(w, v.name)
			binary.Write(w, ncEndian, int32(len(v.dimIDs)))
			for _, id := range v.dimIDs {
				binary.Write(w, ncEndian, id)
			}
			// vatt_list : ABSENT
			binary.Write(w, ncEndian, int32(0))
			binary.Write(w, ncEndian, int32(0))
			binary.Write(w, ncEndian, ncType)
			binary.Write(w, ncEndian, int32(len(v.data))) // vsize (paddé)
			binary.Write(w, ncEndian, begins[i])          // begin
		}
	}
}

// WriteNetCDF écrit le Dataset au format netCDF classique (sous-ensemble).
func (ds *Dataset[T]) WriteNetCDF(w io.Writer) error {
	ncType, _, err := ncTypeInfo[T]()
	if err != nil {
		return err
	}

	// Dimensions triées pour un ordre déterministe.
	dimNames := make([]string, 0, len(ds.dims))
	for d := range ds.dims {
		dimNames = append(dimNames, d)
	}
	sort.Strings(dimNames)
	dimLens := make([]int32, len(dimNames))
	dimID := make(map[string]int32, len(dimNames))
	for i, d := range dimNames {
		dimLens[i] = int32(ds.dims[d])
		dimID[d] = int32(i)
	}

	var vars []ncVarDesc
	// Variables de coordonnées (1D nommées comme la dimension).
	coordNames := make([]string, 0, len(ds.coords))
	for d := range ds.coords {
		coordNames = append(coordNames, d)
	}
	sort.Strings(coordNames)
	for _, d := range coordNames {
		cv := ds.coords[d]
		vars = append(vars, ncVarDesc{
			name:   d,
			dimIDs: []int32{dimID[d]},
			data:   serializeNC(cv.data, ncType),
		})
	}
	// Variables de données.
	for _, name := range ds.VarNames() {
		da := ds.vars[name]
		ids := make([]int32, len(da.variable.dims))
		for i, d := range da.variable.dims {
			ids[i] = dimID[d]
		}
		vars = append(vars, ncVarDesc{
			name:   name,
			dimIDs: ids,
			data:   serializeNC(da.variable.data, ncType),
		})
	}

	// Mesure de la taille de l'en-tête (les begins ont une taille fixe).
	var probe bytes.Buffer
	zeros := make([]int32, len(vars))
	writeNCHeader(&probe, dimNames, dimLens, vars, zeros, ncType)
	headerSize := int32(probe.Len())

	begins := make([]int32, len(vars))
	cursor := headerSize
	for i, v := range vars {
		begins[i] = cursor
		cursor += int32(len(v.data))
	}

	writeNCHeader(w, dimNames, dimLens, vars, begins, ncType)
	for _, v := range vars {
		if _, err := w.Write(v.data); err != nil {
			return err
		}
	}
	return nil
}

// --- Lecture ----------------------------------------------------------------

type ncReader struct {
	b   []byte
	pos int
}

func (r *ncReader) i32() (int32, error) {
	if r.pos+4 > len(r.b) {
		return 0, io.ErrUnexpectedEOF
	}
	v := int32(ncEndian.Uint32(r.b[r.pos:]))
	r.pos += 4
	return v, nil
}

func (r *ncReader) str() (string, error) {
	n, err := r.i32()
	if err != nil {
		return "", err
	}
	if r.pos+int(n) > len(r.b) {
		return "", io.ErrUnexpectedEOF
	}
	s := string(r.b[r.pos : r.pos+int(n)])
	r.pos += int(n) + pad4(int(n))
	return s, nil
}

type ncVarRead struct {
	name    string
	dimIDs  []int32
	ncType  int32
	begin   int32
	nelems  int
	dimSize []int
}

// ReadDatasetNetCDF lit un Dataset depuis un flux netCDF classique (sous-ensemble).
// Les valeurs stockées sont converties vers le type T demandé.
func ReadDatasetNetCDF[T Number](rd io.Reader) (*Dataset[T], error) {
	raw, err := io.ReadAll(rd)
	if err != nil {
		return nil, err
	}
	if len(raw) < 4 || raw[0] != 'C' || raw[1] != 'D' || raw[2] != 'F' {
		return nil, fmt.Errorf("xarray: signature netCDF invalide")
	}
	if raw[3] != 1 {
		return nil, fmt.Errorf("xarray: seul le format netCDF classique (CDF-1) est pris en charge (version %d)", raw[3])
	}
	r := &ncReader{b: raw, pos: 4}
	if _, err := r.i32(); err != nil { // numrecs
		return nil, err
	}

	// dim_list
	dimNames := []string{}
	dimLens := []int{}
	tag, err := r.i32()
	if err != nil {
		return nil, err
	}
	count, err := r.i32()
	if err != nil {
		return nil, err
	}
	if tag == tagDimension {
		for i := int32(0); i < count; i++ {
			name, err := r.str()
			if err != nil {
				return nil, err
			}
			ln, err := r.i32()
			if err != nil {
				return nil, err
			}
			dimNames = append(dimNames, name)
			dimLens = append(dimLens, int(ln))
		}
	}

	// gatt_list (ignorée : on attend ABSENT)
	if _, err := r.i32(); err != nil {
		return nil, err
	}
	if _, err := r.i32(); err != nil {
		return nil, err
	}

	// var_list
	vtag, err := r.i32()
	if err != nil {
		return nil, err
	}
	vcount, err := r.i32()
	if err != nil {
		return nil, err
	}
	var vars []ncVarRead
	if vtag == tagVariable {
		for i := int32(0); i < vcount; i++ {
			name, err := r.str()
			if err != nil {
				return nil, err
			}
			nd, err := r.i32()
			if err != nil {
				return nil, err
			}
			ids := make([]int32, nd)
			sizes := make([]int, nd)
			nelems := 1
			for k := int32(0); k < nd; k++ {
				id, err := r.i32()
				if err != nil {
					return nil, err
				}
				ids[k] = id
				sizes[k] = dimLens[id]
				nelems *= dimLens[id]
			}
			// vatt_list (ABSENT attendu)
			if _, err := r.i32(); err != nil {
				return nil, err
			}
			if _, err := r.i32(); err != nil {
				return nil, err
			}
			ncType, err := r.i32()
			if err != nil {
				return nil, err
			}
			if _, err := r.i32(); err != nil { // vsize
				return nil, err
			}
			begin, err := r.i32()
			if err != nil {
				return nil, err
			}
			vars = append(vars, ncVarRead{name: name, dimIDs: ids, ncType: ncType, begin: begin, nelems: nelems, dimSize: sizes})
		}
	}

	// Lecture des données de chaque variable.
	coordSet := make(map[string]struct{}, len(dimNames))
	for _, d := range dimNames {
		coordSet[d] = struct{}{}
	}
	coordLabels := map[string][]T{}
	dataVars := map[string]*DataArray[T]{}

	for _, v := range vars {
		values, err := readNCValues[T](raw, int(v.begin), v.nelems, v.ncType)
		if err != nil {
			return nil, err
		}
		vdims := make([]string, len(v.dimIDs))
		for i, id := range v.dimIDs {
			vdims[i] = dimNames[id]
		}
		if _, isCoord := coordSet[v.name]; isCoord && len(vdims) == 1 && vdims[0] == v.name {
			coordLabels[v.name] = values
			continue
		}
		coords := map[string][]T{}
		da, err := NewDataArray(vdims, v.dimSize, values, coords, v.name)
		if err != nil {
			return nil, err
		}
		dataVars[v.name] = da
	}

	// Attache les coordonnées lues aux variables de données.
	for name, da := range dataVars {
		coords := map[string][]T{}
		for _, d := range da.variable.dims {
			if lbl, ok := coordLabels[d]; ok {
				coords[d] = lbl
			}
		}
		if len(coords) > 0 {
			rebuilt, err := NewDataArray(da.variable.Dims(), da.variable.Shape(), da.variable.Data(), coords, name)
			if err != nil {
				return nil, err
			}
			dataVars[name] = rebuilt
		}
	}

	return NewDataset(dataVars)
}

func readNCValues[T Number](raw []byte, begin, nelems int, ncType int32) ([]T, error) {
	size := ncElemSize(ncType)
	if size == 0 {
		return nil, fmt.Errorf("xarray: type netCDF %d non pris en charge à la lecture", ncType)
	}
	if begin+nelems*size > len(raw) {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]T, nelems)
	p := begin
	for i := 0; i < nelems; i++ {
		switch ncType {
		case ncDouble:
			out[i] = T(math.Float64frombits(ncEndian.Uint64(raw[p:])))
		case ncFloat:
			out[i] = T(math.Float32frombits(ncEndian.Uint32(raw[p:])))
		case ncInt:
			out[i] = T(int32(ncEndian.Uint32(raw[p:])))
		case ncShort:
			out[i] = T(int16(ncEndian.Uint16(raw[p:])))
		case ncByte:
			out[i] = T(int8(raw[p]))
		}
		p += size
	}
	return out, nil
}

func ncElemSize(ncType int32) int {
	switch ncType {
	case ncDouble:
		return 8
	case ncFloat, ncInt:
		return 4
	case ncShort:
		return 2
	case ncByte:
		return 1
	default:
		return 0
	}
}

// WriteNetCDF écrit un DataArray comme un dataset netCDF à une seule variable.
func (da *DataArray[T]) WriteNetCDF(w io.Writer) error {
	name := da.name
	if name == "" {
		name = "data"
	}
	ds, err := NewDataset(map[string]*DataArray[T]{name: da.clone()})
	if err != nil {
		return err
	}
	return ds.WriteNetCDF(w)
}

// ReadDataArrayNetCDF lit un DataArray depuis un dataset netCDF ne contenant
// qu'une seule variable de données.
func ReadDataArrayNetCDF[T Number](r io.Reader) (*DataArray[T], error) {
	ds, err := ReadDatasetNetCDF[T](r)
	if err != nil {
		return nil, err
	}
	names := ds.VarNames()
	if len(names) != 1 {
		return nil, fmt.Errorf("xarray: %d variable(s) de données, une seule attendue", len(names))
	}
	return ds.Get(names[0])
}
