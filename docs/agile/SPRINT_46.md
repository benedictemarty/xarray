# Sprint 46 — Dataset.Rolling

- **Période** : démarrage 2026-08-07.
- **Objectif** : propager la fenêtre glissante au niveau Dataset (US-45).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-45 | `Dataset.Rolling` (Mean/Sum/Min/Max) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : rolling multi-variables (2D + 1D), variable partielle conservée.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- `DatasetRolling` applique `DataArray.Rolling` aux variables portant la dimension
  (résultat float64, NaN de bord) ; les autres sont converties. Reconstruction via
  `NewDataset`. Mutualisation par `dsRolling` + closure d'agrégation.

## Rétrospective

- **Bien** : réutilisation directe du `Rolling` de DataArray et du patron de
  propagation (cf. `reduceDatasetAxis`, `Dataset.GroupBy`).
- **Suite possible** : `Dataset.Resample`/`ResampleCalendar`/`GroupByTime`.
