# Paquet `geoapi` — servir des données xarray-go façon OGC API

Le paquet `github.com/bmarty/xarray/geoapi` fournit des briques pour exposer des
données `xarray-go` à la manière d'un serveur **OGC API** (comme *pygeoapi*, mais
en Go). Il ne dépend que de l'API publique de xarray-go.

> **Positionnement (important).** Il existe déjà des serveurs OGC API **en Go** —
> notamment **gogeoapi** (projet distinct). Le rôle de `xarray-go` n'est **pas** de
> réimplémenter un serveur OGC API, mais de fournir la **couche de données**
> (« provider ») : lire netCDF/Zarr/GRIB, sous-échantillonner (bbox, temps, z,
> variable), extraire les valeurs, produire du CoverageJSON.
>
> Autrement dit, la bonne architecture est :
>
>     gogeoapi (serveur OGC API, HTTP)  ──►  xarray-go + geoapi (provider de données)
>
> exactement comme *pygeoapi* (serveur Python) s'appuie sur *xarray* (données).
> Ce paquet `geoapi` regroupe les briques de **provider**, pas le serveur.

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

## Fonctions de provider fournies

- **`SubsetBBox(da, xDim, yDim, bbox)`** : sous-cube dans une emprise (requête
  area/bbox), via `SelRange`.
- **`Position(da, xDim, yDim, x, y)`** : valeur au point le plus proche (requête
  EDR *position*), via `SelNearest`.
- **`ToCoverageJSON(...)`** : sérialisation de sortie.

## Ce qui reste (côté serveur — plutôt du ressort de gogeoapi)

- **Endpoints OGC API** (HTTP) : collections, coverage, requêtes EDR — c'est le
  rôle d'un serveur comme **gogeoapi**, pas de xarray-go.
- **CRS / reprojection** : seul CRS84 (WGS84 lon/lat) est visé.
- Sélection **temporelle** et de **niveau z** : disponibles via `Sel`/`SelRange`
  de xarray-go sur les dimensions correspondantes.

## Portée / limites

- CoverageJSON : domaine **Grid** 2D lat/lon uniquement (pas encore d'axe temps/z
  dans le domaine, ni PointSeries/Trajectory).
- Sortie float ; valeurs manquantes non encore encodées (`null`).
