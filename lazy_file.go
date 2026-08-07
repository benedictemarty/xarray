package xarray

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// Source hors-mémoire : un fichier binaire de float64 (little-endian, ordre C).
// Chaque chunk est lu à la demande par seek+read, donc on ne charge jamais tout
// le tableau — ce qui permet de traiter des données plus grandes que la RAM.

type fileSource struct {
	path      string
	dims      []string
	shape     []int
	coords    map[string][]float64
	chunkSize int
}

func (f *fileSource) Dims() []string               { return append([]string(nil), f.dims...) }
func (f *fileSource) Shape() []int                 { return append([]int(nil), f.shape...) }
func (f *fileSource) Coords() map[string][]float64 { return f.coords }
func (f *fileSource) rowLen() int                  { return product(f.shape[1:]) }
func (f *fileSource) NumChunks() int {
	n := f.shape[0]
	return (n + f.chunkSize - 1) / f.chunkSize
}
func (f *fileSource) ChunkRows(i int) (int, int) {
	start := i * f.chunkSize
	end := start + f.chunkSize
	if end > f.shape[0] {
		end = f.shape[0]
	}
	return start, end
}
func (f *fileSource) ChunkData(i int) ([]float64, error) {
	s, e := f.ChunkRows(i)
	rl := f.rowLen()
	n := (e - s) * rl

	fh, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	if _, err := fh.Seek(int64(s*rl*8), 0); err != nil {
		return nil, err
	}
	buf := make([]byte, n*8)
	if _, err := fh.Read(buf); err != nil {
		return nil, err
	}
	out := make([]float64, n)
	for j := 0; j < n; j++ {
		out[j] = math.Float64frombits(binary.LittleEndian.Uint64(buf[j*8:]))
	}
	return out, nil
}

// ChunkFile crée un LazyArray adossé à un fichier binaire de float64 (ordre C),
// découpé en blocs de chunkSize lignes le long de l'axe 0. Les données ne sont
// jamais chargées entièrement en mémoire (hors-mémoire / out-of-core).
func ChunkFile(path string, dims []string, shape []int, coords map[string][]float64, chunkSize int) (*LazyArray, error) {
	if chunkSize < 1 {
		return nil, fmt.Errorf("xarray: taille de chunk invalide %d", chunkSize)
	}
	if len(shape) == 0 {
		return nil, fmt.Errorf("xarray: forme vide")
	}
	src := &fileSource{
		path: path, dims: dims, shape: shape,
		coords: coords, chunkSize: chunkSize,
	}
	return &LazyArray{src: src}, nil
}

// WriteRawF64 écrit des données float64 dans un fichier binaire (little-endian,
// ordre C), lisible par ChunkFile.
func WriteRawF64(path string, data []float64) error {
	buf := make([]byte, len(data)*8)
	for i, v := range data {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
	}
	return os.WriteFile(path, buf, 0o644)
}
