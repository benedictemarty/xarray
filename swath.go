package xarray

import (
	"fmt"
	"math"
)

// Rééchantillonnage de données en fauchée (swath) de satellite défilant vers une
// grille régulière. Une fauchée n'a pas de grille régulière : chaque pixel est
// géolocalisé par ses propres (lon, lat) (tableaux 2D de même forme que les
// données). On projette les pixels sources dans les cellules d'une grille cible
// lon/lat par affectation au plus proche (binning « drop-in-the-bucket ») ;
// plusieurs pixels dans une cellule → moyenne ; cellules vides → NaN.
//
// C'est l'équivalent minimal de pyresample (plus proche voisin) pour rendre une
// fauchée (Sentinel-3, MetOp, VIIRS…) subsettable comme une grille.

// ResampleSwathNearest projette les valeurs data, géolocalisées par lon/lat (de
// même longueur), sur la grille cible définie par la géotransformation dstT
// (degrés, lon/lat) et dstW×dstH. Renvoie les valeurs grillées (ordre C) et le
// nombre de pixels sources tombés dans chaque cellule.
func ResampleSwathNearest(data, lon, lat []float64, dstT Affine, dstW, dstH int) (grid []float64, counts []int, err error) {
	if len(data) != len(lon) || len(lon) != len(lat) {
		return nil, nil, fmt.Errorf("xarray: data/lon/lat de longueurs différentes (%d/%d/%d)", len(data), len(lon), len(lat))
	}
	if dstW <= 0 || dstH <= 0 {
		return nil, nil, fmt.Errorf("xarray: dimensions cibles invalides")
	}
	inv, err := dstT.Inverse()
	if err != nil {
		return nil, nil, err
	}
	sum := make([]float64, dstW*dstH)
	counts = make([]int, dstW*dstH)
	for k := range data {
		v := data[k]
		if math.IsNaN(v) {
			continue
		}
		col, row := inv.Apply(lon[k], lat[k])
		i, j := int(math.Floor(col)), int(math.Floor(row))
		if i < 0 || i >= dstW || j < 0 || j >= dstH {
			continue
		}
		sum[j*dstW+i] += v
		counts[j*dstW+i]++
	}
	grid = make([]float64, dstW*dstH)
	for idx := range grid {
		if counts[idx] == 0 {
			grid[idx] = math.NaN()
		} else {
			grid[idx] = sum[idx] / float64(counts[idx])
		}
	}
	return grid, counts, nil
}

// ResampleSwathNearestRadius rééchantillonne une fauchée par plus proche voisin
// avec un rayon de recherche : pour chaque cellule cible, on retient la valeur du
// pixel source le plus proche (distance en lon/lat) trouvé dans une fenêtre de
// ±radiusCells cellules ; aucune → NaN. Contrairement au binning simple, cela
// COMBLE les cellules vides quand la grille cible est plus fine que la fauchée
// (équivalent au radius_of_influence de pyresample, exprimé ici en cellules).
func ResampleSwathNearestRadius(data, lon, lat []float64, dstT Affine, dstW, dstH, radiusCells int) ([]float64, error) {
	if len(data) != len(lon) || len(lon) != len(lat) {
		return nil, fmt.Errorf("xarray: data/lon/lat de longueurs différentes")
	}
	if dstW <= 0 || dstH <= 0 || radiusCells < 0 {
		return nil, fmt.Errorf("xarray: paramètres cibles invalides")
	}
	inv, err := dstT.Inverse()
	if err != nil {
		return nil, err
	}
	// Répartit les pixels sources dans les cellules (index par cellule).
	buckets := make([][]int, dstW*dstH)
	for k := range data {
		if math.IsNaN(data[k]) {
			continue
		}
		col, row := inv.Apply(lon[k], lat[k])
		i, j := int(math.Floor(col)), int(math.Floor(row))
		if i >= 0 && i < dstW && j >= 0 && j < dstH {
			buckets[j*dstW+i] = append(buckets[j*dstW+i], k)
		}
	}
	out := make([]float64, dstW*dstH)
	for j := 0; j < dstH; j++ {
		for i := 0; i < dstW; i++ {
			cx, cy := dstT.Apply(float64(i)+0.5, float64(j)+0.5) // centre de la cellule (monde)
			best, bestD := -1, math.Inf(1)
			for dj := -radiusCells; dj <= radiusCells; dj++ {
				jj := j + dj
				if jj < 0 || jj >= dstH {
					continue
				}
				for di := -radiusCells; di <= radiusCells; di++ {
					ii := i + di
					if ii < 0 || ii >= dstW {
						continue
					}
					for _, k := range buckets[jj*dstW+ii] {
						d := math.Hypot(lon[k]-cx, lat[k]-cy)
						if d < bestD {
							bestD, best = d, k
						}
					}
				}
			}
			if best >= 0 {
				out[j*dstW+i] = data[best]
			} else {
				out[j*dstW+i] = math.NaN()
			}
		}
	}
	return out, nil
}

// SwathToDataArray rééchantillonne une fauchée en un DataArray[float64]
// géoréférencé sur la grille cible (dimensions [yDim, xDim]), directement
// subsettable (SubsetBBox, Query…). name est le nom de la variable produite.
func SwathToDataArray(data, lon, lat []float64, dstT Affine, dstW, dstH int, name, yDim, xDim string) (*DataArray[float64], error) {
	grid, _, err := ResampleSwathNearest(data, lon, lat, dstT, dstW, dstH)
	if err != nil {
		return nil, err
	}
	xs, ys, err := GeoCoords(dstT, dstW, dstH)
	if err != nil {
		return nil, err
	}
	da, err := NewDataArray([]string{yDim, xDim}, []int{dstH, dstW}, grid,
		map[string][]float64{xDim: xs, yDim: ys}, name)
	if err != nil {
		return nil, err
	}
	da.variable.SetAttr("crs", "EPSG:4326")
	return da, nil
}
