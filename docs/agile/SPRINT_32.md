# Sprint 32 — Indexation `sel` avancée

- **Période** : démarrage 2026-08-07.
- **Objectif** : enrichir l'indexation par label au-delà du match exact (US-33).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-33 | `SelNearest` / `SelRange` / `SelMany` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests nearest, range (bornes inversées, vide), many (ordre, absent), erreurs.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- `SelNearest` : minimise |étiquette − cible| (distance en float64) puis `Isel`
  (dimension réduite), comme `sel(method="nearest")` de xarray.
- `SelRange` : garde les positions dont l'étiquette ∈ [lo, hi] via `takeAlong`
  (dimension conservée) ; bornes tolérées dans n'importe quel ordre.
- `SelMany` : match exact de chaque étiquette, ordre de la liste respecté.

## Rétrospective

- **Bien** : réutilisation de `Isel`/`takeAlong` ; API cohérente avec `Sel`.
- **Suite possible** : slices avec `method="nearest"`, tolérance, `sel` par booléen.
