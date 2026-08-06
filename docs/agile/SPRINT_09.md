# Sprint 9 — GroupBy

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : ajouter le regroupement et l'agrégation par valeur de
  coordonnée (US-18), fonctionnalité analytique la plus attendue.

## Périmètre engagé

| ID    | User story | État |
|-------|------------|------|
| US-18 | `groupby` + agrégations (`Sum`/`Mean`/`Min`/`Max`) | ✅ |

## Critères d'acceptation (Definition of Done)

- [x] `gofmt`, `go vet` propres.
- [x] `go test ./...` : tous les tests passent.
- [x] Regroupement testé en 1D et 2D, avec un type entier.
- [x] `CHANGELOG.md`, backlog à jour.
- [x] Commit atomique.

## Décisions de conception

- **Regroupement par coordonnée de dimension** : conforme à notre modèle où une
  coordonnée porte le nom de sa dimension. `GroupBy(dim)` regroupe par les valeurs
  (potentiellement répétées) de `coords[dim]`.
- **Mécanique** : pour chaque groupe, `takeAlong` sélectionne les positions puis
  la réduction supprime la dimension ; les tranches sont empilées via `stackDim`
  sur une dimension reprenant le nom `dim`, à étiquettes uniques triées.
- **Type de sortie** : `Mean` produit du `float64` (comme les réductions par axe),
  les autres conservent `T`, via la fonction libre générique `groupReduce`.

## Différence avec xarray (Python)

xarray permet de regrouper par **n'importe quelle** coordonnée (y compris
non-dimensionnelle) et par « bins ». Ici, faute de coordonnées non-dimensionnelles
dans le modèle, on regroupe par la coordonnée de dimension. C'est un sous-ensemble
utile ; l'extension aux coordonnées arbitraires dépend d'une évolution du modèle.

## Rétrospective

- **Bien** : `stackDim` est réutilisable (futur `concat`) ; réutilisation des
  réductions par axe existantes.
- **Prochain sprint** : volet performance comparé xarray-go vs xarray Python
  (US-19).
