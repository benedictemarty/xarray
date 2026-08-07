package xarray

import (
	"math"
	"os"
	"testing"
)

// valeurs de référence produites par ecCodes pour testdata/complex_synth.grib2
// (grille 8×6, complex packing + différenciation spatiale d'ordre 1, template 5.3).
var complexSynthExpected = []float64{
	0.000000, 1.500000, 3.000000, 4.500000, 6.000000, 7.500000, 9.000000, 10.500000,
	0.703125, 2.500000, 4.296875, 6.101562, 7.898438, 8.203125, 10.000000, 11.796875,
	1.398438, 3.500000, 5.601562, 6.203125, 8.296875, 8.898438, 11.000000, 13.101562,
	2.101562, 4.500000, 5.398438, 7.796875, 8.703125, 9.601562, 12.000000, 12.898438,
	2.796875, 5.500000, 6.703125, 7.898438, 9.101562, 10.296875, 13.000000, 14.203125,
	3.500000, 5.000000, 6.500000, 8.000000, 9.500000, 11.000000, 12.500000, 14.000000,
}

// TestGribComplexPacking décode un fichier GRIB2 en complex packing avec
// différenciation spatiale (template 5.3) et compare aux valeurs d'ecCodes.
// Le fichier de référence est un champ synthétique (données inventées), généré
// par ecCodes, versionné dans testdata/.
func TestGribComplexPacking(t *testing.T) {
	f, err := os.Open("testdata/complex_synth.grib2")
	if err != nil {
		t.Skipf("fichier de référence absent : %v", err)
	}
	defer f.Close()

	msgs, err := ReadGrib(f)
	if err != nil {
		t.Fatalf("ReadGrib : %v", err)
	}
	m := msgs[0]
	if m.Ni != 8 || m.Nj != 6 {
		t.Fatalf("Ni/Nj = %d/%d", m.Ni, m.Nj)
	}
	if len(m.Values) != len(complexSynthExpected) {
		t.Fatalf("%d valeurs, attendu %d", len(m.Values), len(complexSynthExpected))
	}
	for i, exp := range complexSynthExpected {
		if math.Abs(m.Values[i]-exp) > 1e-5 {
			t.Errorf("valeur[%d] = %v, attendu %v (ecCodes)", i, m.Values[i], exp)
		}
	}
}
