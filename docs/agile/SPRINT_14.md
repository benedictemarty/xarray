# Sprint 14 — SIMD (assembleur) et chemin direct float64

- **Période** : démarrage 2026-08-06.
- **Objectif** : explorer le SIMD manuel et supprimer l'overhead de la closure
  générique pour `Add`.

## Périmètre

| Sujet | État |
|-------|------|
| Noyau AVX (`simd_amd64.s`) + détection CPUID/XGETBV | ✅ (expérience) |
| Chemin direct `Add` float64 sans closure | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe (asm + `noasm`).
- [x] Résultats mesurés et documentés ; cœur reste pur-Go.
- [x] `CHANGELOG.md` / `docs/BENCHMARKS.md` à jour ; commit atomique.

## Décisions et constats

- **SIMD manuel : conclusion négative.** Le noyau AVX déroulé ×4 (30 Go/s) est
  **plus lent** que la boucle Go (77 Go/s) sur cette opération memory-bound → non
  branché sur `Add`. Conservé comme expérience documentée.
- **La closure était le vrai coût** : `Add` sans closure passe de 3,58 ms à
  1,63 ms sur 1 M (100×100 : 48 → 18 µs).

## Rétrospective

- **Bien** : démarche honnête (ne pas expédier un chemin plus lent).
- **Suite** : moteur `ndarray` (Sprint 15).
