# Sprint 56 — Masques de nullité

- **Période** : démarrage 2026-08-07.
- **Objectif** : compléter la gestion des NaN (masques et comptages) (US-55).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-55 | `IsNull`/`NotNull`/`Count`/`CountAxis` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : IsNull/NotNull/Count, CountAxis (2 axes), cas entier.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Masques via `Apply` (1/0 en T). `CountAxis` via `reduceAxisDA[T,float64]`.
  Aucun NaN possible pour les entiers (masque tout à 0, count = taille).

## Rétrospective

- **Bien** : complète `skipna`/`FillNA`/`DropNA` (comptage effectif des présents).
