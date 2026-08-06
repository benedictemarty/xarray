# Sprint 22 — Algèbre linéaire `ndarray`

- **Période** : démarrage 2026-08-06.
- **Objectif** : fournir le produit matriciel, brique de base du ML.

## Périmètre

| Sujet | État |
|-------|------|
| `Matmul` (2D), `MatVec`, `T` (transposée 2D) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./ndarray/` passe.
- [x] Correctness testée + benchmark ; perf comparée à NumPy/BLAS.
- [x] `CHANGELOG.md` / `docs/NDARRAY.md` à jour ; commit atomique.

## Décisions de conception

- `Matmul` en ordre de boucle **ikj** (lignes de `b` et `c` parcourues de façon
  contiguë), l'ordre le plus rapide en Go pur ; saut des `aip == 0`.

## Constat honnête

`Matmul` naïf 256×256 : **~11,9 ms** contre **~4,7 ms** pour NumPy/BLAS (~2,5×).
Un BLAS (blocking + SIMD + multithread) serait requis pour de grosses matrices —
via cgo/gonum, au prix du binaire pur-Go.

## Rétrospective

- **Suite** : ML classique sur cette base (Sprint 23).
