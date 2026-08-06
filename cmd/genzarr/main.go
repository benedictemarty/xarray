// Commande genzarr : écrit un store Zarr v2 de démonstration, pour vérifier
// l'interopérabilité avec zarr-python.
//
//	go run ./cmd/genzarr /tmp/interop.zarr
package main

import (
	"fmt"
	"os"

	"github.com/bmarty/xarray"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: genzarr <dir>")
		os.Exit(1)
	}
	dir := os.Args[1]

	data := make([]float64, 20)
	for i := range data {
		data[i] = float64(i)
	}
	da, err := xarray.NewDataArray(
		[]string{"x", "y"}, []int{5, 4}, data,
		map[string][]float64{"x": {0, 1, 2, 3, 4}, "y": {10, 20, 30, 40}},
		"demo",
	)
	if err != nil {
		panic(err)
	}
	// Chunks 2×3 (non alignés) + compression zlib : cas le plus exigeant.
	if err := xarray.WriteDataArrayZarr(dir, da, []int{2, 3}, xarray.ZarrZlib); err != nil {
		panic(err)
	}
	fmt.Println("store Zarr v2 écrit :", dir)
}
