# Performances — xarray-go vs xarray (Python)

Comparaison de performance entre **xarray-go** et le **xarray Python** de
référence (au-dessus de NumPy), sur des opérations équivalentes et des tailles
identiques.

> ⚠️ Chiffres **indicatifs** : une seule exécution, machine locale, sujets à
> variabilité. À reproduire pour toute conclusion ferme (voir plus bas).

## Environnement

| | Version |
|---|---|
| Go | 1.26.5 (linux/amd64) |
| Python | 3.13.7 (CPython) |
| xarray | 2026.4.0 |
| NumPy | 2.2.4 |

## Résultats (temps par opération, plus bas = mieux)

| Opération | Taille | xarray-go | xarray (Python) | Rapport | Gagnant |
|-----------|--------|-----------|-----------------|---------|---------|
| `Add` (élément par élément) | 100×100 | **272 µs** | 232 µs | 1,2× | Python |
| `Broadcast` (par nom) | 200×200 | 274 µs | **128 µs** | 2,1× | Python |
| `SumAxis` (réduction) | 100×100 | **14 µs** | 89 µs | 6,3× | **Go** |
| `MeanAxis` (réduction) | 100×100 | **15 µs** | 82 µs | 5,6× | **Go** |
| `GroupBy.Sum` | 1000×10, 10 groupes | **143 µs** | 2017 µs | 14× | **Go** |

## Lecture des résultats

Le résultat est **nuancé** — aucun des deux n'est uniformément meilleur :

- **NumPy gagne sur le calcul élément par élément à grande taille** (`Add`,
  `Broadcast`). Ses boucles sont du C vectorisé (SIMD) ; notre `binaryOp` est une
  boucle Go scalaire. L'écart reste modéré (1,2× à 2,1×) et se resserre depuis
  l'ajout du chemin rapide « dimensions identiques ».

- **xarray-go gagne largement sur les réductions et le `groupby`** (5×–14×). Là,
  le coût de xarray est dominé par l'**overhead Python** : création d'objets,
  dispatch dynamique, et — pour `groupby` — le passage par pandas. Go compile en
  code natif sans cette surcharge.

En résumé :

- **Latence / petites opérations / orchestration** → avantage **Go** (pas
  d'overhead d'interpréteur).
- **Débit brut sur gros tableaux numériques** → avantage **NumPy** (vectorisation).

## Pistes d'amélioration côté Go

- Vectorisation manuelle (déroulage, `//go:` intrinsics) ou usage de SIMD pour
  `binaryOp` sur gros tableaux.
- Réduire les allocations de `Add` : l'alignement sur coordonnées (copies via
  `takeAlong`/`reindex`) domine encore les 135 allocations mesurées.
- Chemin rapide « pas de coordonnées » pour éviter l'alignement quand inutile.

## Reproduire

```bash
# Go
go test -bench='Add$|Broadcast|SumAxis|MeanAxis|GroupBySum' -benchmem -run='^$' ./...

# Python (nécessite xarray + numpy)
python3 bench/xr_bench.py
```

Le script `bench/xr_bench.py` mesure les opérations xarray équivalentes, aux
mêmes tailles que `bench_test.go`, via un calibrage adaptatif du nombre
d'itérations (cible ~0,5 s par mesure, après warmup).

## Équité de la comparaison

- Mêmes tailles, mêmes opérations sémantiques des deux côtés.
- Chaque mesure inclut la construction du résultat (nouvel objet), des deux côtés.
- Aucune parallélisation explicite ; NumPy peut néanmoins utiliser des routines
  vectorisées internes.
- Go mesuré via `testing.B` (inclut les allocations) ; Python via
  `time.perf_counter` après warmup.
