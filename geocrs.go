package xarray

import (
	"fmt"
	"math"
	"strings"
)

// Transformation de coordonnées entre systèmes de référence courants, en
// formules fermées et sans dépendance (pas de PROJ). Périmètre volontairement
// restreint : WGS84 géographique (EPSG:4326) ↔ Web Mercator sphérique
// (EPSG:3857), la paire de loin la plus utilisée (tuiles web, fonds de carte).
//
// Ce n'est PAS un moteur de reprojection généraliste : les autres CRS
// (projections coniques, géostationnaire MTG, transformations de datum…)
// nécessitent PROJ. Il s'agit d'une transformation ponctuelle de coordonnées,
// pas d'un rééchantillonnage de grille.

// webMercatorR est le rayon de la sphère utilisée par EPSG:3857 (demi-grand axe
// WGS84), en mètres.
const webMercatorR = 6378137.0

// WebMercatorForward projette (lon, lat) en degrés (EPSG:4326) vers (x, y) en
// mètres (EPSG:3857).
func WebMercatorForward(lon, lat float64) (x, y float64) {
	x = webMercatorR * lon * math.Pi / 180
	y = webMercatorR * math.Log(math.Tan(math.Pi/4+(lat*math.Pi/180)/2))
	return x, y
}

// WebMercatorInverse projette (x, y) en mètres (EPSG:3857) vers (lon, lat) en
// degrés (EPSG:4326).
func WebMercatorInverse(x, y float64) (lon, lat float64) {
	lon = x / webMercatorR * 180 / math.Pi
	lat = (2*math.Atan(math.Exp(y/webMercatorR)) - math.Pi/2) * 180 / math.Pi
	return lon, lat
}

// normalizeCRS ramène quelques écritures usuelles d'un CRS à un code EPSG.
func normalizeCRS(crs string) string {
	c := strings.ToUpper(strings.TrimSpace(crs))
	c = strings.TrimPrefix(c, "EPSG:")
	switch c {
	case "4326", "CRS84", "WGS84", "OGC:CRS84":
		return "4326"
	case "3857", "900913", "3785":
		return "3857"
	}
	return c
}

// TransformXY transforme un couple de coordonnées de fromCRS vers toCRS. Les
// paires prises en charge : EPSG:4326 ↔ EPSG:3857 (et l'identité). Renvoie une
// erreur pour toute autre paire (utiliser PROJ pour les CRS non gérés).
//
// Convention x/y : en 4326, x = longitude, y = latitude (degrés) ; en 3857,
// (x, y) en mètres.
func TransformXY(fromCRS, toCRS string, x, y float64) (float64, float64, error) {
	from, to := normalizeCRS(fromCRS), normalizeCRS(toCRS)
	switch {
	case from == to:
		return x, y, nil
	case from == "4326" && to == "3857":
		nx, ny := WebMercatorForward(x, y)
		return nx, ny, nil
	case from == "3857" && to == "4326":
		nlon, nlat := WebMercatorInverse(x, y)
		return nlon, nlat, nil
	default:
		return 0, 0, fmt.Errorf("xarray: transformation %s→%s non prise en charge (seul 4326↔3857 ; utilisez PROJ pour les autres)", fromCRS, toCRS)
	}
}
