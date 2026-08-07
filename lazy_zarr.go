package xarray

import (
	"fmt"
	"path/filepath"
)

// Source hors-mémoire adossée à un store Zarr v2 (tableaux 1D/2D, float64).
// Chaque bloc de lignes est reconstruit en ne lisant que les chunks Zarr qui le
// recouvrent — on ne charge donc jamais tout le tableau.

type zarrRowSource struct {
	dir       string
	shape     []int
	chunks    []int
	dec       decompressor
	dims      []string
	coords    map[string][]float64
	chunkSize int // lignes par chunk lazy (axe 0)
}

func (z *zarrRowSource) Dims() []string               { return append([]string(nil), z.dims...) }
func (z *zarrRowSource) Shape() []int                 { return append([]int(nil), z.shape...) }
func (z *zarrRowSource) Coords() map[string][]float64 { return z.coords }
func (z *zarrRowSource) NumChunks() int {
	return (z.shape[0] + z.chunkSize - 1) / z.chunkSize
}
func (z *zarrRowSource) ChunkRows(i int) (int, int) {
	start := i * z.chunkSize
	end := start + z.chunkSize
	if end > z.shape[0] {
		end = z.shape[0]
	}
	return start, end
}

func (z *zarrRowSource) ChunkData(i int) ([]float64, error) {
	s, e := z.ChunkRows(i)
	return z.readBlock(s, e)
}

// readBlock reconstruit les lignes [rowStart, rowEnd) en lisant les chunks Zarr
// nécessaires.
func (z *zarrRowSource) readBlock(rowStart, rowEnd int) ([]float64, error) {
	nRows := rowEnd - rowStart
	rl := product(z.shape[1:]) // 1 en 1D, C en 2D
	out := make([]float64, nRows*rl)

	cr := z.chunks[0]
	rcStart, rcEnd := rowStart/cr, (rowEnd-1)/cr

	if len(z.shape) == 1 {
		for rc := rcStart; rc <= rcEnd; rc++ {
			cd, ok, err := readChunk(z.dir, []int{rc}, cr, z.dec)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			for li := 0; li < cr; li++ {
				gr := rc*cr + li
				if gr < rowStart || gr >= rowEnd || gr >= z.shape[0] {
					continue
				}
				out[gr-rowStart] = cd[li]
			}
		}
		return out, nil
	}

	// 2D
	C := z.shape[1]
	cc := z.chunks[1]
	nChunkSize := cr * cc
	ncCol := (C + cc - 1) / cc
	for rc := rcStart; rc <= rcEnd; rc++ {
		for cci := 0; cci < ncCol; cci++ {
			cd, ok, err := readChunk(z.dir, []int{rc, cci}, nChunkSize, z.dec)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			for li := 0; li < cr; li++ {
				gr := rc*cr + li
				if gr < rowStart || gr >= rowEnd || gr >= z.shape[0] {
					continue
				}
				for lj := 0; lj < cc; lj++ {
					gc := cci*cc + lj
					if gc >= C {
						continue
					}
					out[(gr-rowStart)*C+gc] = cd[li*cc+lj]
				}
			}
		}
	}
	return out, nil
}

// ChunkZarr crée un LazyArray hors-mémoire adossé à un store Zarr v2 (tableau
// 1D ou 2D, dtype <f8). Les chunks sont lus à la demande. chunkSize est le
// nombre de lignes (axe 0) par bloc lazy.
func ChunkZarr(dir string, chunkSize int) (*LazyArray, error) {
	if chunkSize < 1 {
		return nil, fmt.Errorf("xarray: taille de chunk invalide %d", chunkSize)
	}
	var meta zarrayMeta
	if err := readJSONFile(filepath.Join(dir, ".zarray"), &meta); err != nil {
		return nil, err
	}
	if meta.ZarrFormat != 2 {
		return nil, fmt.Errorf("xarray: seul Zarr v2 est pris en charge")
	}
	if meta.Dtype != "<f8" {
		return nil, fmt.Errorf("xarray: seul le dtype \"<f8\" est pris en charge (%q)", meta.Dtype)
	}
	if len(meta.Shape) < 1 || len(meta.Shape) > 2 {
		return nil, fmt.Errorf("xarray: ChunkZarr gère les tableaux 1D/2D (%dD)", len(meta.Shape))
	}
	dec, err := newDecompressor(meta.Compressor)
	if err != nil {
		return nil, err
	}

	var attrs zattrsMeta
	_ = readJSONFile(filepath.Join(dir, ".zattrs"), &attrs) // dims/coords optionnels
	dims := attrs.Dims
	if len(dims) != len(meta.Shape) {
		dims = make([]string, len(meta.Shape))
		for i := range dims {
			dims[i] = fmt.Sprintf("dim_%d", i)
		}
	}

	src := &zarrRowSource{
		dir: dir, shape: meta.Shape, chunks: meta.Chunks, dec: dec,
		dims: dims, coords: attrs.Coords, chunkSize: chunkSize,
	}
	return &LazyArray{src: src, name: attrs.Name}, nil
}
