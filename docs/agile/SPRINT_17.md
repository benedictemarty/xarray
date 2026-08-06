# Sprint 17 — skipna

- **Période** : démarrage 2026-08-06.
- **Objectif** : réductions ignorant les NaN (US-22), comportement par défaut de
  xarray.

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-22 | Réductions `skipna` (globales et par axe) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Cas avec NaN, tout-NaN, par axe, type entier testés.
- [x] `CHANGELOG.md` à jour ; commit atomique.

## Décisions de conception

- `SumSkipNA`/`MeanSkipNA`/`MinSkipNA`/`MaxSkipNA` et variantes `*AxisSkipNA`.
- `math.IsNaN(float64(x))` : toujours faux pour les entiers → comportement
  identique aux réductions normales pour les types entiers.
- `MeanSkipNA` divise par le nombre de valeurs non-NaN.

## Rétrospective

- **Suite** : `Dataset.GroupBy` (Sprint 18).
