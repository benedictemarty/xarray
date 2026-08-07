// Démo runnable : calcul d'un NDVI géoréférencé de bout en bout avec xarray-go.
//
//	NDVI = (NIR - ROUGE) / (NIR + ROUGE)
//
// Enchaîne : construction d'une scène 2 bandes → arithmétique DataArray →
// géoréférencement (affine + CRS) → découpe par emprise (coordonnées monde) →
// export Zarr. Dans un vrai flux, les bandes viendraient de ReadDatasetZarr /
// OpenNetCDFFile + DecodeCF (voir docs/SATELLITE.md).
//
//	go run ./cmd/ndvi
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/benedictemarty/xarray"
	"github.com/benedictemarty/xarray/geoapi"
)

func main() {
	const ny, nx = 6, 8
	// Scène synthétique : une bande de "végétation" (NIR fort, ROUGE faible) au
	// centre, de l'eau/sol nu ailleurs (NIR faible).
	red := make([]float64, ny*nx)
	nir := make([]float64, ny*nx)
	for y := 0; y < ny; y++ {
		for x := 0; x < nx; x++ {
			i := y*nx + x
			veg := y >= 2 && y <= 3 && x >= 2 && x <= 5 // parcelle végétale
			if veg {
				red[i], nir[i] = 0.08, 0.55 // NDVI ≈ 0.75
			} else if x < 2 {
				red[i], nir[i] = 0.04, 0.03 // eau : NDVI < 0
			} else {
				red[i], nir[i] = 0.20, 0.28 // sol nu : NDVI ≈ 0.17
			}
		}
	}
	redDA, _ := xarray.NewDataArray([]string{"y", "x"}, []int{ny, nx}, red, nil, "red")
	nirDA, _ := xarray.NewDataArray([]string{"y", "x"}, []int{ny, nx}, nir, nil, "nir")

	// NDVI = (NIR - ROUGE) / (NIR + ROUGE) via l'arithmétique DataArray.
	num, err := nirDA.Sub(redDA)
	must(err)
	den, err := nirDA.Add(redDA)
	must(err)
	ndvi, err := num.Div(den)
	must(err)
	ndvi = ndvi.Rename("ndvi")

	// Géoréférencement : origine (2°E, 51°N), pixel 0,5°, y décroissant.
	gr := xarray.GeoRef{
		Transform: xarray.Affine{A: 0.5, C: 2, E: -0.5, F: 51},
		CRS:       "EPSG:4326",
	}
	geo, err := ndvi.Georeference(gr, "x", "y")
	must(err)

	xs, _ := geo.Coord("x")
	ys, _ := geo.Coord("y")
	fmt.Println("=== NDVI géoréférencé (", ny, "×", nx, ", CRS", gr.CRS, ") ===")
	fmt.Printf("longitudes: %.2f\nlatitudes : %.2f\n\n", xs, ys)
	printNDVI(geo.Data(), ny, nx)
	fmt.Printf("\nstats NDVI : min=%.3f  max=%.3f  moyenne=%.3f\n",
		geo.Min(), geo.Max(), geo.Mean())

	// Découpe par emprise (coordonnées monde) autour de la parcelle végétale.
	sub, err := geoapi.SubsetBBox(geo, "x", "y", geoapi.BBox{MinX: 3, MinY: 49, MaxX: 5, MaxY: 50})
	must(err)
	sxs, _ := sub.Coord("x")
	sys, _ := sub.Coord("y")
	fmt.Printf("\n=== emprise [3..5°E, 49..50°N] -> %v ===\n", sub.Shape())
	fmt.Printf("lon=%.2f lat=%.2f  NDVI moyen=%.3f\n", sxs, sys, sub.Mean())

	// Export Zarr (groupe, zstd) — relisible par zarr-python / xarray.
	out := "/tmp/ndvi.zarr"
	os.RemoveAll(out)
	outDS, err := xarray.NewDataset(map[string]*xarray.DataArray[float64]{"ndvi": geo})
	must(err)
	must(xarray.WriteDatasetZarr(out, outDS, xarray.ZarrZstd))
	fmt.Printf("\nNDVI exporté en Zarr (groupe) : %s\n", out)
	fmt.Println("Vérif Python : python -c \"import xarray as xr; print(xr.open_zarr('" + out + "')['ndvi'].values)\"")
}

// printNDVI affiche une carte ASCII (- eau, . sol, o/O/# végétation croissante).
func printNDVI(d []float64, ny, nx int) {
	fmt.Println("carte NDVI (# = végétation dense) :")
	for y := 0; y < ny; y++ {
		for x := 0; x < nx; x++ {
			v := d[y*nx+x]
			var c byte
			switch {
			case math.IsNaN(v):
				c = ' '
			case v < 0:
				c = '~' // eau
			case v < 0.2:
				c = '.' // sol nu
			case v < 0.5:
				c = 'o'
			default:
				c = '#' // végétation dense
			}
			fmt.Printf(" %c", c)
		}
		fmt.Println()
	}
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "erreur:", err)
		os.Exit(1)
	}
}
