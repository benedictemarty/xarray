package xarray

import (
	"fmt"
	"strconv"
	"strings"
)

// Géoréférencement raster : géotransformation affine (pixel ↔ monde) et système
// de coordonnées de référence (CRS) traité comme identifiant opaque
// (EPSG:xxxx / WKT / proj). Il n'y a PAS de reprojection ici — seulement le
// mapping pixel↔coordonnées et le transport du CRS en métadonnée.
//
// Convention identique à la bibliothèque `affine` (rasterio/GDAL) :
//
//	x = A·col + B·row + C
//	y = D·col + E·row + F
//
// Pour une grille « nord en haut » sans rotation : B = D = 0, A = largeur de
// pixel, E = -hauteur de pixel (y décroissant), (C, F) = coin supérieur gauche.

// Affine est une transformation affine 2D (6 coefficients, convention affine/GDAL).
type Affine struct {
	A, B, C, D, E, F float64
}

// Apply transforme des coordonnées pixel (col, row) en coordonnées monde (x, y).
func (t Affine) Apply(col, row float64) (x, y float64) {
	return t.A*col + t.B*row + t.C, t.D*col + t.E*row + t.F
}

// Inverse renvoie la transformation inverse (monde → pixel). Erreur si le
// déterminant est nul (transformation dégénérée).
func (t Affine) Inverse() (Affine, error) {
	det := t.A*t.E - t.B*t.D
	if det == 0 {
		return Affine{}, fmt.Errorf("xarray: géotransformation non inversible (déterminant nul)")
	}
	id := 1 / det
	ia := t.E * id
	ib := -t.B * id
	id2 := -t.D * id
	ie := t.A * id
	return Affine{
		A: ia, B: ib, C: -(ia*t.C + ib*t.F),
		D: id2, E: ie, F: -(id2*t.C + ie*t.F),
	}, nil
}

// FromGDAL construit une Affine à partir d'un GeoTransform GDAL
// [x0, dx, rx, y0, ry, dy] (ordre GDAL : origine, pas, rotations).
func FromGDAL(gt [6]float64) Affine {
	return Affine{A: gt[1], B: gt[2], C: gt[0], D: gt[4], E: gt[5], F: gt[3]}
}

// GDAL renvoie le GeoTransform GDAL équivalent [x0, dx, rx, y0, ry, dy].
func (t Affine) GDAL() [6]float64 {
	return [6]float64{t.C, t.A, t.B, t.F, t.D, t.E}
}

// GeoRef associe une géotransformation et un CRS (identifiant opaque).
type GeoRef struct {
	Transform Affine
	CRS       string // ex. "EPSG:4326", WKT, proj4…
}

// GeoCoords renvoie les coordonnées monde des CENTRES de pixels le long des axes
// x (par colonne) et y (par ligne), pour une grille axis-aligned (B = D = 0).
// Renvoie une erreur si la transformation comporte une rotation (B ou D ≠ 0),
// cas où x et y ne sont pas séparables en axes 1D.
func GeoCoords(t Affine, width, height int) (xs, ys []float64, err error) {
	if t.B != 0 || t.D != 0 {
		return nil, nil, fmt.Errorf("xarray: géotransformation avec rotation non représentable en axes 1D")
	}
	xs = make([]float64, width)
	for i := 0; i < width; i++ {
		xs[i], _ = t.Apply(float64(i)+0.5, 0.5) // centre du pixel
	}
	ys = make([]float64, height)
	for j := 0; j < height; j++ {
		_, ys[j] = t.Apply(0.5, float64(j)+0.5)
	}
	return xs, ys, nil
}

// Georeference renvoie une copie du DataArray 2D avec des coordonnées monde
// (centres de pixels) attachées à xDim et yDim, et le CRS stocké dans l'attribut
// "crs" de la variable. da doit avoir exactement les dimensions [yDim, xDim].
func (da *DataArray[T]) Georeference(gr GeoRef, xDim, yDim string) (*DataArray[T], error) {
	dims := da.variable.Dims()
	if len(dims) != 2 || dims[0] != yDim || dims[1] != xDim {
		return nil, fmt.Errorf("xarray: Georeference attend des dimensions [%q, %q], obtenu %v", yDim, xDim, dims)
	}
	shape := da.variable.Shape()
	height, width := shape[0], shape[1]
	xs, ys, err := GeoCoords(gr.Transform, width, height)
	if err != nil {
		return nil, err
	}
	// Coordonnées converties vers T (les coords monde sont réelles : n'utiliser
	// Georeference que sur un DataArray[float64] pour éviter la troncature).
	coords := map[string][]T{
		xDim: toT[T](xs),
		yDim: toT[T](ys),
	}
	out, err := NewDataArray(da.variable.Dims(), shape, da.variable.data, coords, da.name)
	if err != nil {
		return nil, err
	}
	for k, v := range da.variable.Attrs() {
		out.variable.SetAttr(k, v)
	}
	out.variable.SetAttr("crs", gr.CRS)
	return out, nil
}

func toT[T Number](xs []float64) []T {
	out := make([]T, len(xs))
	for i, x := range xs {
		out[i] = T(x)
	}
	return out
}

// ParseGDALGeoTransform lit une chaîne GeoTransform GDAL « x0 dx rx y0 ry dy »
// (telle qu'écrite par GDAL/rioxarray dans l'attribut GeoTransform) en Affine.
func ParseGDALGeoTransform(s string) (Affine, error) {
	fields := strings.Fields(s)
	if len(fields) != 6 {
		return Affine{}, fmt.Errorf("xarray: GeoTransform attend 6 valeurs, %d trouvées", len(fields))
	}
	var gt [6]float64
	for i, f := range fields {
		v, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return Affine{}, fmt.Errorf("xarray: GeoTransform champ %d invalide: %w", i, err)
		}
		gt[i] = v
	}
	return FromGDAL(gt), nil
}

// GeoRefFromCF extrait le géoréférencement d'une variable de données à partir des
// métadonnées CF (convention rioxarray/GDAL) : l'attribut `grid_mapping` de la
// variable nomme une variable de CRS portant `crs_wkt`/`spatial_ref` (identifiant
// du CRS) et `GeoTransform` (géotransformation GDAL). Renvoie le GeoRef et true
// si une géotransformation exploitable a été trouvée.
func (ds *Dataset[T]) GeoRefFromCF(name string) (GeoRef, bool) {
	da, err := ds.Get(name)
	if err != nil {
		return GeoRef{}, false
	}
	gm := da.variable.Attrs()["grid_mapping"]
	if gm == "" {
		return GeoRef{}, false
	}
	crsVar, err := ds.Get(gm)
	if err != nil {
		return GeoRef{}, false
	}
	at := crsVar.variable.Attrs()
	crs := at["spatial_ref"]
	if crs == "" {
		crs = at["crs_wkt"]
	}
	gt := at["GeoTransform"]
	if gt == "" {
		return GeoRef{CRS: crs}, false
	}
	tr, err := ParseGDALGeoTransform(gt)
	if err != nil {
		return GeoRef{CRS: crs}, false
	}
	return GeoRef{Transform: tr, CRS: crs}, true
}
