package xarray

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Support d'un sous-ensemble du format de stockage Zarr v2, sur système de
// fichiers.
//
// Périmètre assumé :
//   - Zarr v2 (le format le plus répandu ; v3 non géré) ;
//   - store = répertoire du système de fichiers ;
//   - DataArray[float64], dtype "<f8" (float64 little-endian), ordre C ;
//   - découpage en chunks (les chunks de bord sont complétés par fill_value) ;
//   - compression : aucune ou "zlib" (via la bibliothèque standard Go) ;
//   - dimensions, nom et coordonnées stockés dans `.zattrs`
//     (`_ARRAY_DIMENSIONS` suit la convention xarray).
//
// Non géré : Zarr v3, dtypes autres que float64, ordre Fortran, compresseurs
// blosc/zstd/lz4, filtres, groupes hiérarchiques. L'aller-retour est validé en
// interne ; l'interopérabilité avec zarr-python n'est pas garantie.

// ZarrCompression désigne le compresseur utilisé pour les chunks.
type ZarrCompression int

const (
	ZarrNone ZarrCompression = iota // aucune compression
	ZarrZlib                        // zlib (numcodecs "zlib")
)

type zarrCompressorMeta struct {
	ID    string `json:"id"`
	Level int    `json:"level,omitempty"`
	// Champs Blosc (présents à la lecture d'un store zarr-python typique).
	Cname     string `json:"cname,omitempty"`
	Shuffle   int    `json:"shuffle,omitempty"`
	Blocksize int    `json:"blocksize,omitempty"`
}

type zarrayMeta struct {
	ZarrFormat int                 `json:"zarr_format"`
	Shape      []int               `json:"shape"`
	Chunks     []int               `json:"chunks"`
	Dtype      string              `json:"dtype"`
	Compressor *zarrCompressorMeta `json:"compressor"`
	FillValue  json.RawMessage     `json:"fill_value"`
	Order      string              `json:"order"`
	Filters    interface{}         `json:"filters"`
}

// parseZarrFill interprète un fill_value Zarr : nombre, null, ou les chaînes
// spéciales "NaN"/"Infinity"/"-Infinity" (utilisées par zarr-python pour les
// flottants). Renvoie la valeur de remplissage (0 si null/absent).
func parseZarrFill(raw json.RawMessage) (float64, error) {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" {
		return 0, nil
	}
	if s[0] == '"' {
		var str string
		if err := json.Unmarshal(raw, &str); err != nil {
			return 0, err
		}
		switch str {
		case "NaN":
			return math.NaN(), nil
		case "Infinity":
			return math.Inf(1), nil
		case "-Infinity":
			return math.Inf(-1), nil
		default:
			return 0, fmt.Errorf("xarray: fill_value %q non reconnu", str)
		}
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0, err
	}
	return f, nil
}

type zattrsMeta struct {
	Dims   []string             `json:"_ARRAY_DIMENSIONS"`
	Name   string               `json:"name,omitempty"`
	Coords map[string][]float64 `json:"coords,omitempty"`
}

// WriteDataArrayZarr écrit un DataArray[float64] dans le répertoire dir au format
// Zarr v2. chunks donne la taille de chunk par dimension (même longueur que la
// forme). comp choisit la compression des chunks.
func WriteDataArrayZarr(dir string, da *DataArray[float64], chunks []int, comp ZarrCompression) error {
	coords := map[string][]float64{}
	for k, cv := range da.coords {
		coords[k] = cv.Data()
	}
	if len(coords) == 0 {
		coords = nil
	}
	return writeZarrArrayInternal(dir, da.variable.Dims(), da.variable.Shape(), da.variable.data, da.name, coords, chunks, comp)
}

// writeZarrArrayInternal écrit un array Zarr v2 dans dir. coords peut être nil
// (cas d'un array « nu » dans un groupe, où les coordonnées sont des arrays
// séparés) ou porter les coordonnées d'un DataArray autonome (stockées en
// `.zattrs`).
func writeZarrArrayInternal(dir string, dims []string, shape []int, data []float64, name string, coords map[string][]float64, chunks []int, comp ZarrCompression) error {
	if len(chunks) != len(shape) {
		return fmt.Errorf("xarray: %d taille(s) de chunk pour %d dimension(s)", len(chunks), len(shape))
	}
	for i, c := range chunks {
		if c <= 0 {
			return fmt.Errorf("xarray: taille de chunk invalide %d sur la dimension %d", c, i)
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// .zarray — fill_value: null (et non 0). Un fill_value numérique est
	// interprété par xarray/zarr-python comme _FillValue et masque toutes les
	// valeurs égales (ici les 0 légitimes deviendraient NaN). null = pas de
	// masquage ; l'aller-retour reste correct car tous les chunks sont écrits.
	meta := zarrayMeta{
		ZarrFormat: 2, Shape: shape, Chunks: chunks, Dtype: "<f8",
		FillValue: nil, Order: "C",
	}
	if comp == ZarrZlib {
		meta.Compressor = &zarrCompressorMeta{ID: "zlib", Level: 1}
	}
	if err := writeJSONFile(filepath.Join(dir, ".zarray"), meta); err != nil {
		return err
	}

	// .zattrs
	if len(coords) == 0 {
		coords = nil
	}
	attrs := zattrsMeta{Dims: dims, Name: name, Coords: coords}
	if err := writeJSONFile(filepath.Join(dir, ".zattrs"), attrs); err != nil {
		return err
	}

	// Chunks.
	ndim := len(shape)
	nchunks := make([]int, ndim)
	for d := 0; d < ndim; d++ {
		nchunks[d] = (shape[d] + chunks[d] - 1) / chunks[d]
	}
	dataStrides := cOrderStrides(shape)
	chunkStrides := cOrderStrides(chunks)
	chunkSize := product(chunks)

	coord := make([]int, ndim)
	for done := false; !done; {
		buf := make([]float64, chunkSize) // fill_value = 0 par défaut
		local := make([]int, ndim)
		for l := 0; l < chunkSize; l++ {
			inBounds := true
			flatGlobal := 0
			for d := 0; d < ndim; d++ {
				g := coord[d]*chunks[d] + local[d]
				if g >= shape[d] {
					inBounds = false
					break
				}
				flatGlobal += g * dataStrides[d]
			}
			if inBounds {
				flatLocal := 0
				for d := 0; d < ndim; d++ {
					flatLocal += local[d] * chunkStrides[d]
				}
				buf[flatLocal] = data[flatGlobal]
			}
			incrementCounter(local, chunks)
		}
		if err := writeChunk(dir, coord, buf, comp); err != nil {
			return err
		}
		if ndim == 0 {
			done = true
		} else {
			done = !nextCounter(coord, nchunks)
		}
	}
	return nil
}

// ReadDataArrayZarr lit un DataArray[float64] depuis un store Zarr v2 (dir).
func ReadDataArrayZarr(dir string) (*DataArray[float64], error) {
	dims, shape, data, name, coords, err := readZarrArrayInternal(dir)
	if err != nil {
		return nil, err
	}
	return NewDataArray(dims, shape, data, coords, name)
}

// readZarrArrayInternal lit un array Zarr v2 et renvoie ses composants bruts
// (dimensions, forme, données, nom, coordonnées éventuelles issues de `.zattrs`).
func readZarrArrayInternal(dir string) (dims []string, shape []int, data []float64, name string, coords map[string][]float64, err error) {
	var meta zarrayMeta
	if err = readJSONFile(filepath.Join(dir, ".zarray"), &meta); err != nil {
		return nil, nil, nil, "", nil, err
	}
	if meta.ZarrFormat != 2 {
		return nil, nil, nil, "", nil, fmt.Errorf("xarray: seul Zarr v2 est pris en charge (format %d)", meta.ZarrFormat)
	}
	dt, err := parseZDtype(meta.Dtype)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}
	if meta.Order != "" && meta.Order != "C" {
		return nil, nil, nil, "", nil, fmt.Errorf("xarray: seul l'ordre C est pris en charge (%q)", meta.Order)
	}
	dec, err := newDecompressor(meta.Compressor)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	shape = meta.Shape
	chunks := meta.Chunks
	ndim := len(shape)
	fill, err := parseZarrFill(meta.FillValue)
	if err != nil {
		return nil, nil, nil, "", nil, err
	}

	data = make([]float64, product(shape))
	if fill != 0 {
		for i := range data {
			data[i] = fill
		}
	}

	nchunks := make([]int, ndim)
	for d := 0; d < ndim; d++ {
		nchunks[d] = (shape[d] + chunks[d] - 1) / chunks[d]
	}
	dataStrides := cOrderStrides(shape)
	chunkStrides := cOrderStrides(chunks)
	chunkSize := product(chunks)

	coord := make([]int, ndim)
	for done := false; !done; {
		buf, ok, rerr := readChunk(dir, coord, chunkSize, dec, dt)
		if rerr != nil {
			return nil, nil, nil, "", nil, rerr
		}
		if ok {
			local := make([]int, ndim)
			for l := 0; l < chunkSize; l++ {
				inBounds := true
				flatGlobal := 0
				for d := 0; d < ndim; d++ {
					g := coord[d]*chunks[d] + local[d]
					if g >= shape[d] {
						inBounds = false
						break
					}
					flatGlobal += g * dataStrides[d]
				}
				if inBounds {
					flatLocal := 0
					for d := 0; d < ndim; d++ {
						flatLocal += local[d] * chunkStrides[d]
					}
					data[flatGlobal] = buf[flatLocal]
				}
				incrementCounter(local, chunks)
			}
		}
		if ndim == 0 {
			done = true
		} else {
			done = !nextCounter(coord, nchunks)
		}
	}

	var attrs zattrsMeta
	if err = readJSONFile(filepath.Join(dir, ".zattrs"), &attrs); err != nil {
		return nil, nil, nil, "", nil, err
	}
	dims = attrs.Dims
	if len(dims) != ndim {
		// Repli : dimensions anonymes si .zattrs incomplet.
		dims = make([]string, ndim)
		for i := range dims {
			dims[i] = "dim_" + strconv.Itoa(i)
		}
	}
	return dims, shape, data, attrs.Name, attrs.Coords, nil
}

// --- Helpers ----------------------------------------------------------------

func cOrderStrides(shape []int) []int {
	st := make([]int, len(shape))
	acc := 1
	for i := len(shape) - 1; i >= 0; i-- {
		st[i] = acc
		acc *= shape[i]
	}
	return st
}

// incrementCounter incrémente un multi-indice borné par shape (ordre C).
func incrementCounter(counter, shape []int) {
	for k := len(counter) - 1; k >= 0; k-- {
		counter[k]++
		if counter[k] < shape[k] {
			return
		}
		counter[k] = 0
	}
}

// nextCounter incrémente le compteur et renvoie false s'il a rebouclé à zéro
// (fin du parcours).
func nextCounter(counter, shape []int) bool {
	for k := len(counter) - 1; k >= 0; k-- {
		counter[k]++
		if counter[k] < shape[k] {
			return true
		}
		counter[k] = 0
	}
	return false
}

func chunkKey(coord []int) string {
	if len(coord) == 0 {
		return "0"
	}
	parts := make([]string, len(coord))
	for i, c := range coord {
		parts[i] = strconv.Itoa(c)
	}
	return strings.Join(parts, ".")
}

func encodeF64LE(data []float64) []byte {
	buf := make([]byte, len(data)*8)
	for i, v := range data {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
	}
	return buf
}

func decodeF64LE(buf []byte, n int) ([]float64, error) {
	if len(buf) < n*8 {
		return nil, fmt.Errorf("xarray: chunk trop court (%d octets pour %d valeurs)", len(buf), n)
	}
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = math.Float64frombits(binary.LittleEndian.Uint64(buf[i*8:]))
	}
	return out, nil
}

// zdtype décrit un dtype numpy Zarr (ex. "<f8", "<i8", ">f4") : type (f/i/u),
// taille en octets, boutisme. Tous convertis en float64 à la lecture.
type zdtype struct {
	kind byte // 'f', 'i', 'u'
	size int
	be   bool
}

func parseZDtype(s string) (zdtype, error) {
	if len(s) < 3 {
		return zdtype{}, fmt.Errorf("xarray: dtype Zarr invalide %q", s)
	}
	var be bool
	switch s[0] {
	case '<', '|':
		be = false
	case '>':
		be = true
	default:
		return zdtype{}, fmt.Errorf("xarray: boutisme dtype %q non géré", s)
	}
	kind := s[1]
	size, err := strconv.Atoi(s[2:])
	if err != nil || size <= 0 {
		return zdtype{}, fmt.Errorf("xarray: taille dtype %q invalide", s)
	}
	switch kind {
	case 'f':
		if size != 4 && size != 8 {
			return zdtype{}, fmt.Errorf("xarray: flottant %q non géré (f4/f8)", s)
		}
	case 'i', 'u':
		if size != 1 && size != 2 && size != 4 && size != 8 {
			return zdtype{}, fmt.Errorf("xarray: entier %q non géré (1/2/4/8 octets)", s)
		}
	default:
		return zdtype{}, fmt.Errorf("xarray: dtype %q non géré (f/i/u)", s)
	}
	return zdtype{kind, size, be}, nil
}

// decode convertit n éléments du buffer brut en float64 selon le dtype.
func (dt zdtype) decode(buf []byte, n int) ([]float64, error) {
	if len(buf) < n*dt.size {
		return nil, fmt.Errorf("xarray: chunk trop court (%d octets pour %d valeurs de %d)", len(buf), n, dt.size)
	}
	var ord binary.ByteOrder = binary.LittleEndian
	if dt.be {
		ord = binary.BigEndian
	}
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		b := buf[i*dt.size:]
		switch {
		case dt.kind == 'f' && dt.size == 8:
			out[i] = math.Float64frombits(ord.Uint64(b))
		case dt.kind == 'f' && dt.size == 4:
			out[i] = float64(math.Float32frombits(ord.Uint32(b)))
		default: // entiers signés / non signés
			out[i] = decodeInt(b, dt.size, ord, dt.kind == 'i')
		}
	}
	return out, nil
}

func decodeInt(b []byte, size int, ord binary.ByteOrder, signed bool) float64 {
	var u uint64
	switch size {
	case 1:
		u = uint64(b[0])
	case 2:
		u = uint64(ord.Uint16(b))
	case 4:
		u = uint64(ord.Uint32(b))
	case 8:
		u = ord.Uint64(b)
	}
	if signed {
		bits := uint(size * 8)
		if bits < 64 && u&(1<<(bits-1)) != 0 {
			u |= ^uint64(0) << bits // extension de signe
		}
		return float64(int64(u))
	}
	return float64(u)
}

func writeChunk(dir string, coord []int, data []float64, comp ZarrComp) error {
	raw := encodeF64LE(data)
	if comp == ZarrZlib {
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		if _, err := w.Write(raw); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		raw = b.Bytes()
	}
	return os.WriteFile(filepath.Join(dir, chunkKey(coord)), raw, 0o644)
}

func readChunk(dir string, coord []int, n int, dec decompressor, dt zdtype) ([]float64, bool, error) {
	raw, err := os.ReadFile(filepath.Join(dir, chunkKey(coord)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil // chunk absent -> fill_value partout
		}
		return nil, false, err
	}
	if dec != nil {
		if raw, err = dec(raw); err != nil {
			return nil, false, err
		}
	}
	buf, err := dt.decode(raw, n)
	if err != nil {
		return nil, false, err
	}
	return buf, true, nil
}

// ZarrComp est un alias interne pour la lisibilité des helpers.
type ZarrComp = ZarrCompression

func writeJSONFile(path string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func readJSONFile(path string, v interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
