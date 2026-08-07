# Prise en charge de GRIB2 (lecture)

`xarray-go` lit un **sous-ensemble de GRIB2** (format OMM de données en grille,
très utilisé en météorologie).

## API

```go
msgs, _ := xarray.ReadGrib(f)          // f : io.Reader sur un fichier .grib2
da, _ := msgs[0].ToDataArray("champ")  // -> DataArray[float64] (latitude, longitude)
```

## Périmètre

**Géré** :
- GRIB **édition 2** ;
- grille **régulière lat/lon** (`regular_ll`, template de grille 3.0) ;
- **simple packing** (template 5.0) ;
- **complex packing** (template 5.2) et **complex packing + différenciation
  spatiale** d'ordre 1 ou 2 (template 5.3) — avec alignement à l'octet des blocs
  (références, largeurs, longueurs de groupe), conformément à g2clib ;
- sans bitmap (tous les points présents) ;
- décodage signe-magnitude des facteurs d'échelle, formule `Y = (R + X·2^E) / 10^D`.

**NON géré** :
- **templates de packing *locaux*** (ex. **50002**, propriétaire Météo-France) —
  non documentés publiquement, ils **requièrent ecCodes**. C'est le cas de
  certains fichiers opérationnels.
- GRIB **édition 1**, autres grilles (gaussienne, Lambert…), bitmaps, compression
  JPEG2000/PNG, tables de paramètres (pas de `shortName`).

Pour ces cas, la référence est **ecCodes** (ECMWF) via `cfgrib`/cgo.

## Validation

Le décodeur a été **validé contre ecCodes** :

- **simple packing** : un vrai champ (vent `u` 850 hPa, 201×131) réencodé en
  simple packing puis décodé par `ReadGrib` → **26 331 valeurs identiques**
  (diff max = 0,0) ;
- **complex packing + différenciation spatiale (template 5.3)** : le même champ
  réencodé en `grid_complex_spatial_differencing` → **26 331 valeurs identiques**
  (diff max = 0,0) ;
- **tests unitaires versionnés** : message simple assemblé à la main
  (`grib_test.go`) et fichier complex synthétique `testdata/complex_synth.grib2`
  (`grib_complex_test.go`).

Utilitaire : `cmd/readgrib`.

## Note sur l'implémentation du complex packing

L'algorithme suit g2clib (`comunpack`). Le point subtil, source d'erreurs : les
blocs de la section 7 (références, largeurs, longueurs de groupe) sont **alignés à
l'octet** (padding après chaque bloc) ; les longueurs sont lues pour **tous** les
groupes, la dernière étant remplacée par « true length of last group ». La
différenciation spatiale est inversée en ajoutant le minimum global puis par
sommation récursive (ordre 1) ou double récurrence (ordre 2).
