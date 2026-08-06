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

## Sprint 4 — Entrées / sorties

Objectif : interopérer avec des formats de fichiers.

- CSV et JSON (aller-retour).
- netCDF (lecture/écriture) — exploratoire.

## Transverse (continu)

- Couverture de tests, `go vet`, intégration continue.
- Benchmarks.
- Étude de la généralisation du type de données (generics).
