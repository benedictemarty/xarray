package xarray

import (
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

// WriteDatasetZarr écrit un Dataset[float64] comme groupe Zarr v2 dans dir.
// Chaque array est stocké en un seul chunk (taille = forme). comp choisit la
// compression.
func WriteDatasetZarr(dir string, ds *Dataset[float64], comp ZarrCompression) error {
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
		if err := writeZarrArrayInternal(sub, []string{name}, shape, cv.Data(), name, nil, shape, comp); err != nil {
			return err
		}
	}

	// Variables de données : un array chacune (un seul chunk).
	for _, name := range ds.VarNames() {
		da := ds.vars[name]
		sub := filepath.Join(dir, name)
		shape := da.variable.Shape()
		if err := writeZarrArrayInternal(sub, da.variable.Dims(), shape, da.variable.data, name, nil, shape, comp); err != nil {
			return err
		}
	}
	return nil
}

// ReadDatasetZarr lit un Dataset[float64] depuis un groupe Zarr v2 (dir). Les
// arrays 1D nommés comme une dimension sont interprétés comme des coordonnées ;
// les autres comme des variables de données.
func ReadDatasetZarr(dir string) (*Dataset[float64], error) {
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
