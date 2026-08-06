# Sprint 20 — Broadcasting float64 spécialisé (switch)

- **Période** : démarrage 2026-08-06.
- **Objectif** : éviter la closure sur le chemin de broadcasting float64.

## Périmètre

| Sujet | État |
|-------|------|
| `broadcastFloat64` (switch au lieu de closure) | ✅ |
| Refactorisation `broadcastLayout` / `parallelFill` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Équivalence testée (dont repli broadcast).
- [x] `CHANGELOG.md` / `docs/BENCHMARKS.md` à jour ; commit atomique.

## Constat honnête

**Gain marginal** (~10 % sur 1 M, nul sur 40 k) — contrairement au cas même-forme.
Pour le broadcast, le goulot n'est **pas** la closure mais l'**itération strided**
et les **accès mémoire non contigus** (un opérande a un stride 0). La leçon oriente
le Sprint 21.

## Rétrospective

- **Bien** : refactorisation (`broadcastLayout`, `parallelFill`) conservée, utile.
- **À revoir** : la structure d'itération, vrai levier (Sprint 21).
