package xarray

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
)

// Écriture au format Zarr v3 (zarr.json par nœud, clés de chunk « c/0/0 »,
// pipeline de codecs bytes + compression). Symétrique du lecteur zarr_v3.go.
// Les données sont écrites en float64 little-endian ; fill_value = "NaN".

type zarrV3ArrayMeta struct {
	ZarrFormat       int                    `json:"zarr_format"`
	NodeType         string                 `json:"node_type"`
	Shape            []int                  `json:"shape"`
	DataType         string                 `json:"data_type"`
	ChunkGrid        map[string]interface{} `json:"chunk_grid"`
	ChunkKeyEncoding map[string]interface{} `json:"chunk_key_encoding"`
	FillValue        interface{}            `json:"fill_value"`
	Codecs           []interface{}          `json:"codecs"`
	Attributes       map[string]interface{} `json:"attributes"`
	DimensionNames   []string               `json:"dimension_names,omitempty"`
}

type zarrV3GroupMeta struct {
	ZarrFormat int                    `json:"zarr_format"`
	NodeType   string                 `json:"node_type"`
	Attributes map[string]interface{} `json:"attributes"`
}

// v3Codecs construit la liste de codecs (bytes little-endian + compression).
func v3Codecs(comp ZarrCompression) []interface{} {
	codecs := []interface{}{
		map[string]interface{}{"name": "bytes", "configuration": map[string]interface{}{"endian": "little"}},
	}
	switch comp {
	case ZarrZstd:
		codecs = append(codecs, map[string]interface{}{
			"name": "zstd", "configuration": map[string]interface{}{"level": 5, "checksum": false}})
	case ZarrZlib:
		codecs = append(codecs, map[string]interface{}{
			"name": "gzip", "configuration": map[string]interface{}{"level": 5}})
	}
	return codecs
}

// WriteDataArrayZarrV3 écrit un DataArray au format Zarr v3 dans dir.
func WriteDataArrayZarrV3(dir string, da *DataArray[float64], chunks []int, comp ZarrCompression) error {
	coords := map[string][]float64{}
	for _, d := range da.variable.Dims() {
		if cv, ok := da.coords[d]; ok {
			coords[d] = cv.data
		}
	}
	// Un groupe contenant l'array et ses coordonnées (comme xarray).
	if err := writeV3Group(dir); err != nil {
		return err
	}
	if err := writeV3Array(filepath.Join(dir, da.name), da.variable.Dims(),
		da.variable.Shape(), da.variable.data, chunks, comp); err != nil {
		return err
	}
	for name, vals := range coords {
		if err := writeV3Array(filepath.Join(dir, name), []string{name},
			[]int{len(vals)}, vals, []int{len(vals)}, comp); err != nil {
			return err
		}
	}
	return nil
}

// WriteDatasetZarrV3 écrit un Dataset comme groupe Zarr v3 (un array par
// coordonnée et par variable, chacun en un seul chunk).
func WriteDatasetZarrV3(dir string, ds *Dataset[float64], comp ZarrCompression) error {
	if err := writeV3Group(dir); err != nil {
		return err
	}
	for name, cv := range ds.coords {
		shape := cv.Shape()
		if err := writeV3Array(filepath.Join(dir, name), []string{name}, shape, cv.data, shape, comp); err != nil {
			return err
		}
	}
	for _, name := range ds.VarNames() {
		da := ds.vars[name]
		shape := da.variable.Shape()
		if err := writeV3Array(filepath.Join(dir, name), da.variable.Dims(), shape, da.variable.data, shape, comp); err != nil {
			return err
		}
	}
	return nil
}

func writeV3Group(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return writeJSONFile(filepath.Join(dir, "zarr.json"), zarrV3GroupMeta{
		ZarrFormat: 3, NodeType: "group", Attributes: map[string]interface{}{},
	})
}

func writeV3Array(dir string, dims []string, shape []int, data []float64, chunks []int, comp ZarrCompression) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	meta := zarrV3ArrayMeta{
		ZarrFormat: 3,
		NodeType:   "array",
		Shape:      shape,
		DataType:   "float64",
		ChunkGrid: map[string]interface{}{
			"name": "regular", "configuration": map[string]interface{}{"chunk_shape": chunks},
		},
		ChunkKeyEncoding: map[string]interface{}{
			"name": "default", "configuration": map[string]interface{}{"separator": "/"},
		},
		FillValue:      "NaN",
		Codecs:         v3Codecs(comp),
		Attributes:     map[string]interface{}{},
		DimensionNames: dims,
	}
	if err := writeJSONFile(filepath.Join(dir, "zarr.json"), meta); err != nil {
		return err
	}

	// Chunks (identique à v2 pour le découpage ; clé « c/0/0 » et compression v3).
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
		buf := make([]float64, chunkSize)
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
		if err := writeChunkV3(dir, v3ChunkKey(coord, "default", "/"), buf, comp); err != nil {
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

func writeChunkV3(dir, key string, data []float64, comp ZarrCompression) error {
	raw := encodeF64LE(data)
	switch comp {
	case ZarrZstd:
		raw = zstdEnc.EncodeAll(raw, nil)
	case ZarrZlib:
		raw = gzipEncode(raw)
	}
	path := filepath.Join(dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func gzipEncode(raw []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write(raw)
	_ = w.Close()
	return b.Bytes()
}
