# Sprint 10 — Volet performance (xarray-go vs xarray Python)

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : mesurer objectivement les performances de xarray-go
  face au xarray Python de référence (US-19).

## Périmètre engagé

| ID    | User story | État |
|-------|------------|------|
| US-19 | Comparaison de performance documentée xarray-go vs xarray Python | ✅ |

## Critères d'acceptation (Definition of Done)

- [x] Harnais de mesure des deux côtés (Go `bench_test.go`, Python `bench/xr_bench.py`).
- [x] Mesures réelles (xarray/NumPy effectivement installés et exécutés).
- [x] Résultats documentés et analysés honnêtement (`docs/BENCHMARKS.md`).
- [x] `go test ./...` : tous les tests passent.
- [x] `CHANGELOG.md`, backlog à jour ; commit atomique.

## Résultats (voir docs/BENCHMARKS.md)

| Opération | xarray-go | xarray Python | Gagnant |
|-----------|-----------|---------------|---------|
| Add 100×100 | 272 µs | 232 µs | Python (1,2×) |
| Broadcast 200×200 | 274 µs | 128 µs | Python (2,1×) |
| SumAxis 100×100 | 14 µs | 89 µs | Go (6,3×) |
| MeanAxis 100×100 | 15 µs | 82 µs | Go (5,6×) |
| GroupBy.Sum | 143 µs | 2017 µs | Go (14×) |

## Décisions / constats

- **Mesure réelle, pas d'estimation** : xarray 2026.4.0 et NumPy 2.2.4 étaient
  disponibles ; les chiffres proviennent d'exécutions effectives, conformément au
  principe « ne pas inventer ».
- **Conclusion nuancée** : Go l'emporte sur l'overhead (réductions, groupby),
  NumPy sur le débit vectorisé (opérations élément par élément à grande taille).
- **Optimisation induite** : ajout d'un chemin rapide dans `binaryOp` pour les
  dimensions identiques, réduisant nettement le coût de `Add`.

## Rétrospective

- **Bien** : comparaison honnête et reproductible ; a directement motivé une
  optimisation utile.
- **À surveiller** : mesures mono-exécution (variabilité) ; pas de vectorisation
  SIMD côté Go — piste principale pour combler l'écart sur les gros tableaux.
- **Suite possible** : vectorisation, réduction des allocations d'alignement,
  extension du comparatif (I/O, tailles multiples).
