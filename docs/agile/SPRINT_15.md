# Sprint 15 — Moteur `ndarray` (mini-NumPy)

- **Période** : démarrage 2026-08-06.
- **Objectif** : un moteur de calcul dense N-D float64 (US-24), sans closure ni
  overhead d'étiquettes.

## Périmètre

| ID | Sujet | État |
|----|-------|------|
| US-24 | Paquet `ndarray` (dense N-D float64) | ✅ (cœur) |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./ndarray/` passe.
- [x] Benchmarks vs NumPy pur ; conclusion documentée.
- [x] `CHANGELOG.md` / `docs/NDARRAY.md` à jour ; commit atomique.

## Décisions de conception

- Broadcasting **positionnel façon NumPy** (aligné à droite), complémentaire du
  broadcasting par nom de xarray.
- Ops/réductions spécialisées float64.

## Constat honnête

Même un moteur Go nu reste **2,3×–4,5× plus lent que NumPy pur**. Rattraper NumPy
exige C+SIMD+BLAS. Valeur du paquet : architecturale (socle propre), pas de
supériorité de débit.

## Rétrospective

- **Suite** : API in-place pour supprimer le coût d'allocation (Sprint 16).
