# Sprint 11 — Performance de l'alignement

- **Période** : démarrage 2026-08-06.
- **Objectif** : réduire le coût de l'arithmétique lorsque les opérandes sont déjà
  alignés (cas très fréquent).

## Périmètre

| Sujet | État |
|-------|------|
| Chemin rapide « coordonnées identiques » dans `align` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Gain mesuré sans régression.
- [x] `CHANGELOG.md` à jour ; commit atomique.

## Décision de conception

Lorsqu'une dimension porte des coordonnées **identiques** des deux côtés, aucune
réindexation n'est nécessaire : on saute les deux copies `takeAlong`. Détection
via `sameSlice`.

## Résultat

`Add` 100×100 : **272 µs → 48 µs**, **135 → 18 allocations**. xarray-go passe
devant la couche xarray sur l'arithmétique alignée.

## Rétrospective

- **Bien** : gain majeur pour un risque faible.
- **Suite** : `concat`/`stack` (Sprint 12).
