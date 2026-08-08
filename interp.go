package xarray

import "fmt"

// Interpolation bilinéaire d'un DataArray sur ses axes x/y, à une position monde
// (x, y). Réduit les deux dimensions x/y en combinant les 4 pixels encadrants
// pondérés ; les dimensions restantes (temps, niveau…) sont conservées. Utile
// pour échantillonner une valeur au point exact plutôt qu'au plus proche voisin.

// bracketF trouve, dans un axe de coordonnées monotone (croissant ou
// décroissant), les indices i0, i1 encadrant v et le poids w tel que
// v ≈ coord[i0]*(1-w) + coord[i1]*w. ok=false si v est hors de l'axe.
func bracketF(coords []float64, v float64) (i0, i1 int, w float64, ok bool) {
	n := len(coords)
	if n == 0 {
		return 0, 0, 0, false
	}
	if n == 1 {
		return 0, 0, 0, coords[0] == v
	}
	for i := 0; i < n-1; i++ {
		a, b := coords[i], coords[i+1]
		lo, hi := a, b
		if lo > hi {
			lo, hi = hi, lo
		}
		if v >= lo && v <= hi {
			if b == a {
				return i, i + 1, 0, true
			}
			return i, i + 1, (v - a) / (b - a), true
		}
	}
	return 0, 0, 0, false
}

// InterpBilinear renvoie le DataArray réduit sur xDim/yDim par interpolation
// bilinéaire à la position (x, y). Erreur si le point est hors de la grille.
func InterpBilinear(da *DataArray[float64], xDim, yDim string, x, y float64) (*DataArray[float64], error) {
	xs, err := da.Coord(xDim)
	if err != nil {
		return nil, err
	}
	ys, err := da.Coord(yDim)
	if err != nil {
		return nil, err
	}
	ix0, ix1, wx, okx := bracketF(xs, x)
	iy0, iy1, wy, oky := bracketF(ys, y)
	if !okx || !oky {
		return nil, fmt.Errorf("xarray: point (%.6g, %.6g) hors de la grille pour l'interpolation", x, y)
	}
	corner := func(ix, iy int) (*DataArray[float64], error) {
		s, e := da.Isel(xDim, ix)
		if e != nil {
			return nil, e
		}
		return s.Isel(yDim, iy)
	}
	c00, err := corner(ix0, iy0)
	if err != nil {
		return nil, err
	}
	c10, err := corner(ix1, iy0)
	if err != nil {
		return nil, err
	}
	c01, err := corner(ix0, iy1)
	if err != nil {
		return nil, err
	}
	c11, err := corner(ix1, iy1)
	if err != nil {
		return nil, err
	}
	res := c00.MulScalar((1 - wx) * (1 - wy))
	if res, err = res.Add(c10.MulScalar(wx * (1 - wy))); err != nil {
		return nil, err
	}
	if res, err = res.Add(c01.MulScalar((1 - wx) * wy)); err != nil {
		return nil, err
	}
	if res, err = res.Add(c11.MulScalar(wx * wy)); err != nil {
		return nil, err
	}
	return res, nil
}
