// Commande benchzarr : mesure le temps de calcul de la moyenne d'un tableau Zarr
// lu paresseusement (out-of-core) par le moteur lazy.
//
//	go run ./cmd/benchzarr /tmp/lazy.zarr 400
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bmarty/xarray"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: benchzarr <dir.zarr> <chunkSize>")
		os.Exit(1)
	}
	chunkSize, _ := strconv.Atoi(os.Args[2])

	// Warmup + mesure sur plusieurs itérations.
	var best time.Duration
	var mean float64
	for i := 0; i < 7; i++ {
		lz, err := xarray.ChunkZarr(os.Args[1], chunkSize)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERREUR:", err)
			os.Exit(1)
		}
		t0 := time.Now()
		m, err := lz.Mean()
		d := time.Since(t0)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERREUR:", err)
			os.Exit(1)
		}
		mean = m
		if i == 0 || d < best {
			best = d
		}
	}
	fmt.Printf("xarray-go ChunkZarr.Mean : %.2f ms (moyenne=%.4f)\n", float64(best.Microseconds())/1000, mean)
}
