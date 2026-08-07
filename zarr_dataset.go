package xarray

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Prise en charge d'un Dataset comme groupe Zarr v2 : un répertoire contenant un
// fichier `.zgroup`, puis un sous-répertoire array par coordonnée et par variable
// de données (convention xarray : chaque array porte `_ARRAY_DIMENSIONS`, les
// coordonnées sont des arrays 1D nommés comme leur dimension).

type zgroupMeta struct {
	ZarrFormat int `json:"zarr_format"`
}

// chunkShapeFor calcule la forme de chunk d'un array à partir d'une spec
// « dimension → taille de chunk ». Une dimension absente de la spec (ou spec nil)
// n'est pas découpée (chunk = taille de la dimension). Les tailles sont bornées à
// [1, taille de la dimension].
func chunkShapeFor(dims []string, shape []int, spec map[string]int) []int {
	cs := make([]int, len(shape))
	for i, d := range dims {
		cs[i] = shape[i]
		if v, ok := spec[d]; ok && v > 0 && v < shape[i] {
			cs[i] = v
		}
		if cs[i] < 1 {
			cs[i] = 1
		}
	}
	return cs
}

// WriteDatasetZarr écrit un Dataset[float64] comme groupe Zarr v2 dans dir.
// Chaque array est stocké en un seul chunk (taille = forme). comp choisit la
// compression.
func WriteDatasetZarr(dir string, ds *Dataset[float64], comp ZarrCompression) error {
	return WriteDatasetZarrChunked(dir, ds, nil, comp)
}

// WriteDatasetZarrChunked écrit un Dataset[float64] en Zarr v2 avec un découpage
// configurable : chunks associe un nom de dimension à une taille de chunk (façon
// ds.chunk({...}) de xarray). Les dimensions absentes ne sont pas découpées.
// Permet des accès partiels efficaces sur de grands tableaux.
func WriteDatasetZarrChunked(dir string, ds *Dataset[float64], chunks map[string]int, comp ZarrCompression) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeJSONFile(filepath.Join(dir, ".zgroup"), zgroupMeta{ZarrFormat: 2}); err != nil {
		return err
	}

	// Coordonnées partagées : arrays 1D nommés comme leur dimension.
	for name, cv := range ds.coords {
		sub := filepath.Join(dir, name)
		shape := cv.Shape()
		cs := chunkShapeFor([]string{name}, shape, chunks)
		if err := writeZarrArrayInternal(sub, []string{name}, shape, cv.Data(), name, nil, cs, comp); err != nil {
			return err
		}
	}

	// Variables de données.
	for _, name := range ds.VarNames() {
		da := ds.vars[name]
		sub := filepath.Join(dir, name)
		shape := da.variable.Shape()
		cs := chunkShapeFor(da.variable.Dims(), shape, chunks)
		if err := writeZarrArrayInternal(sub, da.variable.Dims(), shape, da.variable.data, name, nil, cs, comp); err != nil {
			return err
		}
	}

	// Métadonnées consolidées (.zmetadata) : permet à zarr-python/xarray d'ouvrir
	// le store en une seule lecture, sans parcourir l'arborescence, et sans le
	// RuntimeWarning « consolidated metadata not found ».
	return consolidateZarrMetadata(dir)
}

// zconsolidated est le contenu de .zmetadata (Zarr v2, consolidated_format 1).
type zconsolidated struct {
	ZarrConsolidatedFormat int                        `json:"zarr_consolidated_format"`
	Metadata               map[string]json.RawMessage `json:"metadata"`
}

// consolidateZarrMetadata agrège tous les .zgroup/.zarray/.zattrs du store en un
// unique .zmetadata à la racine (clés = chemins relatifs en « / »).
func consolidateZarrMetadata(dir string) error {
	meta := map[string]json.RawMessage{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		switch info.Name() {
		case ".zgroup", ".zarray", ".zattrs":
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			meta[filepath.ToSlash(rel)] = json.RawMessage(raw)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return writeJSONFile(filepath.Join(dir, ".zmetadata"), zconsolidated{ZarrConsolidatedFormat: 1, Metadata: meta})
}

// ReadDatasetZarr lit un Dataset[float64] depuis un groupe Zarr v2 (dir). Les
// arrays 1D nommés comme une dimension sont interprétés comme des coordonnées ;
// les autres comme des variables de données.
func ReadDatasetZarr(dir string) (*Dataset[float64], error) {
	if isZarrV3(dir) {
		return readZarrV3Dataset(dir)
	}
	if err := readJSONFile(filepath.Join(dir, ".zgroup"), &zgroupMeta{}); err != nil {
		return nil, fmt.Errorf("xarray: %q n'est pas un groupe Zarr (.zgroup absent) : %w", dir, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type arr struct {
		dims  []string
		shape []int
		data  []float64
		name  string
	}
	arrays := map[string]arr{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sub := filepath.Join(dir, e.Name())
		if _, statErr := os.Stat(filepath.Join(sub, ".zarray")); statErr != nil {
			continue // pas un array Zarr
		}
		dims, shape, data, name, _, rerr := readZarrArrayInternal(sub)
		if rerr != nil {
			return nil, fmt.Errorf("xarray: lecture de l'array %q : %w", e.Name(), rerr)
		}
		if name == "" {
			name = e.Name()
		}
		arrays[e.Name()] = arr{dims: dims, shape: shape, data: data, name: name}
	}

	// Coordonnées = arrays 1D dont le nom correspond à leur unique dimension.
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
		da, derr := NewDataArray(a.dims, a.shape, a.data, coords, a.name)
		if derr != nil {
			return nil, fmt.Errorf("xarray: reconstruction de %q : %w", key, derr)
		}
		vars[key] = da
	}
	return NewDataset(vars)
}
