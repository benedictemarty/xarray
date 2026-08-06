# Sprint 18 — Dataset.GroupBy

- **Période** : démarrage 2026-08-06.
- **Objectif** : regroupement au niveau `Dataset` (US-23).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-23 | `Dataset.GroupBy(dim)` + agrégations | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Regroupement propagé (variable multi-dim et variable partielle) testé.
- [x] `CHANGELOG.md` à jour ; commit atomique.

## Décisions de conception

- Regroupement par la coordonnée partagée de `dim` ; agrégation propagée aux
  variables portant la dimension, les autres conservées (converties si besoin).
- Refactorisation : `groupReduceOn` (réduction de groupe découplée de la
  coordonnée propre du tableau) réutilisée par `DataArray.GroupBy` et
  `Dataset.GroupBy`.

## Rétrospective

- **Bien** : mutualisation propre entre les deux niveaux de groupby.
- **Suite** : noyaux directs float64 sur toute l'arithmétique (Sprint 19).
