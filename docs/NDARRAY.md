# Paquet `ndarray` — « mini-NumPy » en Go

Le paquet `github.com/bmarty/xarray/ndarray` fournit un **tableau dense
N-dimensionnel de `float64`**, pensé comme moteur de calcul : spécialisé (pas de
générique), sans closure sur le chemin chaud, avec broadcasting **positionnel
façon NumPy** (aligné à droite, dimensions de taille 1 étirées).

## Pourquoi ce paquet ?

La comparaison de performance (voir `BENCHMARKS.md`) a montré que le chemin
générique de `xarray.Variable` paie deux surcoûts : la **closure `func(T,T) T`**
non inlinée (un appel par élément) et l'**overhead d'alignement/coordonnées**.
`ndarray` isole un moteur de calcul propre, sans ces surcoûts.

## Portée (et non-portée)

**Inclus** : construction (`New`, `Zeros`, `Arange`), accès (`At`, `Shape`,
`Data`…), arithmétique (`Add`/`Sub`/`Mul`/`Div`, même forme **et** broadcasting),
scalaires (`AddScalar`, `MulScalar`), réductions (`Sum`, `Mean`, `SumAxis`,
`MeanAxis`).

**Volontairement exclu** (ce n'est PAS un portage complet de NumPy) : algèbre
linéaire (matmul, décompositions — voir gonum), dtypes multiples, vues/slicing
avancé, fonctions universelles complètes, indexation booléenne/fancy, fftpack,
etc. Porter *tout* NumPy (des décennies de C + BLAS/LAPACK + noyaux SIMD) n'est
pas un objectif réaliste. (« SIMD » = *Single Instruction, Multiple Data*.)

## Résultats — et la conclusion honnête

Benchmarks (mêmes tailles que `BENCHMARKS.md`) :

| Opération | `ndarray` (Go nu) | `xarray.Variable` | **NumPy pur** |
|-----------|-------------------|-------------------|---------------|
| `Add` même forme, 100×100 | 19 µs | 19 µs | **4,2 µs** |
| `Add` même forme, 1000×1000 | 1682 µs | 1801 µs | **733 µs** |
| `SumAxis`, 1000×1000 | 1542 µs | — | — |

**Conclusion : même un moteur Go nu (sans closure, sans étiquettes) ne bat pas
NumPy** — il reste 2,3× à 4,5× plus lent. Le gain sur `xarray.Variable` est réel
mais marginal.

Pourquoi, une fois de plus :

1. **NumPy = C + SIMD** : plusieurs `float64` par instruction ; le compilateur Go
   ne vectorise pas (et notre noyau AVX manuel s'est révélé plus lent, cf.
   `BENCHMARKS.md`).
2. **Allocateur** : `make([]float64, n)` zéro-initialise (écriture perdue).
3. **Opérations memory-bound** : à grande taille, on est limité par la bande
   passante mémoire, pas par le calcul — le SIMD n'aide plus, mais NumPy garde
   l'avantage grâce à un allocateur et des boucles plus serrés.

## API in-place : rejoindre NumPy sur le memory-bound

Les benchmarks ont montré que le coût de `Add` sur gros tableau est dominé par
l'**allocation** du résultat (`make` zéro-initialise 8 Mo), pas par le calcul. Les
opérations **in-place** (`AddInto`, `SubInto`, `MulInto`, `DivInto`,
`AddInPlace`) écrivent dans un `dst` fourni, donc **zéro allocation** :

| `Add` 1000×1000 | Temps | Allocations |
|-----------------|-------|-------------|
| `Add` (alloue le résultat) | 1632 µs | 8 Mo, 4 allocs |
| **`AddInto` (in-place)** | **856 µs** | **0 B, 0 alloc** |
| NumPy pur | 733 µs | — |

**Résultat : en réutilisant le buffer de destination, on passe de 2,2× à 1,17× de
NumPy** — soit ~17 % d'écart résiduel (probablement le SIMD de NumPy sur une
boucle memory-bound). L'allocation était bien le vrai goulot, et il se règle en
**Go pur, sans dépendance ni cgo**.

Usage typique (réutiliser `dst` dans une boucle) :

```go
dst := ndarray.Zeros(1000, 1000)
for _, b := range lots {
    _ = ndarray.AddInto(dst, a, b) // aucune allocation par itération
    // ... consommer dst ...
}
```

## Ce que `ndarray` apporte réellement

Pas la victoire sur NumPy (impossible en Go idiomatique), mais :

- un **moteur de calcul propre et testé**, sans dépendance externe ;
- un **socle** sur lequel `xarray.Variable[float64]` pourrait déléguer ses
  opérations pour gagner le facteur closure/alignement ;
- le **broadcasting positionnel NumPy** (complémentaire du broadcasting par nom
  de xarray).

## Si l'objectif est la performance de calcul pure

La voie réaliste n'est pas de réécrire NumPy, mais :

- soit **appeler NumPy/BLAS** depuis Go (cgo) pour les noyaux critiques ;
- soit utiliser **gonum** (`gonum/floats`, `gonum/mat`) qui embarque déjà de
  l'assembleur SIMD et BLAS — au prix d'une dépendance et d'un modèle 2D
  (matrices), pas N-D.

## Exemple

```go
import "github.com/bmarty/xarray/ndarray"

a, _ := ndarray.New([]int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
b, _ := ndarray.New([]int{3}, []float64{10, 20, 30}) // diffusé sur les lignes
c, _ := a.Add(b)         // broadcasting positionnel -> (2,3)
s, _ := a.SumAxis(0)     // somme le long de l'axe 0
```
