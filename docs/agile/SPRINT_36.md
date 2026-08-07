# Sprint 36 — IdxMin / IdxMax

- **Période** : démarrage 2026-08-07.
- **Objectif** : renvoyer l'étiquette de coordonnée à l'extremum (US-37).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-37 | `IdxMinAxis`/`IdxMaxAxis` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests 1D/2D + cas sans coordonnée.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Le réducteur (fourni à `reduceAxisDA`) capture la coordonnée de `dim` par
  closure : il calcule l'indice de l'extremum dans la tranche puis renvoie
  `labels[indice]`. Type de sortie `T` (les étiquettes sont de type `T`).

## Rétrospective

- **Bien** : complète `ArgMin/MaxAxis` (indice) par la variante « étiquette ».
