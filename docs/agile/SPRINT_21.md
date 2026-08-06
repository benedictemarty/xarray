# Sprint 21 — Broadcast par lignes (itération strided optimisée)

- **Période** : démarrage 2026-08-06.
- **Objectif** : s'attaquer au vrai goulot du broadcast (structure d'itération),
  identifié au Sprint 20.

## Périmètre

| Sujet | État |
|-------|------|
| Broadcast float64 réécrit en itération par lignes | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Équivalence rapide/générique toujours vérifiée.
- [x] `CHANGELOG.md` / `docs/BENCHMARKS.md` à jour ; commit atomique.

## Décisions de conception

- Itération **par lignes** : boucle interne **contiguë** sur le dernier axe ;
  scalaire hissé hors de la boucle quand un stride interne est nul (cas typique).
- Noyaux contigus spécialisés `fillScalarVec`/`fillVecScalar`/`fillVecVec`.
- `parallelLines` : parallélise selon le volume total (et non le nombre de
  lignes, qui peut être petit alors que chaque ligne est longue).

## Résultat

`Broadcast` 1 M : **2650 µs → 1350 µs** (~2×) ; écart avec NumPy réduit de 2,5× à
**1,6×**. Confirme que le coût était la structure d'itération, pas la closure.

## Rétrospective

- **Bien** : le diagnostic du Sprint 20 a mené au bon correctif.
- **Suite** : algèbre linéaire (Sprint 22).
