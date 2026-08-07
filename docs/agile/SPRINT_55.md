# Sprint 55 — Arithmétique entre Datasets

- **Période** : démarrage 2026-08-07.
- **Objectif** : opérations arithmétiques au niveau Dataset (US-54).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-54 | `Dataset.Add/Sub/Mul/Div` + `AddScalar/MulScalar` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : Add/Mul entre datasets, AddScalar, variable manquante.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Opère **variable par variable** sur les variables de même nom (via l'opération
  DataArray → alignement/broadcasting hérités). Variable absente de l'autre
  dataset → erreur explicite. Scalaires appliqués à toutes les variables.

## Rétrospective

- **Bien** : réutilisation directe de l'arithmétique DataArray ; complète la
  symétrie Dataset/DataArray.
