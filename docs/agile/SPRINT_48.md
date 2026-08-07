# Sprint 48 — Restructuration de dimensions

- **Période** : démarrage 2026-08-07.
- **Objectif** : squeeze/expand_dims/rename, manques du domaine restructuration
  (US-47).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-47 | `Squeeze`/`ExpandDims`/`RenameDim` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : squeeze (+ erreur non unitaire), expand (+ aller-retour), rename
      (+ coord suit + sel).
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Opérations de **métadonnées** : les données plates sont inchangées (la taille
  totale est préservée). Reconstruction via `NewDataArray` (validation).
- `Squeeze` retire la coordonnée de la dimension supprimée ; `RenameDim` déplace
  la coordonnée sous le nouveau nom (sa Variable 1D est recréée avec la bonne dim).

## Rétrospective

- **Bien** : `ExpandDims`/`Squeeze` sont inverses ; testé en aller-retour.
- **Suite possible** : `SwapDims` (échanger dim/coord), `stack`/`unstack`
  multi-dimensions.
