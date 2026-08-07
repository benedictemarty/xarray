// Package geoapi fournit des briques pour servir des données xarray-go à la
// manière d'un serveur OGC API (type pygeoapi) : export CoverageJSON,
// sous-échantillonnage géospatial, etc.
//
// Ce paquet ne dépend que de l'API publique de xarray-go.
package geoapi

import (
	"encoding/json"
	"fmt"

	"github.com/bmarty/xarray"
)

// Structures CoverageJSON (sous-ensemble « Grid »).

type covJSON struct {
	Type       string               `json:"type"` // "Coverage"
	Domain     domain               `json:"domain"`
	Parameters map[string]parameter `json:"parameters"`
	Ranges     map[string]ndArray   `json:"ranges"`
}

type domain struct {
	Type        string          `json:"type"`       // "Domain"
	DomainType  string          `json:"domainType"` // "Grid"
	Axes        map[string]axis `json:"axes"`
	Referencing []referencing   `json:"referencing"`
}

type axis struct {
	Values []float64 `json:"values"`
}

type referencing struct {
	Coordinates []string `json:"coordinates"`
	System      system   `json:"system"`
}

type system struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type parameter struct {
	Type             string           `json:"type"` // "Parameter"
	ObservedProperty observedProperty `json:"observedProperty"`
}

type observedProperty struct {
	Label map[string]string `json:"label"`
}

type ndArray struct {
	Type      string    `json:"type"` // "NdArray"
	DataType  string    `json:"dataType"`
	AxisNames []string  `json:"axisNames"`
	Shape     []int     `json:"shape"`
	Values    []float64 `json:"values"`
}

const crs84 = "http://www.opengis.net/def/crs/OGC/1.3/CRS84"

// ToCoverageJSON produit un document CoverageJSON (domaine Grid, CRS84) à partir
// d'un DataArray[float64] 2D dont les dimensions sont (yDim, xDim) — typiquement
// (latitude, longitude). param est le nom du paramètre exposé.
func ToCoverageJSON(da *xarray.DataArray[float64], param, xDim, yDim string) ([]byte, error) {
	dims := da.Dims()
	if len(dims) != 2 {
		return nil, fmt.Errorf("geoapi: CoverageJSON attend un DataArray 2D (%dD)", len(dims))
	}
	if dims[0] != yDim || dims[1] != xDim {
		return nil, fmt.Errorf("geoapi: dimensions attendues [%s %s], obtenues %v", yDim, xDim, dims)
	}
	xv, err := da.Coord(xDim)
	if err != nil {
		return nil, fmt.Errorf("geoapi: coordonnée %q requise : %w", xDim, err)
	}
	yv, err := da.Coord(yDim)
	if err != nil {
		return nil, fmt.Errorf("geoapi: coordonnée %q requise : %w", yDim, err)
	}
	shape := da.Shape() // [Ny, Nx]

	doc := covJSON{
		Type: "Coverage",
		Domain: domain{
			Type:       "Domain",
			DomainType: "Grid",
			Axes: map[string]axis{
				"x": {Values: xv},
				"y": {Values: yv},
			},
			Referencing: []referencing{{
				Coordinates: []string{"y", "x"},
				System:      system{Type: "GeographicCRS", ID: crs84},
			}},
		},
		Parameters: map[string]parameter{
			param: {
				Type:             "Parameter",
				ObservedProperty: observedProperty{Label: map[string]string{"en": param}},
			},
		},
		Ranges: map[string]ndArray{
			param: {
				Type:      "NdArray",
				DataType:  "float",
				AxisNames: []string{"y", "x"},
				Shape:     []int{shape[0], shape[1]},
				Values:    da.Data(),
			},
		},
	}
	return json.MarshalIndent(doc, "", "  ")
}
