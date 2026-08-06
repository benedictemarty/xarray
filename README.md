# xarray-go

Tableaux N-dimensionnels **étiquetés** en Go, inspirés de la bibliothèque
Python [xarray](https://xarray.dev/).

L'idée : manipuler des tableaux multi-dimensionnels dont les axes portent un
**nom** (`temps`, `latitude`, `longitude`…) et des **coordonnées** (des
étiquettes réelles), afin d'indexer par label plutôt que par position numérique.

> ⚠️ Projet en cours de construction, livré **par incréments** en méthode agile.
> Pour ce premier incrément, les données sont stockées en `float64`.

## Installation

```bash
go get github.com/bmarty/xarray
```

## Exemple

```go
package main

import (
	"fmt"

	"github.com/bmarty/xarray"
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
| `Dataset`   | Collection de `DataArray` partageant dimensions et coordonnées       | 🔜 Sprint 3 |

Voir [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) pour les détails.

## Feuille de route

Voir [`docs/agile/ROADMAP.md`](docs/agile/ROADMAP.md) et le
[backlog produit](docs/agile/PRODUCT_BACKLOG.md).

- **Sprint 1** — Cœur `Variable` / `DataArray` (indexation `isel`/`sel`, réductions) ✅
- **Sprint 2** — Opérations : broadcasting, alignement, arithmétique, réductions par axe ✅
- **Sprint 3** — `Dataset`
- **Sprint 4** — Entrées/sorties (CSV, JSON, netCDF)

## Développement

```bash
go test ./...    # lancer tous les tests
go vet ./...     # analyse statique
```

Les modifications sont tracées dans [`CHANGELOG.md`](CHANGELOG.md) et versionnées
avec git.

## Licence

MIT — voir [`LICENSE`](LICENSE).
