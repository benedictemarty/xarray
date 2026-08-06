# Sprint 5 — Généralisation des types (generics)

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : lever la limite `float64` (dette **T-01**) en rendant
  toute la bibliothèque générique sur un type numérique.

## Périmètre engagé

| ID   | Sujet | État |
|------|-------|------|
| T-01 | Types génériques `Variable[T]`, `DataArray[T]`, `Dataset[T]` | ✅ |

## Critères d'acceptation (Definition of Done)

- [x] `gofmt`, `go vet` propres.
- [x] `go test ./...` : tous les tests passent.
- [x] Validation avec au moins un type entier et un flottant non-`float64`.
- [x] `CHANGELOG.md`, backlog et roadmap à jour.
- [x] Commit atomique.

## Décisions de conception

- **Contrainte `Number`** : `~int…~uint…~float32|~float64`.
- **Méthodes sans paramètre de type** : le langage Go l'interdit. Toute opération
  changeant le type de sortie (moyenne → `float64`, conversions de `Dataset`)
  est déléguée à des **fonctions libres génériques** (`reduceAxisVar`,
  `reduceAxisDA`, `reduceDatasetAxis`, `convertNum`, `convertDataArray`).
- **`Mean`/`MeanAxis` en `float64`** : sémantique conforme à xarray (moyenne
  d'entiers = flottant). Les autres réductions restent en `T`.
- **Tableau vide** : `Mean` renvoie `NaN` (float) ; `Min`/`Max` renvoient la
  zéro-valeur de `T`, faute de NaN universel pour les entiers.
- **Conversions numériques** : réalisées via `float64` intermédiaire
  (`R(float64(x))`). Une perte de précision est possible pour de très grands
  entiers — limitation documentée et assumée.

## Impact (BREAKING)

- Les appels de construction infèrent le type depuis les données (souvent sans
  changement de code). Les **lectures** exigent un paramètre de type explicite :
  `ReadDataArrayJSON[float64]`, `ReadDataArrayCSV[int]`, etc.

## Rétrospective

- **Bien** : l'architecture en couches a absorbé le passage aux generics sans
  refonte ; l'inférence de type limite la casse côté appelant.
- **À surveiller** : la contrainte « coordonnées de même type que les données »
  (xarray autorise des types hétérogènes) ; les conversions via `float64`.
- **Suite possible** : benchmarks/perf (T-03), jointures externes, netCDF (US-16).
