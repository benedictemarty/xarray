# Sprint 59 — sélection *nearest* conservant la dimension

- **Période** : démarrage 2026-08-07.
- **Objectif** : aligner xarray-go sur le comportement de xarray pour la
  sélection au plus proche voisin en mode *list-like* (US-58).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-58 | `SelNearestKeep` / `SelNearestMany` (nearest conservant la dim) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe (xarray + gocoverage).
- [x] Tests : `SelNearestKeep` (dim + coord conservées), `SelNearestMany`
      (ordre, liste vide → erreur).
- [x] `CHANGELOG.md` / mémoire projet à jour ; `gocoverage` refactoré.

## Décisions de conception

- **Parité avec xarray** : dans xarray, conserver ou supprimer la dimension
  dépend de la forme du label, pas de `method` :
  `sel(x=1.5, method="nearest")` (scalaire) réduit la dim ;
  `sel(x=[1.5], method="nearest")` (liste) la conserve.
  - `SelNearest(label)` ⇔ cas scalaire (déjà présent, dim réduite).
  - `SelNearestMany(labels)` / `SelNearestKeep(label)` ⇔ cas *list-like*
    (nouveau, dim conservée taille N/1).
- **Pourquoi c'est nécessaire** : les exports CoverageJSON/EDR exigent des axes
  explicites ; une dim supprimée casse le domaine. `gocoverage` remplace son
  contournement `SelRange(nearest, nearest)` par `SelNearestKeep`.
- Helper interne `nearestIndex` factorisé (DRY) entre les trois variantes.

## Rétrospective

- **Bien** : le contournement côté service devient une primitive propre et
  générique côté bibliothèque, conforme à la sémantique xarray.
- **Suite** : envisager un `Isel` *list-like* symétrique et l'usage de
  `SelNearestKeep` pour un endpoint « coverage à un point » (PointSeries).
