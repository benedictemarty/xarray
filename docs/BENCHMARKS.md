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

Trois références, médianes (Go : `-count=6` ; Python : calibrage adaptatif) :

| Opération | Taille | xarray-go | **NumPy pur** | xarray |
|-----------|--------|-----------|---------------|--------|
| `Add` (aligné) | 100×100 | 19 µs | **4,2 µs** | 225 µs |
| `Add` (aligné) grand | 1000×1000 | 1801 µs | **733 µs** | 1088 µs |
| `Broadcast` | 200×200 | 184 µs | **34 µs** | 125 µs |
| `Broadcast` grand | 1000×1000 | 2650 µs | **855 µs** | 1000 µs |
| `SumAxis` | 100×100 | 15 µs | **3,2 µs** | 76 µs |
| `MeanAxis` | 100×100 | 15 µs | **4,7 µs** | 73 µs |
| `GroupBy.Sum` | contigu | 173 µs | **0,6 µs**¹ | 732 µs |

¹ `np.add.reduceat` sur des groupes déjà contigus — cas trivial, pas un `groupby`
général ; comparaison indicative seulement.

### Bat-on NumPy ? (réponse honnête)

**Non.** Face au **moteur de calcul réel (NumPy pur, C + SIMD), xarray-go est plus
lent partout** — d'un facteur ~2× à ~5× sur l'arithmétique et les réductions.

Ce qui est vrai, c'est que **xarray-go est plus rapide que la _couche_ xarray**
(qui empile un lourd overhead d'objets Python sur NumPy). Autrement dit :

```
NumPy pur   <   xarray-go   <   xarray
(le plus rapide)            (le plus lent)
```

xarray-go **élimine l'overhead Python de xarray**, mais **n'atteint pas le moteur
NumPy**. Les gros ratios « ×13 » d'une version précédente de ce document
comparaient Go à xarray et non à NumPy : c'était **trompeur** et a été corrigé.

Pourquoi NumPy pur reste devant :

1. **SIMD** — noyaux C traitant 4–8 `float64` par instruction ;
2. **Pas d'auto-vectorisation** du compilateur Go (voir plus bas) ;
3. **Allocateur Go** — `make([]float64, n)` **zéro-initialise** le résultat (une
   écriture de 8 Mo « perdue » avant remplissage ; NumPy alloue sans zéro).

### Où se situe réellement l'intérêt de xarray-go

Pas dans le débit de calcul brut (terrain de NumPy), mais dans : le **typage
statique**, l'**absence de runtime Python**, la **latence** (pas d'overhead
d'interpréteur), et le déploiement (binaire unique). Pour rivaliser sur le
**calcul**, il faudrait un moteur de tableaux dense de niveau NumPy — d'où le
chantier « porter NumPy en Go » (voir `docs/NDARRAY.md`).

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
- **Sprint 19** : le noyau direct float64 (sans closure) est branché sur les
  **quatre** opérations. `Mul`/`Sub`/`Div` 1 M passent de ~3218 µs (closure
  générique) à ~1545 µs (2,1×), comme `Add`.
- **Sprint 20** : spécialisation float64 du **broadcasting** (`broadcastFloat64`,
  switch au lieu de closure). **Gain marginal** (~10 % sur 1 M, nul sur 40 k) —
  contrairement au cas même-forme. Enseignement honnête : pour le broadcast, le
  coût dominant n'est pas la closure mais l'**itération strided** et les **accès
  mémoire non contigus** (un opérande a un stride 0). La closure n'était pas le
  goulot. Le refactoring associé (`broadcastLayout`, `parallelFill`) est en
  revanche une simplification utile, conservée.

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
