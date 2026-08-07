# Sprint 35 — ArgMin/ArgMax, Quantile, Cumprod

- **Période** : démarrage 2026-08-07.
- **Objectif** : compléter les réductions (US-36).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-36 | `ArgMinAxis`/`ArgMaxAxis`, `Quantile`/`QuantileAxis`, `Cumprod` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests argmin/argmax par axe, quantile (bornes + interpolation), cumprod.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- `ArgMin/MaxAxis` renvoient l'**indice** (float64) le long de l'axe (comme
  `argmin`/`argmax` de xarray), via `reduceAxisDA[T,float64]`.
- `Quantile` : interpolation **linéaire** entre rangs (méthode « linear » NumPy) ;
  bornes q≤0 → min, q≥1 → max.
- `Cumprod` : accumulateur multiplicatif via `forEachLine` (init 1).

## Rétrospective

- **Bien** : réutilisation totale de l'infrastructure (`reduceAxisDA`,
  `forEachLine`).
- **Suite possible** : `idxmin`/`idxmax` (label de coordonnée), quantiles
  multiples, ddof paramétrable.
