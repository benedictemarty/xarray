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
- **simple packing** (template de représentation 5.0) ;
- sans bitmap (tous les points présents) ;
- décodage signe-magnitude des facteurs d'échelle, formule
  `Y = (R + X·2^E) / 10^D`.

**NON géré (limite importante)** :
- **complex packing / second-order** (`grid_second_order`, templates 5.2/5.3) —
  or **c'est le packing de la plupart des fichiers opérationnels** (ECMWF,
  Météo-France…). Ces fichiers ne sont donc PAS lisibles par ce décodeur.
- GRIB **édition 1**, autres grilles (gaussienne, Lambert, stéréographique…),
  bitmaps, compression JPEG2000/PNG, tables de paramètres (pas de `shortName`).

Pour le cas général, la référence est **ecCodes** (ECMWF, bibliothèque C) via
`cfgrib`/cgo. Réimplémenter tout GRIB2 à la main n'est pas réaliste.

## Validation

Le décodeur a été **validé contre ecCodes** :

- un vrai champ (vent `u` à 850 hPa, grille 201×131) réencodé en simple packing
  par ecCodes, puis décodé par `ReadGrib` : les **26 331 valeurs sont identiques**
  (différence maximale absolue = 0,0) ;
- test unitaire autonome (`grib_test.go`) sur un message GRIB2 minimal assemblé à
  la main.

Utilitaire : `cmd/readgrib` (affiche la grille et les valeurs, mode dump pour
comparaison).

## Pourquoi pas le complex packing ?

Le second-order packing découpe les valeurs en groupes de largeurs de bits
variables, avec différenciation spatiale du 1er/2e ordre — un décodage bit-level
délicat et sujet à erreurs, difficile à garantir sans la maturité d'ecCodes. Il a
été volontairement laissé de côté plutôt que livré non fiable. Extension possible :
implémenter les templates 5.2/5.3, ou lier ecCodes via cgo.
