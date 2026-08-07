# Sprint 45 — Propagation d'opérations au Dataset

- **Période** : démarrage 2026-08-07.
- **Objectif** : cohérence Dataset/DataArray pour stats, données manquantes et
  cumulatives (US-44).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-44 | `Dataset.VarAxis/StdAxis/MedianAxis`, `FillNA`, `Cumsum` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : StdAxis (multi-variables), FillNA, Cumsum (variable partielle).
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Réductions statistiques par axe via `reduceDatasetAxis[T,float64]` (mutualisé
  avec `SumAxis`/`MeanAxis` du Dataset).
- `FillNA`/`Cumsum` appliquées variable par variable ; `Cumsum` conserve les
  variables ne portant pas la dimension. Reconstruction via `NewDataset`.

## Rétrospective

- **Bien** : réutilisation directe des méthodes DataArray et de la propagation
  existante ; comble l'asymétrie Dataset/DataArray sur les stats.
- **Suite possible** : propager rolling/resample/groupby-temporel au Dataset.
