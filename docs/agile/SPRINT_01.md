# Sprint 1 — Cœur `Variable` / `DataArray`

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : disposer d'un modèle de données N-dimensionnel
  étiqueté, testé, permettant l'indexation par position et par label.

## Périmètre engagé

| ID    | User story | État |
|-------|------------|------|
| US-01 | Tableau N-D à dimensions nommées | ✅ |
| US-02 | Coordonnées attachées aux dimensions | ✅ |
| US-03 | Indexation `isel` (position) et `sel` (label) | ✅ |
| US-04 | Représentation textuelle lisible | ✅ |
| US-05 | Attributs / métadonnées | ✅ |

## Critères d'acceptation (Definition of Done)

- [x] `go vet ./...` sans avertissement.
- [x] `go test ./...` : tous les tests passent.
- [x] Cas nominaux **et** cas d'erreur couverts par des tests.
- [x] `CHANGELOG.md` mis à jour.
- [x] Documentation (README, architecture) à jour.
- [x] Modifications commitées atomiquement dans git.

## Réalisé

- `variable.go` : type `Variable`, construction validée, `At`, `Isel`, `String`,
  attributs.
- `dataarray.go` : type `DataArray`, coordonnées, `Isel`, `Sel`, réductions
  (`Sum`, `Mean`, `Min`, `Max`), `Rename`, `String`.
- `variable_test.go`, `dataarray_test.go` : suites de tests.
- Documentation projet et cadre agile.

## Revue / rétrospective

- **Bien** : architecture en trois couches (Variable → DataArray → Dataset)
  claire et extensible ; base de tests solide dès le départ.
- **À surveiller** : le stockage limité à `float64` (dette T-01) ; le broadcasting
  et l'alignement, cœur de xarray, arrivent au Sprint 2 et conditionnent
  l'ergonomie réelle.
- **Prochain sprint** : opérations (US-06 à US-10).
