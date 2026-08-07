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

Pour ces cas, la référence est **ecCodes** (ECMWF). Un **backend ecCodes via
cgo** est fourni (`experimental/eccodesgrib`, opt-in `-tags eccodes`) : il gère
**tout** le GRIB, y compris les templates locaux comme le 50002, en déléguant le
décodage à ecCodes. Voir `experimental/eccodesgrib/README.md`.

### Différence entre le template local 50002 (Météo-France) et le standard 5.3

- **5.2/5.3 = WMO standard**, documenté publiquement : découpage en groupes
  général + différenciation spatiale globale. Décodable à la main (fait ici).
- **50002 = template *local*** (numéros ≥ 50000 réservés aux centres) : portage
  en GRIB2 du « second-order packing » historique (deux niveaux : valeurs de
  premier ordre + résidus de second ordre). Son format binaire exact **n'est pas
  dans la spec WMO publique** ; il vit dans les définitions d'ecCodes. Le décoder
  à la main reviendrait à l'inventer — d'où l'usage du backend ecCodes.

**Validation** : le backend ecCodes lit un vrai fichier 50002 Météo-France avec
des valeurs **identiques** à ecCodes Python (diff max = 0,0).

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
