# Changelog

Toutes les modifications notables de ce projet sont documentées dans ce fichier.

Le format s'appuie sur [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/)
et le projet suit le [versionnage sémantique](https://semver.org/lang/fr/).

## [Non publié]

### Ajouté (Sprint 3 — `Dataset`)

- **Type `Dataset`** : collection de `DataArray` (« variables de données »)
  partageant un système commun de dimensions et de coordonnées.
  - Construction validée (`NewDataset`) : vérifie la cohérence des tailles de
    dimensions et l'identité des coordonnées partagées entre variables.
  - Accès : `VarNames`, `Get`, `Dims`, `Coord`.
  - Indexation propagée : `Isel` (position) et `Sel` (label via coordonnée
    partagée) appliquées à toutes les variables portant la dimension visée.
  - Réductions propagées : `SumAxis`, `MeanAxis`, `MinAxis`, `MaxAxis` (les
    variables sans la dimension restent inchangées).
  - Gestion des variables : `WithVar`, `DropVars`, `Merge`.
  - Représentation lisible via `String`.
- Helper `DataArray.HasDim`.
- Tests : cohérence, indexation et réductions propagées (y compris dimension
  partielle), fusion, cas d'erreur.

### Ajouté (Sprint 2 — Opérations)

- **`Transpose`** (sur `Variable` et `DataArray`) : réordonnancement des axes par
  permutation des noms de dimensions ; coordonnées conservées.
- **Réductions par axe** (`SumAxis`, `MeanAxis`, `MinAxis`, `MaxAxis`) : réduisent
  une dimension nommée et retirent sa coordonnée du résultat.
- **Arithmétique entre `DataArray`** (`Add`, `Sub`, `Mul`, `Div`) avec :
  - **broadcasting par nom de dimension** (et non par position) ;
  - **alignement automatique** sur les coordonnées (jointure interne sur les
    étiquettes communes) avant l'opération.
- **Opérations scalaires** (`AddScalar`, `MulScalar`) préservant les coordonnées.
- Primitives bas niveau : `Variable.take` (sélection multi-positions),
  `binaryOp` (broadcasting), `reduceAxis`, `mapScalar`.
- Tests : broadcasting, alignement, réductions par axe (2D et 3D), scalaires,
  cas d'erreur (tailles incompatibles, absence d'étiquette commune).

### Ajouté (Sprint 1 — Cœur `Variable` / `DataArray`)

- **Type `Variable`** : tableau N-dimensionnel bas niveau (données `float64` à plat
  en ordre C, dimensions nommées, attributs).
  - Construction validée (`NewVariable`), propriétés (`Dims`, `Shape`, `Ndim`,
    `Size`, `Data`, `Attrs`).
  - Indexation positionnelle : `At`, `Isel` (réduction d'une dimension).
  - Représentation lisible via `String`.
- **Type `DataArray`** : `Variable` + coordonnées étiquetées + nom.
  - Construction validée (`NewDataArray`) avec coordonnées de dimension.
  - Indexation par position (`Isel`) et par label (`Sel`).
  - Réductions globales : `Sum`, `Mean`, `Min`, `Max`.
  - Copie profonde (`clone`), `Rename`, accès aux coordonnées (`Coord`).
- **Tests** : couverture des cas nominaux et d'erreur pour `Variable` et `DataArray`.
- **Documentation projet** : README, backlog produit, roadmap, note de sprint,
  document d'architecture.

## [0.0.0] — 2026-08-06

### Ajouté

- Initialisation du projet : module Go `github.com/bmarty/xarray`, dépôt git,
  cadre de gestion agile.
