# Sprint 13 — Calcul vectoriel et vérification d'équivalence

- **Période** : démarrage 2026-08-06.
- **Objectif** : réduire l'overhead par élément du calcul et garantir la justesse
  vis-à-vis de xarray.

## Périmètre

| Sujet | État |
|-------|------|
| Itération incrémentale de `binaryOp` (flatA/flatB par pas) | ✅ |
| Parallélisation multi-cœurs au-delà de 32 768 éléments | ✅ |
| Test d'équivalence des résultats Go vs xarray | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] `bench/crosscheck.py` + `TestEquivalenceAvecXarray` (tolérance 1e-9).
- [x] `CHANGELOG.md` / `docs/BENCHMARKS.md` à jour ; commit atomique.

## Décisions de conception

- `flatA`/`flatB` maintenus par pas (O(1) amorti) plutôt que recalculés.
- Parallélisation par plages disjointes de `out.data` (aucune course de données).
- Équivalence vérifiée pour add, broadcast, réductions, jointure externe, groupby.

## Résultat

`Broadcast` 40 k : 284 µs → 160 µs. Résultats **identiques** à xarray sur toutes
les opérations testées.

## Rétrospective

- **Bien** : l'équivalence sécurise toutes les optimisations suivantes.
- **Suite** : expérience SIMD (Sprint 14).
