# Sprint 24 — Prise en charge de Zarr v2

- **Période** : démarrage 2026-08-06.
- **Objectif** : lire/écrire le format Zarr (v2), stockage moderne des tableaux
  N-D chunkés (US-25).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-25 | Lecture/écriture Zarr v2 (`DataArray[float64]`) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Aller-retour testé : chunks alignés/non alignés, compression none/zlib, erreurs.
- [x] **Interopérabilité vérifiée** avec zarr-python (les deux sens).
- [x] `CHANGELOG.md` / `docs/ZARR.md` à jour ; commit atomique.

## Décisions de conception

- **Zarr v2**, store filesystem, `float64` (`<f8`), ordre C. Chunks de bord
  complétés par `fill_value` (convention Zarr : tous les chunks ont la taille de
  chunk). Chunks manquants = `fill_value`.
- Compression **none** ou **zlib** (bibliothèque standard Go ; l'id numcodecs
  "zlib" produit un flux zlib compatible). blosc/zstd non gérés (hors stdlib).
- Dimensions/nom/coordonnées dans `.zattrs` (`_ARRAY_DIMENSIONS`).

## Validation d'interopérabilité (point fort)

Contrairement au netCDF (auto-cohérent seulement), l'interop Zarr a été
**réellement vérifiée** avec zarr-python 3.3.0 / numpy 2.5.1 :

- **Go → Python** : store 5×4, chunks 2×3, zlib → relu identique par `zarr.open`.
- **Python → Go** : store créé par `zarr.create(zarr_format=2)` → relu identique
  par `ReadDataArrayZarr`.

## Rétrospective

- **Bien** : interop bidirectionnelle prouvée ; zlib via stdlib évite toute
  dépendance.
- **À étendre** : `Dataset` comme groupe Zarr (`.zgroup` + sous-arrays), autres
  dtypes, compresseurs blosc/zstd, Zarr v3.
