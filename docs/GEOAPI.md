# Paquet `geoapi` — servir des données xarray-go façon OGC API

Le paquet `github.com/bmarty/xarray/geoapi` fournit des briques pour exposer des
données `xarray-go` à la manière d'un serveur **OGC API** (comme *pygeoapi*, mais
en Go). Il ne dépend que de l'API publique de xarray-go.

> Contexte : *pygeoapi* est écrit en Python et utilise le **vrai** xarray Python.
> `xarray-go` ne s'y branche pas. Ce paquet vise la reconstruction d'un service
> équivalent **en Go**.

## CoverageJSON

`ToCoverageJSON` sérialise un `DataArray[float64]` 2D (grille latitude × longitude)
au format **CoverageJSON** (domaine `Grid`, CRS84) — le format de sortie central
d'OGC API - Coverages / EDR.

```go
da, _ := xarray.NewDataArray(
    []string{"latitude", "longitude"}, []int{Ny, Nx}, data,
    map[string][]float64{"latitude": lats, "longitude": lons}, "t2m")

b, _ := geoapi.ToCoverageJSON(da, "t2m", "longitude", "latitude")
// b : document CoverageJSON (axes x/y, referencing CRS84, ranges NdArray)
```

Le tableau doit avoir ses dimensions dans l'ordre `(latitude, longitude)` (y, x),
cohérent avec l'ordre C et le format NdArray de CoverageJSON.

## Ce qui est couvert par xarray-go (côté données)

Pour un service type coverage/EDR, xarray-go fournit déjà : ouverture
netCDF/Zarr/GRIB, lecture des dimensions/coordonnées, **sous-échantillonnage**
(`SelRange` pour une bbox, `Sel`/`SelNearest` pour un point, sélection temporelle
et de niveau), sélection de variable (`Dataset.Get`), et extraction des valeurs.

## Ce qui reste à ajouter pour un « pygeoapi en Go »

- **Sous-échantillonnage géospatial** de haut niveau (bbox + temps + z) — briques
  présentes (`SelRange`), à emballer dans une API pratique.
- **Endpoints OGC API** (HTTP) : collections, coverage, requêtes EDR
  (position/area/cube).
- **CRS / reprojection** : seul CRS84 (WGS84 lon/lat) est visé pour l'instant.

## Portée / limites

- CoverageJSON : domaine **Grid** 2D lat/lon uniquement (pas encore d'axe temps/z
  dans le domaine, ni PointSeries/Trajectory).
- Sortie float ; valeurs manquantes non encore encodées (`null`).
