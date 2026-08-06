# Architecture

`xarray-go` reprend l'architecture en couches de xarray (Python).

```
Dataset            (collection de DataArray — Sprint 3)
   └── DataArray   (Variable + coordonnées + nom)
          └── Variable   (données + dimensions nommées + attributs)
```

## Variable

Brique de base. Un tableau N-dimensionnel « nu » :

- `data []float64` : stockage **plat**, en **ordre C** (row-major : le dernier
  axe varie le plus vite) ;
- `dims []string` : nom de chaque axe (unique, non vide) ;
- `shape []int` : taille de chaque axe ;
- `attrs map[string]string` : métadonnées libres.

L'indice plat d'un multi-indice `(i0, i1, …)` se calcule via les *strides* en
ordre C : `flat = Σ iₖ · strideₖ`, avec `strideₖ = Π shapeⱼ` pour `j > k`.

`Isel(dim, index)` fixe une position sur un axe et renvoie une `Variable` de
rang inférieur (l'axe disparaît).

## DataArray

Une `Variable` enrichie :

- `coords map[string]*Variable` : à chaque dimension peut être associé un
  vecteur d'étiquettes (une `Variable` 1D de même longueur que la dimension) ;
- `name string`.

Cela ajoute l'**indexation par label** : `Sel(dim, label)` recherche l'étiquette
dans la coordonnée puis délègue à `Isel`. Lorsqu'une dimension est réduite, sa
coordonnée est retirée du résultat.

## Choix de conception

- **float64 uniquement** pour ce premier incrément : simplicité et performance.
  La généralisation (generics) est une dette identifiée (T-01).
- **Immutabilité par copie** : les constructeurs et opérations copient les
  tranches d'entrée ; `clone` fait une copie profonde. Cela évite les alias
  surprenants au prix d'allocations — acceptable à ce stade.
- **Erreurs explicites** plutôt que panics : toute construction/indexation
  invalide renvoie une `error` idiomatique.

## Évolutions prévues

- Sprint 2 : broadcasting **par nom de dimension** (et non par position comme
  NumPy), alignement sur coordonnées, arithmétique, réductions par axe,
  `Transpose`.
- Sprint 3 : `Dataset`.
- Sprint 4 : entrées/sorties (CSV, JSON, netCDF).
