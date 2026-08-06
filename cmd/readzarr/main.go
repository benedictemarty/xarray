// Commande readzarr : lit un store Zarr v2 et affiche forme et données.
//
//	go run ./cmd/readzarr /tmp/py.zarr
package main

import (
	"fmt"
	"os"

	"github.com/bmarty/xarray"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: readzarr <dir>")
		os.Exit(1)
	}
	da, err := xarray.ReadDataArrayZarr(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERREUR:", err)
		os.Exit(1)
	}
	fmt.Println("dims", da.Dims(), "shape", da.Shape())
	fmt.Println("data", da.Data())
}
