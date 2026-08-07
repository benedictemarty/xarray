// Commande readgrib-ec : lit un GRIB via le backend ecCodes (tout template).
//
//	go run -tags eccodes ./cmd/readgrib-ec <fichier.grib> [dump.bin]
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/benedictemarty/xarray/experimental/eccodesgrib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: readgrib-ec <fichier.grib> [dump.bin]")
		os.Exit(1)
	}
	fields, err := eccodesgrib.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERREUR:", err)
		os.Exit(1)
	}
	for k, da := range fields {
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
		fmt.Printf("message %d : %s shape=%v n=%d min=%.4f max=%.4f vals[:5]=%.4f\n",
			k, da.Name(), da.Shape(), len(vals), mn, mx, vals[:5])
		if len(os.Args) > 2 && k == 0 {
			buf := make([]byte, len(vals)*8)
			for i, v := range vals {
				binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
			}
			os.WriteFile(os.Args[2], buf, 0o644)
		}
	}
}
