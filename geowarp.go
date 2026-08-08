package xarray

import (
	"fmt"
	"math"
)

// Rééchantillonnage/reprojection de raster par plus proche voisin. Pour chaque
// pixel de la grille cible : coordonnées monde (affine cible) → transformation
// vers le CRS source (TransformXY) → pixel source (affine source inverse) →
// échantillon du plus proche voisin. Les pixels hors emprise source valent NaN.
//
// La transformation de coordonnées est limitée aux paires gérées par TransformXY
// (4326 ↔ 3857 et l'identité) ; les autres CRS nécessitent PROJ.

// Resampling désigne la méthode d'interpolation au rééchantillonnage.
type Resampling int

const (
	Nearest  Resampling = iota // plus proche voisin
	Bilinear                   // interpolation bilinéaire (données continues)
)

// sampleNearest échantillonne src au plus proche voisin d'un pixel fractionnaire
// (col, row) exprimé en indices de pixels (le centre du pixel i est à i+0.5).
func sampleNearest(src []float64, w, h int, col, row float64) float64 {
	sc, sr := int(math.Floor(col)), int(math.Floor(row))
	if sc < 0 || sc >= w || sr < 0 || sr >= h {
		return math.NaN()
	}
	return src[sr*w+sc]
}

// sampleBilinear interpole bilinéairement autour de (col, row). Renvoie NaN si
// l'un des 4 voisins est hors grille ou vaut NaN.
func sampleBilinear(src []float64, w, h int, col, row float64) float64 {
	fc, fr := col-0.5, row-0.5 // coordonnées centrées sur les centres de pixels
	i0, j0 := int(math.Floor(fc)), int(math.Floor(fr))
	if i0 < 0 || i0+1 >= w || j0 < 0 || j0+1 >= h {
		return math.NaN()
	}
	wx, wy := fc-float64(i0), fr-float64(j0)
	v00, v01 := src[j0*w+i0], src[j0*w+i0+1]
	v10, v11 := src[(j0+1)*w+i0], src[(j0+1)*w+i0+1]
	top := v00*(1-wx) + v01*wx
	bot := v10*(1-wx) + v11*wx
	return top*(1-wy) + bot*wy
}

// ReprojectNearest reprojette une grille source (srcW×srcH, géotransformation
// srcT, CRS srcCRS) vers une grille cible (dstW×dstH, géotransformation dstT,
// CRS dstCRS) par plus proche voisin. src et le résultat sont en ordre C (ligne
// par ligne), indexés [row*width + col].
func ReprojectNearest(src []float64, srcW, srcH int, srcT Affine, srcCRS string,
	dstT Affine, dstW, dstH int, dstCRS string) ([]float64, error) {
	return Reproject(src, srcW, srcH, srcT, srcCRS, dstT, dstW, dstH, dstCRS, Nearest)
}

// Reproject reprojette une grille source vers une grille cible selon la méthode
// de rééchantillonnage choisie (Nearest ou Bilinear).
func Reproject(src []float64, srcW, srcH int, srcT Affine, srcCRS string,
	dstT Affine, dstW, dstH int, dstCRS string, method Resampling) ([]float64, error) {
	if len(src) != srcW*srcH {
		return nil, fmt.Errorf("xarray: taille source %d ≠ %d×%d", len(src), srcW, srcH)
	}
	if dstW <= 0 || dstH <= 0 {
		return nil, fmt.Errorf("xarray: dimensions cibles invalides")
	}
	srcInv, err := srcT.Inverse()
	if err != nil {
		return nil, err
	}
	// Vérifie tôt que la paire de CRS est gérée (évite N² erreurs).
	if _, _, err := TransformXY(dstCRS, srcCRS, dstT.C, dstT.F); err != nil {
		return nil, err
	}
	out := make([]float64, dstW*dstH)
	for j := 0; j < dstH; j++ {
		for i := 0; i < dstW; i++ {
			tx, ty := dstT.Apply(float64(i)+0.5, float64(j)+0.5) // centre du pixel cible (monde cible)
			sx, sy, err := TransformXY(dstCRS, srcCRS, tx, ty)   // vers le CRS source
			if err != nil {
				return nil, err
			}
			col, row := srcInv.Apply(sx, sy) // pixel source (fractionnaire)
			if method == Bilinear {
				out[j*dstW+i] = sampleBilinear(src, srcW, srcH, col, row)
			} else {
				out[j*dstW+i] = sampleNearest(src, srcW, srcH, col, row)
			}
		}
	}
	return out, nil
}

// ReprojectDataArray reprojette un DataArray 2D géoréférencé (dimensions
// [yDim, xDim], géotransformation srcT, CRS srcCRS) vers une grille cible définie
// par dstT, dstW×dstH et dstCRS. Le résultat porte les coordonnées monde de la
// grille cible et le CRS cible (attribut "crs").
func ReprojectDataArray(da *DataArray[float64], srcT Affine, srcCRS string,
	dstT Affine, dstW, dstH int, dstCRS string, yDim, xDim string, method Resampling) (*DataArray[float64], error) {
	dims := da.variable.Dims()
	if len(dims) != 2 || dims[0] != yDim || dims[1] != xDim {
		return nil, fmt.Errorf("xarray: Reproject attend des dimensions [%q, %q], obtenu %v", yDim, xDim, dims)
	}
	shape := da.variable.Shape()
	out, err := Reproject(da.variable.data, shape[1], shape[0], srcT, srcCRS, dstT, dstW, dstH, dstCRS, method)
	if err != nil {
		return nil, err
	}
	xs, ys, err := GeoCoords(dstT, dstW, dstH)
	if err != nil {
		return nil, err
	}
	res, err := NewDataArray([]string{yDim, xDim}, []int{dstH, dstW}, out,
		map[string][]float64{xDim: xs, yDim: ys}, da.name)
	if err != nil {
		return nil, err
	}
	res.variable.SetAttr("crs", dstCRS)
	return res, nil
}
