// Commande benchexpr : mesure une expression composée multi-tableaux
// mean(a * b) sur deux stores Zarr, lus paresseusement (out-of-core).
//
//	go run ./cmd/benchexpr /tmp/a.zarr /tmp/b.zarr 800
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/benedictemarty/xarray"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: benchexpr <a.zarr> <b.zarr> <chunkSize>")
		os.Exit(1)
	}
	cs, _ := strconv.Atoi(os.Args[3])

	var best time.Duration
	var mean float64
	for i := 0; i < 7; i++ {
		la, err := xarray.ChunkZarr(os.Args[1], cs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERREUR:", err)
			os.Exit(1)
		}
		lb, err := xarray.ChunkZarr(os.Args[2], cs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERREUR:", err)
			os.Exit(1)
		}
		t0 := time.Now()
		prod, err := la.Mul(lb) // graphe différé : a * b
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERREUR:", err)
			os.Exit(1)
		}
		m, err := prod.Mean() // réduction streaming
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
	fmt.Printf("xarray-go mean(a*b) : %.2f ms (résultat=%.4f)\n", float64(best.Microseconds())/1000, mean)
}
