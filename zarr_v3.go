package xarray

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Lecture du format Zarr v3 (zarr.json par nœud, clés de chunk « c/0/0 »,
// pipeline de codecs). Complète le lecteur v2. Le point d'entrée public
// (ReadDataArrayZarr/ReadDatasetZarr) détecte le format et route ici.

type zarrV3Array struct {
	ZarrFormat int    `json:"zarr_format"`
	NodeType   string `json:"node_type"`
	Shape      []int  `json:"shape"`
	DataType   string `json:"data_type"`
	ChunkGrid  struct {
		Configuration struct {
			ChunkShape []int `json:"chunk_shape"`
		} `json:"configuration"`
	} `json:"chunk_grid"`
	ChunkKeyEncoding struct {
		Name          string `json:"name"`
		Configuration struct {
			Separator string `json:"separator"`
		} `json:"configuration"`
	} `json:"chunk_key_encoding"`
	FillValue      json.RawMessage `json:"fill_value"`
	Codecs         []zarrV3Codec   `json:"codecs"`
	DimensionNames []string        `json:"dimension_names"`
}

type zarrV3Codec struct {
	Name          string          `json:"name"`
	Configuration json.RawMessage `json:"configuration"`
}

// isZarrV3 indique si dir est un nœud Zarr v3 (présence de zarr.json).
func isZarrV3(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "zarr.json"))
	return err == nil
}

// v3DataType convertit un data_type v3 (« float64 », « int32 », …) en zdtype.
// L'endianité est fournie séparément par le codec « bytes ».
func v3DataType(s string, be bool) (zdtype, error) {
	switch s {
	case "float64":
		return zdtype{'f', 8, be}, nil
	case "float32":
		return zdtype{'f', 4, be}, nil
	case "int64":
		return zdtype{'i', 8, be}, nil
	case "int32":
		return zdtype{'i', 4, be}, nil
	case "int16":
		return zdtype{'i', 2, be}, nil
	case "int8":
		return zdtype{'i', 1, be}, nil
	case "uint64":
		return zdtype{'u', 8, be}, nil
	case "uint32":
		return zdtype{'u', 4, be}, nil
	case "uint16":
		return zdtype{'u', 2, be}, nil
	case "uint8":
		return zdtype{'u', 1, be}, nil
	default:
		return zdtype{}, fmt.Errorf("xarray: data_type Zarr v3 %q non pris en charge", s)
	}
}

// v3Pipeline construit le dtype (endianité) et le décompresseur à partir de la
// liste de codecs v3 : « bytes » (endianité), puis une compression éventuelle
// (« zstd », « blosc », « gzip »).
func v3Pipeline(dataType string, codecs []zarrV3Codec) (zdtype, decompressor, error) {
	be := false
	var dec decompressor
	for _, c := range codecs {
		switch c.Name {
		case "bytes":
			var cfg struct {
				Endian string `json:"endian"`
			}
			_ = json.Unmarshal(c.Configuration, &cfg)
			be = cfg.Endian == "big"
		case "zstd":
			dec = func(src []byte) ([]byte, error) { return zstdDec.DecodeAll(src, nil) }
		case "gzip":
			dec = func(src []byte) ([]byte, error) {
				r, err := gzip.NewReader(strings.NewReader(string(src)))
				if err != nil {
					return nil, err
				}
				defer r.Close()
				return io.ReadAll(r)
			}
		case "blosc":
			var cfg struct {
				Cname string `json:"cname"`
			}
			_ = json.Unmarshal(c.Configuration, &cfg)
			cname := cfg.Cname
			dec = func(src []byte) ([]byte, error) { return bloscDecompress(src, cname) }
		case "crc32c":
			// Codec de contrôle d'intégrité : ignoré (les 4 octets de checksum en
			// fin de chunk le seraient aussi ; non géré ici).
			return zdtype{}, nil, fmt.Errorf("xarray: codec Zarr v3 crc32c non pris en charge")
		default:
			return zdtype{}, nil, fmt.Errorf("xarray: codec Zarr v3 %q non pris en charge", c.Name)
		}
	}
	dt, err := v3DataType(dataType, be)
	return dt, dec, err
}

// v3ChunkKey construit la clé d'un chunk (encodage « default » : préfixe « c »
// puis indices joints par le séparateur ; ou « v2 » : indices seuls).
func v3ChunkKey(coord []int, name, sep string) string {
	if sep == "" {
		sep = "/"
	}
	parts := make([]string, 0, len(coord)+1)
	if name != "v2" {
		parts = append(parts, "c")
	}
	if len(coord) == 0 {
		parts = append(parts, "0") // tableau scalaire : clé « c/0 »
	}
	for _, c := range coord {
		parts = append(parts, strconv.Itoa(c))
	}
	return strings.Join(parts, sep)
}

// readZarrV3Array lit un array Zarr v3 et renvoie ses composants bruts.
func readZarrV3Array(dir string) (dims []string, shape []int, data []float64, name string, coords map[string][]float64, err error) {
	var meta zarrV3Array
	if err = readJSONFile(filepath.Join(dir, "zarr.json"), &meta); err != nil {
		return nil, nil, nil, "", nil, err
	}
	if meta.ZarrFormat != 3 {
		return nil, nil, nil, "", nil, fmt.Errorf("xarray: zarr_format %d inattendu (v3 attendu)", meta.ZarrFormat)
	}
	dt, dec, perr := v3Pipeline(meta.DataType, meta.Codecs)
	if perr != nil {
		return nil, nil, nil, "", nil, perr
	}
	shape = meta.Shape
	chunks := meta.ChunkGrid.Configuration.ChunkShape
	if len(chunks) != len(shape) {
		return nil, nil, nil, "", nil, fmt.Errorf("xarray: chunk_shape incohérent avec shape")
	}
	fill, ferr := parseZarrFill(meta.FillValue)
	if ferr != nil {
		return nil, nil, nil, "", nil, ferr
	}
	ndim := len(shape)
	data = make([]float64, product(shape))
	if fill != 0 {
		for i := range data {
			data[i] = fill
		}
	}

	sep := meta.ChunkKeyEncoding.Configuration.Separator
	keyName := meta.ChunkKeyEncoding.Name

	nchunks := make([]int, ndim)
	for d := 0; d < ndim; d++ {
		nchunks[d] = (shape[d] + chunks[d] - 1) / chunks[d]
	}
	dataStrides := cOrderStrides(shape)
	chunkStrides := cOrderStrides(chunks)
	chunkSize := product(chunks)

	coord := make([]int, ndim)
	for done := false; !done; {
		buf, ok, rerr := readChunkV3(dir, v3ChunkKey(coord, keyName, sep), chunkSize, dec, dt)
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

	dims = meta.DimensionNames
	if len(dims) != ndim {
		dims = make([]string, ndim)
		for i := range dims {
			dims[i] = "dim_" + strconv.Itoa(i)
		}
	}
	return dims, shape, data, filepath.Base(dir), nil, nil
}

func readChunkV3(dir, key string, n int, dec decompressor, dt zdtype) ([]float64, bool, error) {
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(key)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
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

// readZarrV3Dataset lit un groupe Zarr v3 (sous-répertoires portant un zarr.json
// de type array) et reconstruit un Dataset, en réattachant les coordonnées.
func readZarrV3Dataset(dir string) (*Dataset[float64], error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	type arr struct {
		dims  []string
		shape []int
		data  []float64
	}
	arrays := map[string]arr{}
	for _, e := range entries {
		if !e.IsDir() || !isZarrV3(filepath.Join(dir, e.Name())) {
			continue
		}
		dims, shape, data, _, _, rerr := readZarrV3Array(filepath.Join(dir, e.Name()))
		if rerr != nil {
			return nil, fmt.Errorf("xarray: lecture de l'array %q : %w", e.Name(), rerr)
		}
		arrays[e.Name()] = arr{dims: dims, shape: shape, data: data}
	}
	coordLabels := map[string][]float64{}
	for key, a := range arrays {
		if len(a.dims) == 1 && a.dims[0] == key {
			coordLabels[key] = a.data
		}
	}
	vars := map[string]*DataArray[float64]{}
	for key, a := range arrays {
		if _, isCoord := coordLabels[key]; isCoord {
			continue
		}
		coords := map[string][]float64{}
		for _, d := range a.dims {
			if lbl, ok := coordLabels[d]; ok {
				coords[d] = lbl
			}
		}
		da, derr := NewDataArray(a.dims, a.shape, a.data, coords, key)
		if derr != nil {
			return nil, derr
		}
		vars[key] = da
	}
	if len(vars) == 0 {
		return nil, fmt.Errorf("xarray: aucun tableau de données dans le groupe Zarr v3 %q", dir)
	}
	return NewDataset(vars)
}
