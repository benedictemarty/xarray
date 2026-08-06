# Sprint 12 — concat / stack

- **Période** : démarrage 2026-08-06.
- **Objectif** : restructuration de tableaux (US-21).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-21 | `Concat` (dimension existante) et `Stack` (nouvelle dimension) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Concat 1D/2D (axe 0 et axe 1), stack, cas d'erreur testés.
- [x] `CHANGELOG.md` à jour ; commit atomique.

## Décisions de conception

- `Concat(arrays, dim)` : concaténation le long d'une dimension existante ;
  coordonnée de la dimension = concaténation des coordonnées.
- `Stack(arrays, newDim, labels)` : empilement sur une nouvelle dimension en tête
  (expose la primitive `stackDim`).

## Rétrospective

- **Bien** : `stackDim` déjà écrit au Sprint 9 réutilisé.
- **Suite** : optimisation du calcul vectoriel (Sprint 13).
