# xarray-go

Tableaux N-dimensionnels **étiquetés** en Go, inspirés de la bibliothèque
Python [xarray](https://xarray.dev/).

L'idée : manipuler des tableaux multi-dimensionnels dont les axes portent un
**nom** (`temps`, `latitude`, `longitude`…) et des **coordonnées** (des
étiquettes réelles), afin d'indexer par label plutôt que par position numérique.

> ⚠️ Projet en cours de construction, livré **par incréments** en méthode agile.
> Les tableaux sont **génériques** sur un type numérique (`int`, `uint`,
> `float32`, `float64`… via la contrainte `Number`).

## Installation

```bash
go get github.com/benedictemarty/xarray
```

## Exemple

```go
package main

import (
	"fmt"

	"github.com/benedictemarty/xarray"
)

func main() {
	da, _ := xarray.NewDataArray(
		[]string{"temps", "lieu"},
		[]int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{
			"temps": {2020, 2021},
			"lieu":  {10, 20, 30},
		},
		"température",
	)

	fmt.Println(da)

	// Indexation par label : la ligne correspondant à l'année 2021.
	sub, _ := da.Sel("temps", 2021)
	fmt.Println(sub.Data()) // [4 5 6]

	fmt.Println("moyenne :", da.Mean()) // 3.5
}
```

## Architecture

Trois niveaux, calqués sur xarray :

| Niveau      | Rôle                                                                 | État        |
|-------------|----------------------------------------------------------------------|-------------|
| `Variable`  | Données plates + dimensions nommées + attributs (brique de base)     | ✅ Sprint 1 |
| `DataArray` | `Variable` + coordonnées étiquetées + nom (indexation par label)     | ✅ Sprint 1 |
| `Dataset`   | Collection de `DataArray` partageant dimensions et coordonnées       | ✅ Sprint 3 |

Voir [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) pour les détails.

## Feuille de route

Voir [`docs/agile/ROADMAP.md`](docs/agile/ROADMAP.md) et le
[backlog produit](docs/agile/PRODUCT_BACKLOG.md).

- **Sprint 1** — Cœur `Variable` / `DataArray` (indexation `isel`/`sel`, réductions) ✅
- **Sprint 2** — Opérations : broadcasting, alignement, arithmétique, réductions par axe ✅
- **Sprint 3** — `Dataset` (regroupement, `sel`/`isel` et réductions propagés, fusion) ✅
- **Sprint 4** — Entrées/sorties : JSON et CSV (aller-retour) ✅
- **Sprint 5** — Généralisation des types (generics `int`/`float32`/`float64`…) ✅
- **Sprint 6** — Jointures externes (inner/outer/left/right) ✅
- **Sprint 7** — Benchmarks et optimisations ✅
- **Sprint 8** — netCDF, sous-ensemble classique CDF-1 (aller-retour) ✅

> Note generics : les constructeurs infèrent le type depuis les données.
> Les fonctions de lecture demandent un paramètre de type explicite, ex.
> `xarray.ReadDataArrayCSV[float64](r)`.

## Entrées / sorties

```go
// JSON (DataArray ou Dataset)
_ = da.WriteJSON(w)
da2, _ := xarray.ReadDataArrayJSON[float64](r)
_ = ds.WriteJSON(w)
ds2, _ := xarray.ReadDatasetJSON[float64](r)

// CSV « tidy » : une ligne par cellule (colonnes = dimensions + valeur)
_ = da.WriteCSV(w)
da3, _ := xarray.ReadDataArrayCSV[float64](r)

// netCDF classique (CDF-1), sous-ensemble
_ = ds.WriteNetCDF(w)
ds3, _ := xarray.ReadDatasetNetCDF[float64](r)

// Zarr v2 (chunké, compression none/zlib) — interop zarr-python vérifiée
_ = xarray.WriteDataArrayZarr("data.zarr", da, []int{2, 3}, xarray.ZarrZlib)
da4, _ := xarray.ReadDataArrayZarr("data.zarr")
```

Voir [`docs/ZARR.md`](docs/ZARR.md) pour le périmètre Zarr et la validation
d'interopérabilité.

> Limites netCDF : sous-ensemble du format classique (CDF-1) — pas de
> NetCDF-4/HDF5, ni records illimités, ni attributs. Types exportables :
> `float64`, `float32`, `int32`, `int16`, `int8`. L'aller-retour est validé en
> interne ; l'interopérabilité avec les outils netCDF de référence reste à
> confirmer.

Exemple de CSV produit pour un tableau `température(temps, lieu)` :

```csv
temps,lieu,temperature
2020,10,1
2020,20,2
...
```

## Performances

Comparaison mesurée face à xarray (Python/NumPy) — voir
[`docs/BENCHMARKS.md`](docs/BENCHMARKS.md). En résumé : xarray-go domine les
réductions et le `groupby` (5×–14× plus rapide), tandis que NumPy garde
l'avantage sur le calcul élément par élément à grande taille (vectorisation).

## Développement

```bash
go test ./...    # lancer tous les tests
go vet ./...     # analyse statique
```

Les modifications sont tracées dans [`CHANGELOG.md`](CHANGELOG.md) et versionnées
avec git.

## Licence

MIT — voir [`LICENSE`](LICENSE).
