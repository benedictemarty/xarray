package geoapi

import (
	"fmt"

	"github.com/bmarty/xarray"
)

// BBox est une emprise spatiale (coordonnées de la grille : longitudes X,
// latitudes Y).
type BBox struct {
	MinX, MinY, MaxX, MaxY float64
}

// SubsetBBox extrait le sous-cube d'un DataArray contenu dans l'emprise bb
// (sélection par plage sur les dimensions xDim et yDim). Équivalent d'une requête
// « area » / bbox d'OGC API.
func SubsetBBox(da *xarray.DataArray[float64], xDim, yDim string, bb BBox) (*xarray.DataArray[float64], error) {
	r, err := da.SelRange(xDim, bb.MinX, bb.MaxX)
	if err != nil {
		return nil, fmt.Errorf("geoapi: bbox sur %q : %w", xDim, err)
	}
	r, err = r.SelRange(yDim, bb.MinY, bb.MaxY)
	if err != nil {
		return nil, fmt.Errorf("geoapi: bbox sur %q : %w", yDim, err)
	}
	return r, nil
}

// Position renvoie la valeur au point (x, y) le plus proche — l'équivalent d'une
// requête « position » d'OGC API - EDR sur une grille 2D (xDim, yDim).
func Position(da *xarray.DataArray[float64], xDim, yDim string, x, y float64) (float64, error) {
	r, err := da.SelNearest(xDim, x)
	if err != nil {
		return 0, fmt.Errorf("geoapi: position sur %q : %w", xDim, err)
	}
	r, err = r.SelNearest(yDim, y)
	if err != nil {
		return 0, fmt.Errorf("geoapi: position sur %q : %w", yDim, err)
	}
	d := r.Data()
	if len(d) != 1 {
		return 0, fmt.Errorf("geoapi: la sélection ne réduit pas à un point (taille %d)", len(d))
	}
	return d[0], nil
}
