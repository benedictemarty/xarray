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
| `Add` (élément par élément) | 100×100 | **48 µs** | 222 µs | 4,6× | **Go** |
| `Broadcast` (par nom) | 200×200 | 284 µs | **106 µs** | 2,7× | Python |
| `SumAxis` (réduction) | 100×100 | **14 µs** | 75 µs | 5,5× | **Go** |
| `MeanAxis` (réduction) | 100×100 | **15 µs** | 69 µs | 4,7× | **Go** |
| `GroupBy.Sum` | 1000×10, 10 groupes | **137 µs** | 1671 µs | 12× | **Go** |

## Lecture des résultats

xarray-go gagne désormais **4 opérations sur 5** :

- **xarray-go gagne sur les réductions, le `groupby` et l'arithmétique alignée**
  (4,6×–12×). Le coût de xarray y est dominé par l'**overhead Python** (création
  d'objets, dispatch dynamique, passage par pandas pour `groupby`). Pour `Add`,
  l'optimisation de l'alignement (voir Sprint 11 : les coordonnées identiques ne
  sont plus recopiées) a fait passer l'opération de 272 µs à 48 µs.

- **NumPy garde l'avantage sur `Broadcast`** — le seul cas restant. C'est du
  calcul élément par élément pur, à grande taille (40 000 éléments), **sans
  coordonnées** (donc rien à optimiser côté alignement) : les boucles C
  vectorisées (SIMD) de NumPy battent notre boucle Go scalaire. C'est l'écart de
  débit brut vectoriel, structurel.

En résumé :

- **Latence, orchestration, réductions, arithmétique alignée** → avantage **Go**.
- **Débit vectoriel brut sur gros tableaux sans coordonnées** → avantage **NumPy**.

## Pistes d'amélioration côté Go

- **`Broadcast`** reste le point faible : vectorisation manuelle (déroulage,
  SIMD) de `binaryOp` pour rapprocher NumPy sur le calcul élément par élément à
  grande taille.
- Réduire encore les allocations de `Add` (18 restantes, essentiellement le
  clonage des coordonnées du résultat).

## Historique

- **Sprint 10** : mise en place du comparatif ; chemin rapide « dimensions
  identiques » dans `binaryOp` (`Add` 348 µs → 250 µs).
- **Sprint 11** : l'alignement ne recopie plus les coordonnées déjà identiques
  (`Add` 272 µs → 48 µs, 135 → 18 allocations) — Go passe devant NumPy sur `Add`.

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
