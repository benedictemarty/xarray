# Sprint 38 — Lazy adossé à Zarr (out-of-core sur format standard)

- **Période** : démarrage 2026-08-07.
- **Objectif** : brancher le moteur lazy sur un store Zarr pour un out-of-core
  réel et interopérable (US-39).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-39 | `ChunkZarr` : LazyArray hors-mémoire adossé à Zarr v2 (1D/2D) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : Zarr 2D (chunks non alignés + zlib) et 1D ; Sum streaming + Compute.
- [x] `CHANGELOG.md` / backlog / `docs/LAZY.md` à jour ; commit atomique.

## Décisions de conception

- `zarrRowSource` implémente `ChunkSource`. Pour un bloc de lignes [s,e),
  `readBlock` ne lit que les chunks Zarr qui le recouvrent (via `readChunk`),
  reconstruit le sous-bloc, gère chunks non alignés et compression.
- Limité aux tableaux **1D/2D** (cas dominant des données géo) ; ndim>2 rejeté.
- Réutilisation totale de l'infrastructure Zarr (`.zarray`/`.zattrs`, `readChunk`).

## Rétrospective

- **Bien** : bouclage élégant lazy + Zarr → out-of-core sur un format **standard
  et interopérable** (zarr-python). Empreinte mémoire ~1 bloc.
- **Suite possible** : chunking N-D, réductions par axe en lazy, source Zarr v3.
