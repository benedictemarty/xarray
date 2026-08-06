# Feuille de route — xarray-go

Livraison incrémentale, un incrément par sprint. Chaque sprint est livrable :
code testé, documentation et CHANGELOG à jour, commits atomiques.

## Sprint 1 — Cœur `Variable` / `DataArray` ✅

Objectif : poser une fondation solide et testée du modèle de données étiqueté.

- `Variable` : données `float64` à plat, dimensions nommées, attributs, `isel`, `At`.
- `DataArray` : coordonnées, `isel`/`sel`, réductions globales, `String`.
- Documentation projet et cadre agile.

**Statut : terminé.**

## Sprint 2 — Opérations ✅

Objectif : rendre les tableaux « calculables ».

- Réductions le long d'une dimension (`SumAxis`, `MeanAxis`, `MinAxis`, `MaxAxis`).
- Arithmétique élément par élément (`Add`, `Sub`, `Mul`, `Div`).
- Broadcasting par nom de dimension.
- Alignement automatique sur les coordonnées (jointure interne).
- `Transpose` ; opérations scalaires (`AddScalar`, `MulScalar`).

**Statut : terminé.**

## Sprint 3 — `Dataset` ✅

Objectif : manipuler des collections cohérentes de variables.

- Type `Dataset` (map de `DataArray` partageant dimensions/coordonnées).
- Propagation de `sel`/`isel` et des réductions par axe.
- Ajout/suppression de variables (`WithVar`, `DropVars`), fusion (`Merge`).

**Statut : terminé.**

## Sprint 4 — Entrées / sorties ✅

Objectif : interopérer avec des formats de fichiers.

- JSON (aller-retour) pour `DataArray` et `Dataset`.
- CSV format « tidy » (aller-retour) pour `DataArray`.
- netCDF : **reporté** (format binaire complexe, dépendance externe requise).

**Statut : terminé** (hors netCDF, reporté au backlog).

## Sprint 5 — Généralisation des types (generics) ✅

Objectif : lever la limite `float64` (dette T-01).

- `Variable[T]`, `DataArray[T]`, `Dataset[T]` avec la contrainte `Number`.
- `Mean`/`MeanAxis` en `float64` ; autres réductions en `T`.
- Validation avec `int`, `float32`, `float64`.

**Statut : terminé.**

## Transverse (continu)

- Couverture de tests, `go vet`, intégration continue.
- Benchmarks (dette T-03).
- Jointures externes pour l'alignement.
