# Sprint 40 — Réductions par axe en lazy

- **Période** : démarrage 2026-08-07.
- **Objectif** : réductions par dimension sur un LazyArray, en streaming (US-41).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-41 | `SumAxis`/`MeanAxis`/`MinAxis`/`MaxAxis` sur LazyArray (1D/2D) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : axe 0 (accumulation) / axe 1 (par ligne), mean/min/max, 1D
      (scalaire), cohérence lazy vs direct.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- **Axe 0** (celui du découpage) : accumulateur par colonne combiné chunk par
  chunk. **Axe 1** : réduction par ligne à l'intérieur de chaque bloc. **1D** :
  réduction totale → DataArray 0-dimension.
- Parcours **séquentiel** des chunks (accumulation par colonne partagée) — simple
  et correct ; reste out-of-core (un chunk à la fois). Coordonnées de l'axe
  restant préservées.
- Limité aux tableaux **1D/2D** (cohérent avec les sources chunk/zarr).

## Rétrospective

- **Bien** : complète le moteur lazy (opérations, combinaison, réductions
  globales et désormais par axe) ; test de cohérence avec la réduction directe.
- **Suite possible** : réductions par axe N-D, parallélisation avec
  accumulateurs par worker.
