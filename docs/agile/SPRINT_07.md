# Sprint 7 — Benchmarks / Performances

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : mesurer les performances et supprimer les
  inefficacités évidentes (dette **T-03**).

## Périmètre engagé

| ID   | Sujet | État |
|------|-------|------|
| T-03 | Benchmarks + optimisations sûres | ✅ |

## Critères d'acceptation (Definition of Done)

- [x] Suite de benchmarks Go (`bench_test.go`).
- [x] Au moins une optimisation mesurée sans régression fonctionnelle.
- [x] `go test ./...` : tous les tests passent.
- [x] `CHANGELOG.md`, backlog à jour.
- [x] Commit atomique.

## Optimisations réalisées

1. **`binaryOp` — pré-calcul des strides.** La boucle interne (exécutée à chaque
   élément) faisait deux accès à une `map[string]int` par dimension. Les strides
   sont désormais pré-calculés dans des slices indexés par position (0 pour une
   dimension absente, ce qui la neutralise). Suppression totale des accès map en
   boucle chaude.
2. **`cloneVar` — clonage sans revalidation.** Les chemins internes clonaient via
   `NewVariable`, qui recopie (après que `Data()`/`Dims()`/`Shape()` ont déjà
   copié) **et** revalide unicité des dimensions et cohérence des tailles. Le
   nouveau `Variable.cloneVar` fait une unique copie profonde sans validation.

## Résultats (indicatifs, machine locale)

| Benchmark            | Avant     | Après    | Gain   |
|----------------------|-----------|----------|--------|
| `Add` (100×100)      | 561 µs    | 255 µs   | ×2,2   |
| `Broadcast` (200×200)| 1435 µs   | 230 µs   | ×6,2   |
| `OuterJoin`          | 629 µs    | 309 µs   | ×2,0   |

## Rétrospective

- **Bien** : gains importants pour un risque faible (aucune régression de test) ;
  le pré-calcul de strides est un patron réutilisable (`take`, `reduceAxis`
  pourraient en bénéficier).
- **À surveiller** : l'immutabilité par copie reste coûteuse en allocations
  (`BenchmarkClone` ~84 Ko/op) ; une API « vue » (sans copie) pourrait être
  étudiée ultérieurement.
- **Prochain sprint** : netCDF (US-16).
