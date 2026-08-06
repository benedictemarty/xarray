# experimental/cgokernel — réutiliser du C depuis Go (démonstration)

Réponse mesurée à la question : « peut-on réutiliser le C (comme celui de NumPy)
pour aller aussi vite ? ».

> ⚠️ Paquet **isolé** et **expérimental**. Il utilise **cgo** (nécessite
> `CGO_ENABLED=1` + un compilateur C). Le projet principal (`xarray`, `ndarray`)
> reste 100 % Go pur et ne dépend pas de ce paquet.

## Peut-on réutiliser *le C de NumPy* ?

**Pas directement.** NumPy n'est pas une bibliothèque C autonome : c'est un
**module d'extension CPython**. Son API C dépend de l'interpréteur Python
(objets `PyObject`, comptage de références, GIL). Pour l'appeler, il faudrait
**embarquer un interpréteur Python** — ce qui réintroduit le runtime Python, le
GIL et le coût de marshalling, et annule l'intérêt de Go (binaire unique).

**Mais on peut réutiliser DU C** via **cgo** : écrire notre propre noyau C (ou
linker BLAS/OpenBLAS pour l'algèbre linéaire, comme le fait gonum en mode cgo).
C'est ce que démontre `kernel.go` : une boucle C compilée `-O3 -mavx2`,
auto-vectorisée par gcc.

## Résultats mesurés

| `Add` float64 | cgo (C `-O3 -mavx2`) | Go pur | NumPy pur |
|---------------|----------------------|--------|-----------|
| 10 000 (en cache, compute-bound) | **2127 ns** | 5868 ns | 4200 ns |
| 1 000 000 (RAM, memory-bound) | 867 µs | **789 µs** | 733 µs |

## Deux enseignements

1. **Compute-bound (données en cache)** : le C+AVX **bat Go pur (2,8×) ET NumPy**
   (2127 vs 4200 ns). Là, la vectorisation SIMD paie, et gcc la génère
   automatiquement. → Oui, réutiliser du C aide réellement sur ce régime.

2. **Memory-bound (gros tableaux)** : tout le monde converge vers la bande
   passante RAM (733–867 µs). Le SIMD n'aide plus ; **cgo est même légèrement plus
   lent que Go pur** à cause de l'overhead de franchissement de frontière. Le Go
   pur (789 µs) est déjà à ~8 % de NumPy (733 µs) !

## L'insight le plus important

Dans notre bibliothèque, `ndarray.Add` d'un tableau 1 M coûtait **1682 µs**, alors
que le **noyau nu** (buffers pré-alloués) ne coûte que **789 µs**. La différence
n'est PAS le calcul : c'est l'**allocation** du résultat (`make([]float64, n)`
zéro-initialise 8 Mo à chaque opération).

**Conséquence** : pour se rapprocher de NumPy, la priorité n'est ni le SIMD ni le
cgo, mais **réduire les allocations** — une API *in-place* (`AddInto(dst, a, b)`)
et/ou du *buffer pooling*. C'est actionnable en Go pur, sans dépendance.

## Le coût de cgo (pourquoi ce n'est pas gratuit)

- Overhead par appel (~dizaines de ns) : pénalisant pour beaucoup de petits appels.
- Perte du « pur Go » : compilateur C requis, cross-compilation compliquée, build
  plus lourd, plus de binaire statique unique trivial.
- Sécurité mémoire à la frontière Go/C à gérer.

## Reproduire

```bash
CGO_ENABLED=1 go test -bench=Add ./experimental/cgokernel/
```
