# Prise en charge de Zarr (v2 et v3)

`xarray-go` **lit et écrit** le format Zarr (tableaux N-D chunkés et compressés,
stockés en arborescence de fichiers), en **v2 comme en v3**.

Particularité importante : contrairement au netCDF (auto-cohérent seulement),
**chaque brique Zarr est validée contre l'implémentation de référence
`zarr-python` 3.3.0** — pas seulement en Go↔Go. Le format Blosc a notamment été
implémenté en **inspectant les octets réels** produits par zarr-python, chaque
hypothèse étant vérifiée par aller-retour (voir « Méthode de validation »).

## API

```go
// --- Lecture (détection automatique v2 / v3) ---
da, _  := xarray.ReadDataArrayZarr("array.zarr")
ds, _  := xarray.ReadDatasetZarr("group.zarr")

// --- Écriture Zarr v2 ---
_ = xarray.WriteDataArrayZarr("a.zarr", da, []int{2, 3}, xarray.ZarrZstd) // chunks explicites
_ = xarray.WriteDatasetZarr("g.zarr", ds, xarray.ZarrZstd)                // un chunk par array
_ = xarray.WriteDatasetZarrChunked("g.zarr", ds,
        map[string]int{"time": 100, "x": 50}, xarray.ZarrZstd)            // découpage par dimension

// --- Écriture Zarr v3 ---
_ = xarray.WriteDataArrayZarrV3("a.zarr", da, []int{2, 3}, xarray.ZarrZstd)
_ = xarray.WriteDatasetZarrV3("g.zarr", ds, xarray.ZarrZstd)
_ = xarray.WriteDatasetZarrV3Chunked("g.zarr", ds,
        map[string]int{"time": 100}, xarray.ZarrZstd)

// --- Lecture paresseuse (out-of-core), un chunk à la fois ---
lazy, _ := xarray.ChunkZarr("big.zarr", 1000)
```

Compressions à l'écriture : `ZarrNone`, `ZarrZlib` (gzip en v3), `ZarrZstd`.

## Matrice de capacités

|  | **Lecture** | **Écriture** |
|---|---|---|
| **Zarr v2** | ✅ | ✅ |
| **Zarr v3** | ✅ (`zarr.json`, clés `c/0/0`, `dimension_names`) | ✅ |
| **Compression** | aucune, zlib, gzip, **Blosc** (LZ4 / zstd), zstd | aucune, zlib/gzip, zstd |
| **Filtres Blosc** | byte-shuffle, bitshuffle¹ | — |
| **Chunks multiples** | ✅ | ✅ (découpage configurable par dimension) |
| **dtypes** | `f8/f4`, `i8/i4/i2/i1`, `u*` (+ boutisme) → float64 | float64 (`<f8`) |
| **`fill_value`** | nombre, `null`, `"NaN"`/`"Infinity"` | `null` (v2) / `"NaN"` (v3) |
| **Métadonnées consolidées** | tolérées | `.zmetadata` écrit (v2) |
| **Groupes (Dataset)** | ✅ (coordonnées réattachées) | ✅ |

¹ bitshuffle géré quand le nombre d'éléments par bloc est multiple de 8 (cas des
blocs pleins) ; sinon erreur explicite (voir limites).

## Détails notables

- **Décodeur Blosc pur Go** (`zarr_blosc.go`) : conteneur v1, blocs, découpage en
  sous-flux (`BLOSC_MIN_BUFFERSIZE=128`, seuls les blocs pleins), sous-flux non
  compressés, byte-/bit-unshuffle. Codec `zstd` via
  `github.com/klauspost/compress` (seule dépendance externe du module).
- **Filtre appliqué par bloc** : le shuffle et le découpage se font bloc par bloc,
  ce qui est indispensable pour les grands tableaux multi-blocs.
- **`fill_value`** : à l'écriture v2, `null` (et non `0`) — un `fill_value`
  numérique est interprété par xarray comme `_FillValue` et masquerait en NaN
  toutes les valeurs égales (les `0` légitimes).

## Limites (assumées, documentées)

- **bitshuffle** avec un nombre d'éléments par bloc **non multiple de 8** :
  refusé par une erreur explicite (agencement spécifique de la bibliothèque
  *bitshuffle* non reproduit de façon certaine — on préfère échouer que renvoyer
  des données fausses).
- Codec **`crc32c`** (Zarr v3) et encodages de clé de chunk non standard : non
  gérés.
- Écriture : données **float64** uniquement, ordre C ; **Blosc** non écrit
  (l'écriture compressée passe par zstd/gzip, relus par zarr-python).

## Méthode de validation

Les tests hermétiques du paquet embarquent des **fixtures réelles** produites par
zarr-python (`testdata/zarr_blosc_*`, `zarr_int_dtypes`, `zarr_v3_dataset`), et les
tests d'écriture sont relus par zarr-python. Exemples couverts : Blosc/LZ4
(memcpy, découpé, sous-flux brut), multi-blocs (1 M valeurs), bitshuffle, dtypes
entiers (coords int64), v3 (zstd/blosc), écriture zstd/gzip/v3, découpage
configurable. Voir `zarr_test.go`, `zarr_dataset_test.go`.
