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
| `Add` (aligné) | 100×100 | **18 µs** | 232 µs | 13× | **Go** |
| `Add` (aligné) grand | 1000×1000 | 1627 µs | **1172 µs** | 1,4× | Python |
| `Broadcast` (par nom) | 200×200 | 160 µs | **120 µs** | 1,3× | Python |
| `Broadcast` grand | 1000×1000 | 2452 µs | **970 µs** | 2,5× | Python |
| `SumAxis` (réduction) | 100×100 | **14 µs** | 75 µs | 5,5× | **Go** |
| `MeanAxis` (réduction) | 100×100 | **15 µs** | 69 µs | 4,7× | **Go** |
| `GroupBy.Sum` | 1000×10, 10 groupes | **137 µs** | 1671 µs | 12× | **Go** |

### Bat-on NumPy ? (réponse directe)

**Cela dépend du régime :**

- **Petites et moyennes tailles, réductions, `groupby`, arithmétique alignée** →
  **oui, xarray-go gagne**, souvent d'un ordre de grandeur (4×–13×). L'overhead
  Python/pandas de xarray y est écrasant.
- **Débit brut sur des millions d'éléments (`Add`/`Broadcast` grands)** → **non,
  NumPy reste devant** (1,4×–2,5×). Trois raisons structurelles :
  1. **SIMD** : les noyaux C de NumPy traitent 4–8 `float64` par instruction ;
  2. **Allocateur Go** : `make([]float64, n)` **zéro-initialise** le résultat, soit
     une écriture de 8 Mo « perdue » avant de le remplir (NumPy alloue sans
     zéro) ;
  3. **Pas d'auto-vectorisation** du compilateur Go (voir ci-dessous).

En clair : xarray-go gagne sur la **majorité des opérations réalistes**, mais pas
sur le pur *streaming* numérique à grande échelle, qui reste le terrain de NumPy.

## Lecture des résultats

xarray-go gagne désormais **4 opérations sur 5** :

- **xarray-go gagne sur les réductions, le `groupby` et l'arithmétique alignée**
  (4,6×–12×). Le coût de xarray y est dominé par l'**overhead Python** (création
  d'objets, dispatch dynamique, passage par pandas pour `groupby`). Pour `Add`,
  l'optimisation de l'alignement (voir Sprint 11 : les coordonnées identiques ne
  sont plus recopiées) a fait passer l'opération de 272 µs à 48 µs.

- **NumPy garde l'avantage sur `Broadcast`** — le seul cas restant. C'est du
  calcul élément par élément pur, **sans coordonnées** (rien à optimiser côté
  alignement). Après itération incrémentale (O(1) amorti par élément) et
  parallélisation multi-cœurs (Sprint 13), l'écart est tombé de 2,7× à **1,3×**
  sur 40 000 éléments, mais NumPy reste devant (**2,5×** sur 1 M) : ses boucles C
  **vectorisées SIMD** traitent plusieurs valeurs par instruction, là où Go
  exécute une closure scalaire par élément. C'est un écart structurel (voir plus
  bas « Pourquoi pas de SIMD en Go »).

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

## Pourquoi Go n'a pas d'auto-vectorisation SIMD (comme le C de NumPy) ?

NumPy est rapide parce que ses noyaux sont écrits en **C/Cython compilés avec des
compilateurs très optimisants** (GCC/Clang, `-O3`), qui font de
l'**auto-vectorisation** : ils transforment une boucle scalaire en instructions
**SIMD** (AVX2/AVX-512) traitant 4, 8 ou 16 `float64` par instruction. NumPy
fournit aussi des noyaux SIMD écrits à la main pour les opérations courantes.

Le compilateur Go, lui, **ne fait quasiment pas d'auto-vectorisation**. Raisons
principales :

1. **Priorité à la compilation rapide et à la simplicité.** Le compilateur Go
   privilégie des temps de compilation très courts et un code prévisible. Les
   passes d'auto-vectorisation (analyse de dépendances, déroulage, coût/bénéfice)
   sont lourdes et complexes — à l'opposé de la philosophie du compilateur.
2. **Pas (encore) de back-end LLVM.** GCC et Clang ont des décennies
   d'optimisation vectorielle ; le back-end natif de Go (`gc`) est plus jeune et
   n'a pas ces passes. (Le projet gccgo pourrait en profiter, mais n'est pas le
   compilateur de référence.)
3. **Modèle mémoire et sécurité.** Les vérifications de bornes et le style
   idiomatique Go rendent l'auto-vectorisation plus difficile à déclencher.
4. **Overhead de la généricité.** Notre `binaryOp` reçoit l'opération sous forme
   de **closure `func(T,T) T`** : un appel indirect par élément, que le
   compilateur n'inline pas à travers le générique — ce qui bloque toute
   vectorisation même si elle existait.

Ce qui reste possible côté Go (par ordre d'effort) :

- **Réduire l'overhead par élément** (itération incrémentale) — ✅ fait.
- **Paralléliser sur les cœurs** (goroutines) — ✅ fait, mais le SIMD mono-cœur de
  NumPy reste compétitif.
- **Spécialiser les opérations** sans closure (boucles `+`/`*` dédiées) pour aider
  le compilateur à mieux optimiser — piste non faite.
- **Écrire de l'assembleur SIMD** (fichiers `.s`, AVX2) pour les noyaux critiques,
  comme le font certaines bibliothèques Go (gonum). Gain maximal mais coût de
  maintenance élevé et dépendant de l'architecture.

En clair : l'écart sur le pur débit vectoriel n'est pas un défaut de notre code
mais une différence d'outillage (compilateur + noyaux SIMD). Il se comble avec de
l'assembleur, au prix de la simplicité.

### Expérience : un noyau AVX écrit à la main (Sprint 14)

Nous avons **effectivement écrit** un noyau SIMD en assembleur Plan 9
(`simd_amd64.s`) : addition de `float64` avec `VADDPD` sur registres YMM (4
`float64` par instruction), déroulé ×4, avec détection AVX au runtime
(`CPUID`/`XGETBV`) et repli pur-Go. Résultat mesuré sur le **noyau isolé** (8192
`float64`, en cache) :

| Noyau `addFloat64` | Temps | Débit (SetBytes) |
|--------------------|-------|------------------|
| **Boucle Go idiomatique** | **856 ns** | **77 Go/s** |
| AVX manuel (déroulé ×4) | 2199 ns | 30 Go/s |

**Conclusion contre-intuitive mais mesurée : le code généré par le compilateur Go
bat notre assembleur AVX** (2,6×). La boucle scalaire Go atteint la bande
passante du cache L1/L2 (~220 Go/s de trafic réel), alors que notre noyau naïf ne
l'atteint pas — l'opération est **memory-bound**, régime où le SIMD n'aide pas.

Décision d'ingénierie : **ne pas router `Add` vers ce noyau** (ce serait plus
lent). Le code assembleur est conservé, testé et *benchmarké* comme expérience
documentée. Battre le compilateur exigerait le niveau d'optimisation de gonum
(alignement, FMA, *streaming stores*, *prefetch*) — et surtout des opérations
**compute-bound** (ex. `exp`, `sqrt`, produits combinés), pas une simple addition
limitée par la mémoire.

Reproduire : `go test -bench=AddFloat64Kernel ./...` (AVX) vs
`go test -tags noasm -bench=AddFloat64Kernel ./...` (pur Go).

## Vérification d'équivalence des résultats

Les performances ne valent rien si les résultats diffèrent. Un test croisé
vérifie que **xarray-go et xarray produisent des valeurs identiques** :

- `bench/crosscheck.py` calcule plusieurs opérations avec xarray et écrit les
  résultats dans `bench/expected.json` ;
- `crosscheck_test.go` recalcule les mêmes opérations en Go et compare à ce
  fichier (tolérance `1e-9`).

```bash
python3 bench/crosscheck.py      # génère bench/expected.json
go test -run Equivalence ./...   # vérifie l'égalité Go vs xarray
```

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
