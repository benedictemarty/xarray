package xarray

import (
	"fmt"
	"path/filepath"
	"strconv"
)

// Pyramides multi-échelles (overviews) au format Zarr : un groupe racine
// contenant des sous-groupes numérotés « 0 », « 1 »… du plus fin au plus grossier
// (chaque niveau = moyenne de blocs `factor`×`factor` du précédent), plus une
// métadonnée « multiscales » à la racine décrivant les niveaux.
//
// Convention minimale et auto-descriptive (inspirée d'OME-Zarr / xarray-multiscale),
// pensée pour la visualisation zoomable et l'accès rapide à basse résolution ;
// elle ne prétend pas à la conformité NGFF complète.

type pyramidLevel struct {
	Path   string `json:"path"`
	Factor int    `json:"factor"` // facteur de sous-échantillonnage vs niveau 0
}

type pyramidMultiscale struct {
	Type     string         `json:"type"` // "reduce/mean"
	Datasets []pyramidLevel `json:"datasets"`
}

type pyramidAttrs struct {
	Multiscales []pyramidMultiscale `json:"multiscales"`
}

// WritePyramidZarr écrit une pyramide multi-échelles d'un raster 2D (dimensions
// [yDim, xDim]). nlevels niveaux sont produits (≥1) ; chaque niveau suivant est
// une réduction par moyenne de blocs factor×factor (factor ≥ 2) sur y puis x.
// Les niveaux sont des groupes Zarr « 0 »…« nlevels-1 » sous dir.
func WritePyramidZarr(dir string, da *DataArray[float64], yDim, xDim string, nlevels, factor int, comp ZarrCompression) error {
	if nlevels < 1 {
		return fmt.Errorf("xarray: nlevels doit être ≥ 1")
	}
	if factor < 2 {
		return fmt.Errorf("xarray: factor doit être ≥ 2")
	}
	dims := da.variable.Dims()
	if len(dims) != 2 || dims[0] != yDim || dims[1] != xDim {
		return fmt.Errorf("xarray: pyramide attend des dimensions [%q, %q], obtenu %v", yDim, xDim, dims)
	}

	levels := []pyramidLevel{}
	cur := da
	cumFactor := 1
	for k := 0; k < nlevels; k++ {
		ds, err := NewDataset(map[string]*DataArray[float64]{da.name: cur})
		if err != nil {
			return err
		}
		sub := filepath.Join(dir, strconv.Itoa(k))
		if err := WriteDatasetZarr(sub, ds, comp); err != nil {
			return err
		}
		levels = append(levels, pyramidLevel{Path: strconv.Itoa(k), Factor: cumFactor})

		// Niveau suivant : réduction par moyenne de blocs sur y puis x.
		if k < nlevels-1 {
			ry, err := cur.Coarsen(yDim, factor)
			if err != nil {
				return err
			}
			half, err := ry.Mean()
			if err != nil {
				return err
			}
			rx, err := half.Coarsen(xDim, factor)
			if err != nil {
				return err
			}
			cur, err = rx.Mean()
			if err != nil {
				return err
			}
			cumFactor *= factor
		}
	}

	// Racine du groupe pyramide : .zgroup + .zattrs (multiscales).
	if err := writeJSONFile(filepath.Join(dir, ".zgroup"), zgroupMeta{ZarrFormat: 2}); err != nil {
		return err
	}
	attrs := pyramidAttrs{Multiscales: []pyramidMultiscale{{Type: "reduce/mean", Datasets: levels}}}
	return writeJSONFile(filepath.Join(dir, ".zattrs"), attrs)
}

// ReadPyramidLevel lit un niveau donné d'une pyramide écrite par WritePyramidZarr
// (niveau 0 = pleine résolution). Renvoie le Dataset de ce niveau.
func ReadPyramidLevel(dir string, level int) (*Dataset[float64], error) {
	return ReadDatasetZarr(filepath.Join(dir, strconv.Itoa(level)))
}

// PyramidLevels lit la métadonnée « multiscales » d'une pyramide et renvoie la
// liste des niveaux (chemin + facteur).
func PyramidLevels(dir string) ([]pyramidLevel, error) {
	var attrs pyramidAttrs
	if err := readJSONFile(filepath.Join(dir, ".zattrs"), &attrs); err != nil {
		return nil, err
	}
	if len(attrs.Multiscales) == 0 {
		return nil, fmt.Errorf("xarray: aucune métadonnée multiscales dans %q", dir)
	}
	return attrs.Multiscales[0].Datasets, nil
}
