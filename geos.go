package xarray

import "math"

// Projection géostationnaire (satellite « geos »), telle qu'utilisée par
// Meteosat/MTG (imageur FCI) : projette lon/lat ↔ coordonnées dans le plan de
// vue du satellite. Formules ellipsoïdales de PROJ, balayage « sweep = y »
// (convention Meteosat/MTG). Sans dépendance PROJ.
//
// Les CRS géostationnaires n'ont pas de code EPSG simple universel (ils sont
// définis par une chaîne PROJ), d'où un type dédié plutôt qu'une entrée du
// registre EPSG.

// Geostationary décrit une projection géostationnaire.
//   - Lon0   : longitude sub-satellite (degrés)
//   - Height : hauteur de perspective au-dessus de l'ellipsoïde (m)
//   - A, B   : demi-grand et demi-petit axes de l'ellipsoïde (m)
type Geostationary struct {
	Lon0, Height, A, B float64
}

// MTGGeos renvoie la projection géostationnaire de MTG-I (sub-satellite 0°,
// ellipsoïde WGS84, hauteur de perspective 35 785 831 m).
func MTGGeos() Geostationary {
	return Geostationary{Lon0: 0, Height: 35785831, A: 6378137, B: 6356752.31414}
}

// Forward projette (lon, lat) en degrés vers (x, y) en mètres dans le plan
// géostationnaire. ok=false si le point n'est pas visible depuis le satellite
// (au-delà du limbe terrestre).
func (g Geostationary) Forward(lon, lat float64) (x, y float64, ok bool) {
	a, b, h := g.A, g.B, g.Height
	radiusP := b / a
	radiusP2 := radiusP * radiusP
	radiusG1 := h / a
	radiusG := 1 + radiusG1

	lam := (lon - g.Lon0) * math.Pi / 180
	latR := lat * math.Pi / 180
	phi := math.Atan(radiusP2 * math.Tan(latR))
	cphi := math.Cos(phi)
	r := radiusP / math.Hypot(radiusP*cphi, math.Sin(phi))
	vx := r * math.Cos(lam) * cphi
	vy := r * math.Sin(lam) * cphi
	vz := r * math.Sin(phi)

	// Visibilité (le point doit être du bon côté du limbe).
	if ((radiusG-vx)*vx - vy*vy - vz*vz/radiusP2) < 0 {
		return math.NaN(), math.NaN(), false
	}
	tmp := radiusG - vx
	// sweep = y (MTG/Meteosat).
	x = a * radiusG1 * math.Atan(vy/tmp)
	y = a * radiusG1 * math.Atan(vz/math.Hypot(vy, tmp))
	return x, y, true
}

// Inverse projette (x, y) en mètres (plan géostationnaire) vers (lon, lat) en
// degrés. ok=false si (x, y) ne correspond à aucun point de la Terre.
func (g Geostationary) Inverse(x, y float64) (lon, lat float64, ok bool) {
	a, b, h := g.A, g.B, g.Height
	radiusP := b / a
	radiusPInv2 := (a / b) * (a / b)
	radiusG1 := h / a
	radiusG := 1 + radiusG1
	c := radiusG*radiusG - 1

	// x/a puis /radiusG1 = x/h : angles de balayage.
	vx := -1.0
	vy := math.Tan((x / a) / radiusG1)
	vz := math.Tan((y/a)/radiusG1) * math.Hypot(1.0, vy)

	aa := vz / radiusP
	aa = vy*vy + aa*aa + vx*vx
	bb := 2 * radiusG * vx
	det := bb*bb - 4*aa*c
	if det < 0 {
		return math.NaN(), math.NaN(), false
	}
	k := (-bb - math.Sqrt(det)) / (2 * aa)
	vx = radiusG + k*vx
	vy *= k
	vz *= k
	lam := math.Atan2(vy, vx)
	phi := math.Atan(vz * math.Cos(lam) / vx)
	phi = math.Atan(radiusPInv2 * math.Tan(phi))
	return lam*180/math.Pi + g.Lon0, phi * 180 / math.Pi, true
}
