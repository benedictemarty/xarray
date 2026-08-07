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

## Implémenter une autre source

Il suffit d'implémenter l'interface `ChunkSource` (Dims/Shape/Coords/NumChunks/
ChunkRows/ChunkData). On pourrait par exemple adosser un `LazyArray` à un store
Zarr en lisant ses chunks à la demande.
