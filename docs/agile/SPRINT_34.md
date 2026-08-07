# Sprint 34 — Réductions statistiques et cumulatives

- **Période** : démarrage 2026-08-07.
- **Objectif** : compléter les réductions manquantes face à xarray (US-35).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-35 | `Var`/`Std`/`Median`, `Cumsum`, `Diff` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests Var/Std/Median (pair/impair), VarAxis, Cumsum 1D/2D, Diff.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Variance de **population** (ddof=0), comme le défaut de xarray ; `Std` = √Var.
- `Median` par tri d'une copie (moyenne des deux centraux si taille paire).
- Réductions par axe via `reduceAxisDA[T,float64]` (réutilisation).
- `Cumsum` : parcours par ligne (`forEachLine`), accumulateur.
- `Diff` : soustraction **positionnelle** (via `binaryOp`, sans alignement, car
  les coordonnées sont décalées) entre `takeAlong(1..n-1)` et `takeAlong(0..n-2)`.

## Rétrospective

- **Bien** : forte réutilisation (`reduceAxisDA`, `forEachLine`, `takeAlong`,
  `binaryOp`).
- **Suite possible** : `quantile`, `cumprod`, `argmin`/`argmax`, ddof
  paramétrable.
