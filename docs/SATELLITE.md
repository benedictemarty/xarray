# Imagerie satellite / raster géoréférencé avec xarray-go

xarray-go couvre une **chaîne complète d'analyse** de données satellite en
grille (cubes réflectance/radiance, séries temporelles, rasters), chaque maillon
validé contre l'implémentation de référence (zarr-python, lib `affine`).

Périmètre : **lire → dépacker (CF) → géoréférencer → transformer/reprojeter →
découper/analyser → ré-exporter (avec pyramides)**. CRS pris en charge en pur Go
(sans PROJ) : WGS84 (4326), Web Mercator (3857), UTM, Lambert-93, géostationnaire
MTG. Reste hors périmètre : les datums **non-WGS84** (→ PROJ).

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

// 4. Transformer / reprojeter (CRS famille WGS84, sans PROJ).
x, y, _ := xarray.TransformXY("EPSG:4326", "EPSG:3857", 2.35, 48.86) // point
warped, _ := xarray.ReprojectDataArray(geo, srcT, "EPSG:4326",
    dstT, dstW, dstH, "EPSG:3857", "y", "x", xarray.Bilinear)        // raster entier
//   projection géostationnaire MTG : xarray.MTGGeos().Forward(lon, lat)

// 5. Découper par emprise (en coordonnées monde) / extraire un point.
sub, _ := geoapi.SubsetBBox(geo, "x", "y", geoapi.BBox{MinX: 2, MinY: 48, MaxX: 6, MaxY: 51})
val, _ := geoapi.Position(geo, "x", "y", 4.0, 49.5) // plus proche voisin

// 6. Analyser puis ré-exporter (Zarr chunké, pyramides, CoverageJSON).
//    Ex. NDVI = (nir - red) / (nir + red) via l'arithmétique DataArray ;
//    xarray.WriteDatasetZarrChunked("out.zarr", outDS, map[string]int{"y":512,"x":512}, xarray.ZarrZstd)
//    xarray.WritePyramidZarr("pyr.zarr", geo, "y", "x", 4, 2, xarray.ZarrZstd) // overviews
//    geoapi.ToCoverageJSON(...)
```

## Ce qui est géré (et validé)

| Maillon | Détail | Validation |
|---|---|---|
| **Lecture Zarr** | v2 **et** v3, Blosc/LZ4/zstd, byte-/bit-shuffle, chunks multi-blocs, dtypes int/float | fixtures zarr-python (`testdata/`) |
| **Lecture netCDF-4/HDF5** | MTG FCI L1c & co, via convertisseur `nccopy`/`cdo` → CDF-1 | fichier HDF5 réel |
| **Dépacking CF** | `scale_factor`/`add_offset`/`_FillValue` sur Zarr **et** netCDF | `TestZarrReadCFAttrs` |
| **Géoréférencement** | `Affine` (Apply/Inverse/GDAL), `GeoRef`, `Georeference`, `GeoRefFromCF` | lib `affine` (rasterio) |
| **CRS / transformation** | 4326, 3857, UTM (60 zones), Lambert-93, géostationnaire MTG ; `TransformXY` | **pyproj** (< 0,5 mm) |
| **Reprojection raster** | `Reproject`/`ReprojectDataArray`, plus proche voisin **ou bilinéaire** | pyproj + numpy |
| **Découpe** | `SubsetBBox` (emprise monde, axe y décroissant géré), `Position` | `TestSubsetBBoxGeoref` |
| **Analyse** | arithmétique, réductions, rolling/resample, coarsen, out-of-core (`ChunkZarr`) | tests dédiés |
| **Export** | Zarr v2/v3 chunké (zstd/gzip), **pyramides multi-échelles**, netCDF, CoverageJSON | relu par zarr-python |

## Limites assumées

- **CRS : famille WGS84 uniquement** (pas de transformation de datum). Les
  projections gérées (4326/3857/UTM/Lambert-93/geos) partagent WGS84 ou s'en
  approchent (RGF93). Autres datums (ED50, NTF…) → **PROJ**.
- **Reprojection** : plus proche voisin et **bilinéaire** (pas encore cubique).
- **Pyramides** : convention minimale auto-descriptive (pas de conformité
  **OME-Zarr NGFF** complète).
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
