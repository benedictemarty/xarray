// Démo runnable : reprojection d'un raster géoréférencé (WGS84 lon/lat →
// Web Mercator EPSG:3857) avec interpolation bilinéaire, de bout en bout.
//
//	go run ./cmd/reproject
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/benedictemarty/xarray"
)

func main() {
	// Scène source en EPSG:4326 : 8×12, un « disque » de valeurs élevées au centre.
	const sh, sw = 8, 12
	src := make([]float64, sh*sw)
	for r := 0; r < sh; r++ {
		for c := 0; c < sw; c++ {
			dr, dc := float64(r)-3.5, float64(c)-5.5
			if dr*dr+dc*dc/2 < 6 {
				src[r*sw+c] = 1 // motif
			}
		}
	}
	// Affine source : origine (0°E, 55°N), pixel 1°, y décroissant.
	srcT := xarray.Affine{A: 1, C: 0, E: -1, F: 55}

	fmt.Println("=== Source (EPSG:4326, pixel 1°) ===")
	printGrid(src, sh, sw)

	// Grille cible EPSG:3857 couvrant l'emprise source, ~même nombre de pixels.
	x0, y0, _ := xarray.TransformXY("EPSG:4326", "EPSG:3857", 0, 55) // coin haut-gauche
	x1, y1, _ := xarray.TransformXY("EPSG:4326", "EPSG:3857", float64(sw), 55-float64(sh))
	dw, dh := sw, sh
	dstT := xarray.Affine{A: (x1 - x0) / float64(dw), C: x0, E: (y1 - y0) / float64(dh), F: y0}

	for _, m := range []struct {
		name   string
		method xarray.Resampling
	}{{"plus proche voisin", xarray.Nearest}, {"bilinéaire", xarray.Bilinear}} {
		out, err := xarray.Reproject(src, sw, sh, srcT, "EPSG:4326", dstT, dw, dh, "EPSG:3857", m.method)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("\n=== Reprojeté en EPSG:3857 (%s) ===\n", m.name)
		printGrid(out, dh, dw)
	}

	// Export du résultat bilinéaire en Zarr.
	out, _ := xarray.Reproject(src, sw, sh, srcT, "EPSG:4326", dstT, dw, dh, "EPSG:3857", xarray.Bilinear)
	da, _ := xarray.NewDataArray([]string{"y", "x"}, []int{dh, dw}, out, nil, "band")
	geo := xarray.GeoRef{Transform: dstT, CRS: "EPSG:3857"}
	g, _ := da.Georeference(geo, "x", "y")
	ds, _ := xarray.NewDataset(map[string]*xarray.DataArray[float64]{"band": g})
	dir := "/tmp/reprojected.zarr"
	os.RemoveAll(dir)
	if err := xarray.WriteDatasetZarr(dir, ds, xarray.ZarrZstd); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	xs, _ := g.Coord("x")
	fmt.Printf("\nExporté (EPSG:3857) : %s\n", dir)
	fmt.Printf("axe x (mètres) : %.0f … %.0f\n", xs[0], xs[len(xs)-1])
}

func printGrid(d []float64, h, w int) {
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			v := d[r*w+c]
			switch {
			case math.IsNaN(v):
				fmt.Print(" ·")
			case v <= 0.05:
				fmt.Print(" .")
			case v < 0.5:
				fmt.Print(" +")
			default:
				fmt.Print(" #")
			}
		}
		fmt.Println()
	}
}
