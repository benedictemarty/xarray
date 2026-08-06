// Commande genzarrds : écrit un Dataset comme groupe Zarr v2 (démo interop).
//
//	go run ./cmd/genzarrds /tmp/grp.zarr
package main

import (
	"fmt"
	"os"

	"github.com/bmarty/xarray"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: genzarrds <dir>")
		os.Exit(1)
	}
	dir := os.Args[1]

	temp, _ := xarray.NewDataArray([]string{"temps", "lieu"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"temps": {2020, 2021}, "lieu": {10, 20, 30}}, "temperature")
	pluie, _ := xarray.NewDataArray([]string{"temps"}, []int{2}, []float64{100, 200},
		map[string][]float64{"temps": {2020, 2021}}, "pluie")
	ds, _ := xarray.NewDataset(map[string]*xarray.DataArray[float64]{
		"temperature": temp, "pluie": pluie,
	})

	if err := xarray.WriteDatasetZarr(dir, ds, xarray.ZarrZlib); err != nil {
		panic(err)
	}
	fmt.Println("groupe Zarr v2 écrit :", dir)
}
