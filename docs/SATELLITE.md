# Imagerie satellite / raster géoréférencé avec xarray-go

xarray-go couvre une **chaîne complète d'analyse** de données satellite en
grille (cubes réflectance/radiance, séries temporelles, rasters), chaque maillon
validé contre l'implémentation de référence (zarr-python, lib `affine`).

Périmètre : **lire → dépacker (CF) → géoréférencer → découper/analyser →
ré-exporter**. Hors périmètre : **reprojection** (changer de CRS, nécessiterait
PROJ) et **pyramides/overviews** (visualisation multi-échelles).

## La chaîne de bout en bout

```go
import "github.com/benedictemarty/xarray"

// 1. Lire — Zarr v2/v3 (Blosc/LZ4/zstd, int16/uint16…) ou netCDF-4/HDF5.
ds, _ := xarray.ReadDatasetZarr("scene.zarr")
//   netCDF-4/HDF5 (MTG FCI, etc.) : xarray.OpenNetCDFFile("scene.nc", nil)
//   (conversion via nccopy/cdo si le fichier n'est pas déjà CDF-1).

// 2. Dépacking CF — int16 packé → grandeur physique (réflectance…).
//    Lit scale_factor/add_offset/_FillValue depuis les attributs (.zattrs / CF).
ds, _ = xarray.DecodeCF(ds)

// 3. Géoréférencement — lire le CRS + l'affine depuis les métadonnées CF
//    (convention rioxarray : grid_mapping → variable spatial_ref).
gr, ok := ds.GeoRefFromCF("B04")           // ou GeoRef{Transform, CRS} à la main
b, _ := ds.Get("B04")
geo, _ := b.Georeference(gr, "x", "y")     // attache x/y monde (centres de pixels) + CRS

// 4. Découper par emprise (en coordonnées monde) / extraire un point.
sub, _ := geoapi.SubsetBBox(geo, "x", "y", geoapi.BBox{MinX: 2, MinY: 48, MaxX: 6, MaxY: 51})
val, _ := geoapi.Position(geo, "x", "y", 4.0, 49.5) // plus proche voisin

// 5. Analyser (maths sur tableaux, réductions, out-of-core) puis ré-exporter.
//    Ex. NDVI = (nir - red) / (nir + red) via l'arithmétique DataArray ;
//    export Zarr chunké : xarray.WriteDatasetZarrChunked("out.zarr", outDS,
//        map[string]int{"y": 512, "x": 512}, xarray.ZarrZstd)
//    ou CoverageJSON : geoapi.ToCoverageJSON(...).
```

## Ce qui est géré (et validé)

| Maillon | Détail | Validation |
|---|---|---|
| **Lecture Zarr** | v2 **et** v3, Blosc/LZ4/zstd, byte-/bit-shuffle, chunks multi-blocs, dtypes int/float | fixtures zarr-python (`testdata/`) |
| **Lecture netCDF-4/HDF5** | MTG FCI L1c & co, via convertisseur `nccopy`/`cdo` → CDF-1 | fichier HDF5 réel |
| **Dépacking CF** | `scale_factor`/`add_offset`/`_FillValue` sur Zarr **et** netCDF | `TestZarrReadCFAttrs` |
| **Géoréférencement** | `Affine` (Apply/Inverse/GDAL), `GeoRef`, `Georeference`, `GeoRefFromCF` | lib `affine` (rasterio) |
| **Découpe** | `SubsetBBox` (emprise monde, axe y décroissant géré), `Position` | `TestSubsetBBoxGeoref` |
| **Analyse** | arithmétique, réductions, rolling/resample, coarsen, out-of-core (`ChunkZarr`) | tests dédiés |
| **Export** | Zarr v2/v3 chunké (zstd/gzip), netCDF, CoverageJSON | relu par zarr-python |

## Limites assumées

- **Pas de reprojection** : le CRS est transporté comme identifiant opaque
  (`EPSG:xxxx`/WKT), pas de changement de projection (→ PROJ/pyproj requis).
- **Pas d'overviews/pyramides** (OME-Zarr multiscale).
- Géoréférencement **axis-aligned** : une grille avec rotation (affine `B`/`D`≠0)
  n'est pas séparable en axes 1D — refusée par erreur explicite (jamais de
  résultat faux silencieux).
- Données ramenées en **float64** en interne (int16 → ×4 en RAM).
- `bitshuffle` avec un nombre d'éléments par bloc non multiple de 8 : non géré
  (erreur explicite).

## Accès aux données MTG (EUMETSAT)

MTG-I FCI L1c est distribué en **netCDF-4/HDF5** sur le Data Store EUMETSAT
(collections `EO:EUM:DAT:0662` / `0665`), via le client `eumdac` et des
identifiants OAuth2 (clé + secret, inscription gratuite). Une fois un chunk
téléchargé, `xarray.OpenNetCDFFile` l'ouvre (conversion `nccopy`). Le
téléchargement exige d'avoir **accepté la licence de la collection** sur le
portail — sans quoi l'API renvoie `403` malgré une authentification valide.
