package xarray

import (
	"os"
	"os/exec"
	"testing"
)

func TestSniffNetCDFFormat(t *testing.T) {
	cases := map[string]NetCDFFormat{
		"/tmp/ncprobe/A_cdf1_noattr.nc": FormatCDF1,
		"/tmp/ncprobe/C_netcdf4.nc":     FormatHDF5,
		"/tmp/ncprobe/D_cdf2.nc":        FormatCDF2,
	}
	for path, want := range cases {
		if _, err := os.Stat(path); err != nil {
			t.Skipf("fichier de sonde absent : %s", path)
		}
		got, err := SniffNetCDFFormat(path)
		if err != nil {
			t.Fatalf("%s : %v", path, err)
		}
		if got != want {
			t.Errorf("%s : format=%v, attendu %v", path, got, want)
		}
	}
}

// Sans convertisseur injecté ni outil dans le PATH, un HDF5 doit échouer par une
// erreur explicite (jamais un panic, jamais une lecture erronée).
func TestOpenHDF5SansConvertisseur(t *testing.T) {
	path := "/tmp/ncprobe/C_netcdf4.nc"
	if _, err := os.Stat(path); err != nil {
		t.Skip("fichier de sonde absent")
	}
	// Convertisseur qui échoue toujours : simule l'absence d'outil.
	failConv := func(src, dst string) error { return os.ErrNotExist }
	if _, err := OpenNetCDFFile(path, failConv); err == nil {
		t.Fatal("erreur attendue quand la conversion échoue")
	}
}

// Validation de bout en bout : convertisseur stand-in (xarray Python, disponible
// dans cet environnement) HDF5 -> CDF-1, puis lecture par xarray-go. Prouve que
// le pipeline OpenNetCDFFile est correct ; en production le convertisseur serait
// nccopy/cdo.
func TestOpenHDF5ViaPythonStandin(t *testing.T) {
	path := "/tmp/ncprobe/C_netcdf4.nc"
	if _, err := os.Stat(path); err != nil {
		t.Skip("fichier de sonde absent")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 absent")
	}
	pyConv := func(src, dst string) error {
		script := "import sys,xarray as xr; " +
			"xr.open_dataset(sys.argv[1]).to_netcdf(sys.argv[2], format='NETCDF3_CLASSIC')"
		return exec.Command("python3", "-c", script, src, dst).Run()
	}
	ds, err := OpenNetCDFFile(path, pyConv)
	if err != nil {
		t.Fatalf("OpenNetCDFFile via stand-in : %v", err)
	}
	v, err := ds.Get("t2m")
	if err != nil {
		t.Fatalf("variable t2m absente : %v", err)
	}
	d := v.Data()
	// Le fichier C a été généré avec data = 0..11 (grille 3x4).
	if len(d) != 12 || d[0] != 0 || d[11] != 11 {
		t.Errorf("données lues = %v, attendu 0..11", d)
	}
}
