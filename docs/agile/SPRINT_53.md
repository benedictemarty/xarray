# Sprint 53 — Dot (contraction tensorielle nommée)

- **Période** : démarrage 2026-08-07.
- **Objectif** : produit tensoriel sur dimension nommée (US-52).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-52 | `Dot(a, b, dim)` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : matmul, produit scalaire (0-dim), matrice-vecteur, erreurs.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Contraction sur une dimension commune : résultat = dims restantes de a puis de
  b, somme sur la dimension partagée. Itération générique via offsets pré-calculés
  (`basesOf`, `restOf`). MVP : une seule dimension commune autorisée.

## Rétrospective

- **Bien** : générique N-D (pas limité au 2D) ; couvre matmul/matvec/dot avec des
  dimensions nommées.
- **Suite possible** : contraction sur plusieurs dimensions, gestion des
  dimensions communes non contractées (broadcast).
