// Commande readgrib : lit un GRIB2 (regular_ll, simple packing) et affiche la
// grille et quelques valeurs (pour valider contre ecCodes).
//
//	go run ./cmd/readgrib /tmp/test_simple.grib2
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/bmarty/xarray"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: readgrib <fichier.grib2>")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	msgs, err := xarray.ReadGrib(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERREUR:", err)
		os.Exit(1)
	}
	for k, m := range msgs {
		da, _ := m.ToDataArray("champ")
		vals := da.Data()
		mn, mx := vals[0], vals[0]
		for _, v := range vals {
			if v < mn {
				mn = v
			}
			if v > mx {
				mx = v
			}
		}
		fmt.Printf("message %d : Ni=%d Nj=%d n=%d La1=%.3f Lo1=%.3f Di=%.3f Dj=%.3f\n",
			k, m.Ni, m.Nj, len(vals), m.La1, m.Lo1, m.Di, m.Dj)
		fmt.Printf("  min=%.4f max=%.4f vals[:5]=%.4f\n", mn, mx, vals[:5])

		// Mode dump : écrit les valeurs (float64 little-endian) pour comparaison.
		if len(os.Args) > 2 && k == 0 {
			buf := make([]byte, len(vals)*8)
			for i, v := range vals {
				binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
			}
			if err := os.WriteFile(os.Args[2], buf, 0o644); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}
}
