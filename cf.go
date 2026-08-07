package xarray

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Décodage minimal des conventions CF (Climate and Forecast) à la lecture, à la
// manière de xarray.decode_cf. Couvre le « packing » (scale_factor / add_offset),
// les valeurs manquantes (_FillValue / missing_value) et le temps
// (« <unité> since <date> »). Ce n'est pas un décodage CF complet.

// clés d'attributs consommées puis retirées lors du décodage du packing.
var cfDecodeConsumed = map[string]bool{
	"scale_factor":  true,
	"add_offset":    true,
	"_FillValue":    true,
	"missing_value": true,
}

// DecodeCF applique le décodage du packing CF à toutes les variables d'un
// Dataset[float64] : valeur = brut × scale_factor + add_offset, et les valeurs
// égales à _FillValue / missing_value deviennent NaN. Les attributs consommés
// sont retirés (la variable est désormais « décodée »).
func DecodeCF(ds *Dataset[float64]) (*Dataset[float64], error) {
	out := make(map[string]*DataArray[float64], len(ds.vars))
	for _, name := range ds.VarNames() {
		da, err := ds.Get(name)
		if err != nil {
			return nil, err
		}
		dec, err := decodeCFVar(da)
		if err != nil {
			return nil, err
		}
		out[name] = dec
	}
	return NewDataset(out)
}

// decodeCFVar décode le packing d'une variable ; renvoie la variable inchangée
// si aucun attribut de packing n'est présent.
func decodeCFVar(da *DataArray[float64]) (*DataArray[float64], error) {
	attrs := da.variable.Attrs()
	scaleS, hasScale := attrs["scale_factor"]
	offS, hasOff := attrs["add_offset"]
	fillS, hasFill := attrs["_FillValue"]
	if !hasFill {
		fillS, hasFill = attrs["missing_value"]
	}
	if !hasScale && !hasOff && !hasFill {
		return da, nil
	}

	scale, off := 1.0, 0.0
	if hasScale {
		if f, err := strconv.ParseFloat(scaleS, 64); err == nil {
			scale = f
		}
	}
	if hasOff {
		if f, err := strconv.ParseFloat(offS, 64); err == nil {
			off = f
		}
	}
	var fill float64
	if hasFill {
		if f, err := strconv.ParseFloat(fillS, 64); err == nil {
			fill = f
		} else {
			hasFill = false
		}
	}

	src := da.variable.data
	data := make([]float64, len(src))
	for i, v := range src {
		if hasFill && v == fill {
			data[i] = math.NaN()
			continue
		}
		data[i] = v*scale + off
	}

	coords := map[string][]float64{}
	for d, cv := range da.coords {
		coords[d] = append([]float64(nil), cv.data...)
	}
	nd, err := NewDataArray(da.variable.Dims(), da.variable.Shape(), data, coords, da.name)
	if err != nil {
		return nil, err
	}
	for k, val := range attrs {
		if cfDecodeConsumed[k] {
			continue
		}
		nd.variable.SetAttr(k, val)
	}
	for d, cv := range nd.coords {
		for k, val := range da.coords[d].attrs {
			cv.SetAttr(k, val)
		}
	}
	return nd, nil
}

// DecodeTime décode la coordonnée temporelle dim d'un Dataset selon son
// attribut « units » CF (« <unité> since <date> ») et la remplace par des
// secondes depuis l'epoch Unix. Sans attribut units exploitable, le Dataset est
// renvoyé inchangé.
func DecodeTime(ds *Dataset[float64], dim string) (*Dataset[float64], error) {
	cv, ok := ds.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée %q", dim)
	}
	units := cv.attrs["units"]
	if !strings.Contains(units, " since ") {
		return ds, nil
	}
	decoded, err := DecodeCFTime(cv.data, units)
	if err != nil {
		return nil, err
	}
	out := make(map[string]*DataArray[float64], len(ds.vars))
	for _, name := range ds.VarNames() {
		da, err := ds.Get(name)
		if err != nil {
			return nil, err
		}
		if !da.HasDim(dim) {
			out[name] = da
			continue
		}
		coords := map[string][]float64{}
		for d, c := range da.coords {
			if d == dim {
				coords[d] = decoded
			} else {
				coords[d] = append([]float64(nil), c.data...)
			}
		}
		nd, err := NewDataArray(da.variable.Dims(), da.variable.Shape(), da.variable.Data(), coords, name)
		if err != nil {
			return nil, err
		}
		for k, val := range da.variable.Attrs() {
			nd.variable.SetAttr(k, val)
		}
		out[name] = nd
	}
	return NewDataset(out)
}

// DecodeCFTime convertit des valeurs temporelles CF (« <unité> since <date> »)
// en secondes depuis l'epoch Unix. Unités reconnues : seconds, minutes, hours,
// days (et abréviations usuelles).
func DecodeCFTime(values []float64, units string) ([]float64, error) {
	factor, ref, err := parseCFTimeUnits(units)
	if err != nil {
		return nil, err
	}
	refEpoch := EpochSeconds(ref)
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = refEpoch + v*factor
	}
	return out, nil
}

// parseCFTimeUnits analyse « <unité> since <date> » en (secondes par unité, date
// de référence).
func parseCFTimeUnits(units string) (float64, time.Time, error) {
	parts := strings.SplitN(units, " since ", 2)
	if len(parts) != 2 {
		return 0, time.Time{}, fmt.Errorf("xarray: units temporelles CF invalides %q (attendu « <unité> since <date> »)", units)
	}
	var factor float64
	switch strings.TrimSpace(strings.ToLower(parts[0])) {
	case "seconds", "second", "secs", "sec", "s":
		factor = 1
	case "minutes", "minute", "mins", "min":
		factor = 60
	case "hours", "hour", "hrs", "hr", "h":
		factor = 3600
	case "days", "day", "d":
		factor = 86400
	default:
		return 0, time.Time{}, fmt.Errorf("xarray: unité temporelle CF non prise en charge %q", parts[0])
	}
	ref, err := parseCFDate(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, time.Time{}, err
	}
	return factor, ref, nil
}

// parseCFDate tente plusieurs formats de date de référence CF usuels.
func parseCFDate(s string) (time.Time, error) {
	layouts := []string{
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("xarray: date de référence CF non reconnue %q", s)
}
