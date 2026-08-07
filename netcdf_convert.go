package xarray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Ouverture des fichiers netCDF « non classiques » (NetCDF-4/HDF5, CDF-2/5) par
// délégation à un convertisseur externe qui les réécrit en CDF-1 classique,
// format que ReadDatasetNetCDF sait lire (avec les attributs et le décodage CF
// du Sprint 60). Aucune dépendance cgo : on shell-out vers `nccopy` ou `cdo`,
// exactement comme le backend eccodes optionnel pour le GRIB.
//
// Ce n'est donc PAS un lecteur HDF5 en Go : c'est un pont pragmatique qui rend
// les fichiers réels lisibles quand un outil de conversion est présent.

// NetCDFFormat identifie le format d'un fichier d'après ses octets de signature.
type NetCDFFormat int

const (
	FormatUnknown NetCDFFormat = iota
	FormatCDF1                 // CDF\x01 — lisible directement
	FormatCDF2                 // CDF\x02 — 64-bit offset
	FormatCDF5                 // CDF\x05 — 64-bit data
	FormatHDF5                 // \x89HDF — NetCDF-4 / HDF5
)

func (f NetCDFFormat) String() string {
	switch f {
	case FormatCDF1:
		return "CDF-1 (classique)"
	case FormatCDF2:
		return "CDF-2 (64-bit offset)"
	case FormatCDF5:
		return "CDF-5 (64-bit data)"
	case FormatHDF5:
		return "NetCDF-4 / HDF5"
	default:
		return "inconnu"
	}
}

// SniffNetCDFFormat lit les premiers octets d'un fichier et déduit son format.
func SniffNetCDFFormat(path string) (NetCDFFormat, error) {
	f, err := os.Open(path)
	if err != nil {
		return FormatUnknown, err
	}
	defer f.Close()
	var magic [8]byte
	n, _ := f.Read(magic[:])
	b := magic[:n]
	switch {
	case len(b) >= 8 && b[0] == 0x89 && b[1] == 'H' && b[2] == 'D' && b[3] == 'F':
		return FormatHDF5, nil
	case len(b) >= 4 && b[0] == 'C' && b[1] == 'D' && b[2] == 'F':
		switch b[3] {
		case 1:
			return FormatCDF1, nil
		case 2:
			return FormatCDF2, nil
		case 5:
			return FormatCDF5, nil
		}
	}
	return FormatUnknown, nil
}

// NetCDFConverter réécrit le fichier src en CDF-1 classique dans dst.
type NetCDFConverter func(src, dst string) error

// FindNetCDFConverter détecte un convertisseur externe disponible dans le PATH
// (`nccopy` fourni par netCDF, ou `cdo`) et renvoie la fonction associée, son
// nom, ou une erreur si aucun n'est trouvé.
func FindNetCDFConverter() (NetCDFConverter, string, error) {
	if p, err := exec.LookPath("nccopy"); err == nil {
		return func(src, dst string) error {
			// -k classic => CDF-1
			return runConv(p, "-k", "classic", src, dst)
		}, "nccopy", nil
	}
	if p, err := exec.LookPath("cdo"); err == nil {
		return func(src, dst string) error {
			// -f nc => netCDF classique (CDF-1)
			return runConv(p, "-f", "nc", "copy", src, dst)
		}, "cdo", nil
	}
	return nil, "", fmt.Errorf("xarray: aucun convertisseur netCDF trouvé (installez `nccopy` ou `cdo`)")
}

func runConv(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("xarray: échec de conversion (%s): %v — %s", filepath.Base(bin), err, out)
	}
	return nil
}

// OpenNetCDFFile ouvre un fichier netCDF quel que soit son format.
//
//   - CDF-1 : lu directement.
//   - HDF5/CDF-2/CDF-5 : converti en CDF-1 via conv (fichier temporaire) puis lu.
//
// Si conv est nil, un convertisseur est recherché dans le PATH. Le décodage CF
// (packing, temps) n'est PAS appliqué ici : appelez DecodeCF/DecodeTime ensuite,
// comme pour ReadDatasetNetCDF.
func OpenNetCDFFile(path string, conv NetCDFConverter) (*Dataset[float64], error) {
	format, err := SniffNetCDFFormat(path)
	if err != nil {
		return nil, err
	}
	src := path
	if format != FormatCDF1 {
		if format == FormatUnknown {
			return nil, fmt.Errorf("xarray: format de %q non reconnu", filepath.Base(path))
		}
		if conv == nil {
			var name string
			conv, name, err = FindNetCDFConverter()
			if err != nil {
				return nil, fmt.Errorf("%s détecté mais non lisible sans conversion : %w", format, err)
			}
			_ = name
		}
		tmp, err := os.CreateTemp("", "xarray-*.nc")
		if err != nil {
			return nil, err
		}
		tmp.Close()
		defer os.Remove(tmp.Name())
		if err := conv(path, tmp.Name()); err != nil {
			return nil, err
		}
		src = tmp.Name()
	}

	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadDatasetNetCDF[float64](f)
}
