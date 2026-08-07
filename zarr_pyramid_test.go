package xarray

import (
	"path/filepath"
	"testing"
)

// TestWritePyramidZarr vérifie l'écriture d'une pyramide multi-échelles :
// niveaux successifs réduits par moyenne de blocs 2×2, avec la métadonnée
// multiscales et des niveaux relisibles.
func TestWritePyramidZarr(t *testing.T) {
	const n = 8
	data := make([]float64, n*n)
	for i := range data {
		data[i] = float64(i)
	}
	ys := make([]float64, n)
	xs := make([]float64, n)
	for i := 0; i < n; i++ {
		ys[i], xs[i] = float64(i), float64(i)
	}
	da, _ := NewDataArray([]string{"y", "x"}, []int{n, n}, data,
		map[string][]float64{"y": ys, "x": xs}, "v")

	dir := filepath.Join(t.TempDir(), "pyr.zarr")
	if err := WritePyramidZarr(dir, da, "y", "x", 3, 2, ZarrZstd); err != nil {
		t.Fatalf("WritePyramidZarr : %v", err)
	}

	// Métadonnée multiscales : 3 niveaux, facteurs 1/2/4.
	levels, err := PyramidLevels(dir)
	if err != nil {
		t.Fatalf("PyramidLevels : %v", err)
	}
	if len(levels) != 3 || levels[0].Factor != 1 || levels[1].Factor != 2 || levels[2].Factor != 4 {
		t.Errorf("niveaux = %+v", levels)
	}

	// Formes : 8×8 → 4×4 → 2×2.
	wantShapes := [][]int{{8, 8}, {4, 4}, {2, 2}}
	for k := 0; k < 3; k++ {
		ds, err := ReadPyramidLevel(dir, k)
		if err != nil {
			t.Fatalf("niveau %d : %v", k, err)
		}
		v, _ := ds.Get("v")
		s := v.Shape()
		if s[0] != wantShapes[k][0] || s[1] != wantShapes[k][1] {
			t.Errorf("niveau %d shape = %v, attendu %v", k, s, wantShapes[k])
		}
	}

	// Coin du niveau 1 = moyenne du bloc 2×2 {0,1,8,9} = 4.5.
	ds1, _ := ReadPyramidLevel(dir, 1)
	v1, _ := ds1.Get("v")
	if v1.Data()[0] != 4.5 {
		t.Errorf("niveau1[0] = %v, attendu 4.5", v1.Data()[0])
	}

	// Erreurs : nlevels/factor invalides, mauvaises dimensions.
	if err := WritePyramidZarr(dir, da, "y", "x", 0, 2, ZarrNone); err == nil {
		t.Error("erreur attendue : nlevels < 1")
	}
	if err := WritePyramidZarr(dir, da, "y", "x", 3, 1, ZarrNone); err == nil {
		t.Error("erreur attendue : factor < 2")
	}
}
