# Sprint 19 — Noyaux directs float64 sur toute l'arithmétique

- **Période** : démarrage 2026-08-06.
- **Objectif** : propager le chemin rapide float64 (sans closure) d'`Add` aux
  autres opérations.

## Périmètre

| Sujet | État |
|-------|------|
| Chemin rapide mutualisé branché sur `Sub`/`Mul`/`Div` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Équivalence rapide/générique testée pour les 4 opérations.
- [x] `CHANGELOG.md` / `docs/BENCHMARKS.md` à jour ; commit atomique.

## Décisions de conception

- `binaryFloat64Fast(kernel)` mutualise l'alignement et la reconstruction des
  coordonnées ; noyaux `subFloat64`/`mulFloat64`/`divFloat64`.

## Résultat

`Mul`/`Sub`/`Div` 1 M mêmes coords : **~3218 µs (closure) → ~1545 µs** (2,1×).

## Rétrospective

- **Suite** : spécialiser le broadcasting float64 (Sprint 20).
