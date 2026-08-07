# Évaluation paresseuse par chunks (esprit dask)

`xarray-go` fournit un moteur d'**évaluation paresseuse** (`LazyArray`) inspiré de
dask : les opérations sont différées et le calcul se fait **bloc par bloc**, ce
qui permet la **parallélisation** et le traitement de données **plus grandes que
la RAM**.

## Principe

```go
lz, _ := xarray.Chunk(da, 1000)          // découpe en blocs de 1000 lignes
res, _ := lz.MulScalar(2).AddScalar(1).  // opérations DIFFÉRÉES (rien calculé)
    Compute()                            // exécution parallèle par chunk
```

- **`Chunk(da, size)`** / **`ChunkFile(...)`** : crée un `LazyArray` adossé à une
  source (mémoire ou fichier), découpé en blocs de `size` lignes le long de
  l'axe 0.
- **`Map`/`AddScalar`/`MulScalar`** : empilent des transformations (graphe
  différé), sans calcul.
- **`Compute()`** : matérialise en `DataArray[float64]`, en exécutant les chunks
  **en parallèle** (goroutines).
- **`Sum`/`Mean`/`Min`/`Max`** : réductions **en streaming** — un chunk est lu,
  agrégé, puis libéré. On ne charge jamais tout le tableau.

## Hors-mémoire (out-of-core) réel

`ChunkFile` lit chaque bloc à la demande depuis un fichier binaire (seek + read) :

```go
_ = xarray.WriteRawF64("big.f64", data)  // float64 little-endian, ordre C
lz, _ := xarray.ChunkFile("big.f64", []string{"t","x"}, []int{1_000_000, 50}, nil, 10_000)
total, _ := lz.Sum()   // ne charge que 10 000 lignes à la fois -> RAM bornée
```

Un tableau de plusieurs Go peut ainsi être agrégé avec une empreinte mémoire de
l'ordre d'un seul chunk.

## Portée (MVP) et limites

**Géré** : chunking 1D le long de l'axe 0 ; sources mémoire et fichier binaire ;
transformations élément par élément ; réductions globales en streaming ;
`Compute` parallèle.

**Non géré** (par rapport à dask) :
- chunking **multi-dimensionnel** et *rechunk* ;
- graphe de calcul entre **plusieurs** tableaux (ex. `a_lazy + b_lazy`) ;
- réductions **par axe** en lazy ;
- planification/optimisation de graphe, spilling disque, cluster distribué ;
- types autres que `float64`.

C'est une démonstration fidèle du **modèle** (différé + chunké + streaming +
parallèle + out-of-core), pas un portage de dask.

## Combiner deux LazyArray (graphe multi-tableaux)

`Add`/`Sub`/`Mul`/`Div` combinent **deux** LazyArray élément par élément, chunk
par chunk (différé). Les deux doivent avoir la même forme et le même découpage.

```go
la, _ := xarray.ChunkFile("a.f64", dims, shape, nil, 10_000)
lb, _ := xarray.ChunkFile("b.f64", dims, shape, nil, 10_000)
res, _ := la.MulScalar(2).Sub(lb)   // graphe : (2·a − b), toujours différé
total, _ := res.Sum()               // agrège sans charger a ni b entièrement
```

Deux gros tableaux (fichier ou Zarr) sont ainsi combinés avec une empreinte
mémoire de l'ordre d'un bloc de chacun.

## Source Zarr (out-of-core sur un format standard)

`ChunkZarr` adosse un `LazyArray` à un **store Zarr v2** (tableau 1D/2D, `<f8`) :
chaque bloc de lignes est reconstruit en ne lisant que les **chunks Zarr** qui le
recouvrent — un tableau Zarr de plusieurs Go peut donc être agrégé avec une
empreinte mémoire de l'ordre d'un bloc.

```go
_ = xarray.WriteDataArrayZarr("data.zarr", da, []int{1000, 50}, xarray.ZarrZlib)
lz, _ := xarray.ChunkZarr("data.zarr", 1000)   // blocs lazy de 1000 lignes
moy, _ := lz.Mean()                            // streaming, ne charge que ~1 bloc
```

Comme le store est du Zarr v2 standard, il est **interopérable** (écrit/lu aussi
par zarr-python — cf. `docs/ZARR.md`).

## Implémenter une autre source

Il suffit d'implémenter l'interface `ChunkSource` (Dims/Shape/Coords/NumChunks/
ChunkRows/ChunkData) — comme le font `memSource`, `fileSource` et `zarrRowSource`.
