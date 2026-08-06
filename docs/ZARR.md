# Prise en charge de Zarr (v2)

`xarray-go` lit et écrit un **sous-ensemble du format Zarr v2** (stockage de
tableaux N-D chunkés et compressés), sur système de fichiers.

Contrairement au netCDF (auto-cohérent seulement), l'**interopérabilité avec
zarr-python a été vérifiée dans les deux sens** (voir plus bas).

## API

```go
// Écriture : chunks par dimension + compression (ZarrNone ou ZarrZlib)
_ = xarray.WriteDataArrayZarr("data.zarr", da, []int{2, 3}, xarray.ZarrZlib)

// Lecture
da, _ := xarray.ReadDataArrayZarr("data.zarr")

// Dataset entier = groupe Zarr (.zgroup + un array par variable/coordonnée)
_ = xarray.WriteDatasetZarr("ds.zarr", ds, xarray.ZarrZlib)
ds2, _ := xarray.ReadDatasetZarr("ds.zarr")
```

## Périmètre

**Géré** :
- Zarr **v2**, store = répertoire du système de fichiers ;
- `DataArray[float64]`, dtype `<f8` (float64 little-endian), ordre C ;
- **chunking** (les chunks de bord sont complétés par `fill_value`) ;
- compression **aucune** ou **zlib** (bibliothèque standard Go) ;
- dimensions, nom et coordonnées dans `.zattrs` (`_ARRAY_DIMENSIONS` suit la
  convention xarray ; les chunks manquants valent `fill_value`).

- **`Dataset` = groupe Zarr** : `WriteDatasetZarr`/`ReadDatasetZarr` écrivent un
  `.zgroup` et un sous-array par variable et par coordonnée (chaque array en un
  seul chunk). Les coordonnées sont des arrays 1D nommés comme leur dimension.

**Non géré** : Zarr v3, dtypes ≠ float64, ordre Fortran, compresseurs
blosc/zstd/lz4, filtres, chunking par variable dans un groupe (un seul chunk par
array).

## Interopérabilité vérifiée

Testée avec **zarr-python 3.3.0 / numpy 2.5.1** :

- **Go → Python** : un store écrit par `WriteDataArrayZarr` (forme 5×4, chunks 2×3
  non alignés, compression zlib) est relu par `zarr.open` — données **identiques**.
- **Python → Go** : un store créé par `zarr.create(..., zarr_format=2,
  compressor=None)` est relu par `ReadDataArrayZarr` — données **identiques**.
- **Groupe (Dataset) Go → Python** : un groupe écrit par `WriteDatasetZarr` est
  ouvert par `zarr.open_group` ; les arrays (`temperature`, `pluie`) et les
  coordonnées (`temps`, `lieu`) sont lus **identiques**.

Utilitaires de démonstration : `cmd/genzarr` (array), `cmd/genzarrds` (groupe),
`cmd/readzarr` (Go lit).

```bash
go run ./cmd/genzarr /tmp/interop.zarr     # Go écrit un store v2
python -c "import zarr; print(zarr.open('/tmp/interop.zarr')[:])"  # Python le lit
```

## Format produit (`.zarray`)

```json
{
  "zarr_format": 2,
  "shape": [5, 4],
  "chunks": [2, 3],
  "dtype": "<f8",
  "compressor": {"id": "zlib", "level": 1},
  "fill_value": 0,
  "order": "C",
  "filters": null
}
```
