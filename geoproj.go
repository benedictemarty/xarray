package xarray

import (
	"fmt"
	"math"
	"strconv"
)

// Cadre de projections cartographiques (formules fermées, ellipsoïde WGS84, sans
// PROJ). Chaque CRS projeté sait convertir lon/lat ↔ coordonnées projetées ; la
// transformation entre deux CRS passe par le géographique WGS84 (lon/lat) comme
// pivot. Convient aux CRS de la famille WGS84 (pas de décalage de datum) :
// 4326 (géographique), 3857 (Web Mercator), UTM 326xx/327xx.
//
// Datum : pas de transformation de datum (les CRS gérés partagent WGS84, ou
// s'en approchent à quelques cm comme RGF93). Pour un datum différent, PROJ.

const (
	wgs84A = 6378137.0           // demi-grand axe (m)
	wgs84F = 1.0 / 298.257223563 // aplatissement
)

// projection convertit entre lon/lat (degrés) et coordonnées projetées (mètres).
type projection interface {
	forward(lon, lat float64) (x, y float64)
	inverse(x, y float64) (lon, lat float64)
}

// webMercatorProj : EPSG:3857 (déjà en formules fermées dans geocrs.go).
type webMercatorProj struct{}

func (webMercatorProj) forward(lon, lat float64) (float64, float64) {
	return WebMercatorForward(lon, lat)
}
func (webMercatorProj) inverse(x, y float64) (float64, float64) { return WebMercatorInverse(x, y) }

// transverseMercator : Mercator transverse ellipsoïdal (base des zones UTM).
// lon0 (méridien central) en degrés ; k0 facteur d'échelle ; fe/fn faux est/nord.
type transverseMercator struct {
	lon0, k0, fe, fn float64
}

func meridionalArc(lat, a, e2 float64) float64 {
	return a * ((1-e2/4-3*e2*e2/64-5*e2*e2*e2/256)*lat -
		(3*e2/8+3*e2*e2/32+45*e2*e2*e2/1024)*math.Sin(2*lat) +
		(15*e2*e2/256+45*e2*e2*e2/1024)*math.Sin(4*lat) -
		(35*e2*e2*e2/3072)*math.Sin(6*lat))
}

func (t transverseMercator) forward(lon, lat float64) (float64, float64) {
	a, e2 := wgs84A, wgs84F*(2-wgs84F)
	ep2 := e2 / (1 - e2)
	latR := lat * math.Pi / 180
	dLon := (lon - t.lon0) * math.Pi / 180
	sinp, cosp, tanp := math.Sin(latR), math.Cos(latR), math.Tan(latR)
	N := a / math.Sqrt(1-e2*sinp*sinp)
	T := tanp * tanp
	C := ep2 * cosp * cosp
	A := dLon * cosp
	M := meridionalArc(latR, a, e2)
	x := t.fe + t.k0*N*(A+(1-T+C)*A*A*A/6+(5-18*T+T*T+72*C-58*ep2)*math.Pow(A, 5)/120)
	y := t.fn + t.k0*(M+N*tanp*(A*A/2+(5-T+9*C+4*C*C)*math.Pow(A, 4)/24+(61-58*T+T*T+600*C-330*ep2)*math.Pow(A, 6)/720))
	return x, y
}

func (t transverseMercator) inverse(x, y float64) (float64, float64) {
	a, e2 := wgs84A, wgs84F*(2-wgs84F)
	ep2 := e2 / (1 - e2)
	M := (y - t.fn) / t.k0
	mu := M / (a * (1 - e2/4 - 3*e2*e2/64 - 5*e2*e2*e2/256))
	e1 := (1 - math.Sqrt(1-e2)) / (1 + math.Sqrt(1-e2))
	phi := mu + (3*e1/2-27*e1*e1*e1/32)*math.Sin(2*mu) +
		(21*e1*e1/16-55*e1*e1*e1*e1/32)*math.Sin(4*mu) +
		(151*e1*e1*e1/96)*math.Sin(6*mu) +
		(1097*e1*e1*e1*e1/512)*math.Sin(8*mu)
	sinp, cosp, tanp := math.Sin(phi), math.Cos(phi), math.Tan(phi)
	C1 := ep2 * cosp * cosp
	T1 := tanp * tanp
	N1 := a / math.Sqrt(1-e2*sinp*sinp)
	R1 := a * (1 - e2) / math.Pow(1-e2*sinp*sinp, 1.5)
	D := (x - t.fe) / (N1 * t.k0)
	lat := phi - (N1*tanp/R1)*(D*D/2-
		(5+3*T1+10*C1-4*C1*C1-9*ep2)*math.Pow(D, 4)/24+
		(61+90*T1+298*C1+45*T1*T1-252*ep2-3*C1*C1)*math.Pow(D, 6)/720)
	lon := t.lon0*math.Pi/180 + (D-(1+2*T1+C1)*D*D*D/6+
		(5-2*C1+28*T1-3*C1*C1+8*ep2+24*T1*T1)*math.Pow(D, 5)/120)/cosp
	return lon * 180 / math.Pi, lat * 180 / math.Pi
}

// projectionFor renvoie la projection d'un code EPSG normalisé, et un booléen
// « géographique » (auquel cas x/y sont lon/lat, sans projection).
func projectionFor(epsg string) (projection, bool, error) {
	switch epsg {
	case "4326":
		return nil, true, nil
	case "3857":
		return webMercatorProj{}, false, nil
	}
	// UTM : 326zz (nord) / 327zz (sud), zz = numéro de zone 01..60.
	if len(epsg) == 5 && (epsg[:3] == "326" || epsg[:3] == "327") {
		zone, err := strconv.Atoi(epsg[3:])
		if err == nil && zone >= 1 && zone <= 60 {
			fn := 0.0
			if epsg[:3] == "327" {
				fn = 10000000.0 // hémisphère sud
			}
			return transverseMercator{lon0: float64(6*zone - 183), k0: 0.9996, fe: 500000, fn: fn}, false, nil
		}
	}
	return nil, false, fmt.Errorf("xarray: CRS EPSG:%s non pris en charge (utilisez PROJ)", epsg)
}
