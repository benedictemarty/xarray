//go:build eccodes

// Package eccodesgrib est un backend GRIB s'appuyant sur la bibliothèque C
// ecCodes (ECMWF) via cgo. Contrairement au décodeur pur-Go du paquet racine
// (limité aux templates documentés : simple et complex packing standard), ce
// backend délègue TOUT le décodage à ecCodes : il gère donc n'importe quel
// template, y compris les **templates locaux** non documentés publiquement
// (ex. 50002 de Météo-France), GRIB1, toutes les grilles, bitmaps, etc.
//
// Il n'y a AUCUN cas particulier à coder : ecCodes connaît les définitions de
// chaque centre. C'est le compromis « dépendance externe + cgo » contre
// « couverture complète ».
//
// Compilation (opt-in) : `go build -tags eccodes`. Nécessite libeccodes et ses
// en-têtes ; ajuster les chemins cgo ci-dessous selon l'installation.
package eccodesgrib

/*
#cgo CFLAGS: -I/home/bmarty/.local/lib/python3.13/site-packages/eccodeslib/include
#cgo LDFLAGS: -L/home/bmarty/.local/lib/python3.13/site-packages/eccodeslib/lib64 -L/home/bmarty/.local/lib/python3.13/site-packages/eckitlib/lib64 -leccodes -lstdc++ -Wl,-rpath,/home/bmarty/.local/lib/python3.13/site-packages/eccodeslib/lib64 -Wl,-rpath,/home/bmarty/.local/lib/python3.13/site-packages/eckitlib/lib64
#include <eccodes.h>
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/benedictemarty/xarray"
)

func getLong(h *C.codes_handle, key string) (int64, error) {
	ck := C.CString(key)
	defer C.free(unsafe.Pointer(ck))
	var v C.long
	if C.codes_get_long(h, ck, &v) != 0 {
		return 0, fmt.Errorf("eccodes: clé %q absente", key)
	}
	return int64(v), nil
}

func getDouble(h *C.codes_handle, key string) (float64, error) {
	ck := C.CString(key)
	defer C.free(unsafe.Pointer(ck))
	var v C.double
	if C.codes_get_double(h, ck, &v) != 0 {
		return 0, fmt.Errorf("eccodes: clé %q absente", key)
	}
	return float64(v), nil
}

func getDoubleArray(h *C.codes_handle, key string) ([]float64, error) {
	ck := C.CString(key)
	defer C.free(unsafe.Pointer(ck))
	var size C.size_t
	if C.codes_get_size(h, ck, &size) != 0 {
		return nil, fmt.Errorf("eccodes: taille de %q inconnue", key)
	}
	buf := make([]float64, int(size))
	if size > 0 {
		if C.codes_get_double_array(h, ck, (*C.double)(unsafe.Pointer(&buf[0])), &size) != 0 {
			return nil, fmt.Errorf("eccodes: lecture de %q impossible", key)
		}
	}
	return buf, nil
}

func getString(h *C.codes_handle, key string) string {
	ck := C.CString(key)
	defer C.free(unsafe.Pointer(ck))
	var length C.size_t = 256
	buf := make([]C.char, 256)
	if C.codes_get_string(h, ck, &buf[0], &length) != 0 {
		return ""
	}
	return C.GoString(&buf[0])
}

// ReadFile lit tous les messages GRIB d'un fichier via ecCodes et les convertit
// en DataArray[float64] (dimensions latitude, longitude). Fonctionne pour tout
// template pris en charge par ecCodes, templates locaux compris.
func ReadFile(path string) ([]*xarray.DataArray[float64], error) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	cmode := C.CString("rb")
	defer C.free(unsafe.Pointer(cmode))

	f, errno := C.fopen(cpath, cmode)
	if f == nil {
		return nil, fmt.Errorf("eccodes: ouverture de %q impossible: %v", path, errno)
	}
	defer C.fclose(f)

	var out []*xarray.DataArray[float64]
	for {
		var cerr C.int
		h := C.codes_grib_handle_new_from_file(nil, f, &cerr)
		if h == nil {
			break // fin de fichier
		}

		ni, err := getLong(h, "Ni")
		if err != nil {
			C.codes_handle_delete(h)
			return nil, err
		}
		nj, err := getLong(h, "Nj")
		if err != nil {
			C.codes_handle_delete(h)
			return nil, err
		}
		values, err := getDoubleArray(h, "values")
		if err != nil {
			C.codes_handle_delete(h)
			return nil, err
		}
		lats, latErr := getDoubleArray(h, "distinctLatitudes")
		lons, lonErr := getDoubleArray(h, "distinctLongitudes")
		name := getString(h, "shortName")
		C.codes_handle_delete(h)

		coords := map[string][]float64{}
		if latErr == nil && len(lats) == int(nj) {
			coords["latitude"] = lats
		}
		if lonErr == nil && len(lons) == int(ni) {
			coords["longitude"] = lons
		}

		da, err := xarray.NewDataArray(
			[]string{"latitude", "longitude"},
			[]int{int(nj), int(ni)}, values, coords, name)
		if err != nil {
			return nil, err
		}
		out = append(out, da)
	}
	return out, nil
}
