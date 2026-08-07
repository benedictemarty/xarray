# Sprint 50 — Coarsen (downsampling par blocs)

- **Période** : démarrage 2026-08-07.
- **Objectif** : agrégation par blocs non chevauchants (US-49).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-49 | `DataArray.Coarsen` + `Dataset.Coarsen` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : coarsen 1D/2D, trim, Dataset, facteur trop grand.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- `Coarsen` = regroupement par blocs consécutifs d'indices → réutilise `Resample`
  (DataArray) et `DatasetGroupBy` (Dataset). Étiquette = borne gauche de la
  coordonnée du bloc.
- boundary **"trim"** : le reste non divisible est ignoré (cf. xarray).

## Rétrospective

- **Bien** : réutilisation totale de l'infra de regroupement (helper
  `coarsenGroups` partagé DataArray/Dataset). Cas d'usage géo (réduction de
  résolution) couvert.
- **Suite possible** : boundary "pad", coarsen multi-dimensions simultané.
