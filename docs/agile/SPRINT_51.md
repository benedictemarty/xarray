# Sprint 51 — Coordonnées textuelles (types non numériques)

- **Période** : démarrage 2026-08-07.
- **Objectif** : introduire des coordonnées non numériques (US-50).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-50 | Coordonnées string : `WithStrCoord`/`StrCoord`/`SelStr`/`SelStrMany` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe (rien cassé).
- [x] Tests : SelStr, préservation par Isel, SelStrMany, longueur invalide.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- **Contrainte Go assumée** : `DataArray[T]` reste `T Number`. Passer à `T any`
  casserait toutes les méthodes numériques (Go ne conditionne pas une méthode au
  type paramètre). On introduit donc les étiquettes textuelles comme un **champ
  additionnel** `strCoords map[string][]string`, sans toucher aux données ni aux
  opérations.
- Propagation des coordonnées textuelles dans les chemins d'indexation
  (`clone`, `Isel`, `takeAlong`) ; non garantie par les opérations de calcul
  (comme les coordonnées non-index de xarray).

## Rétrospective

- **Bien** : additif, non invasif, couvre le cas réel (stations/catégories) sans
  refonte du cœur générique.
- **Limite honnête** : pas de *données* textuelles/booléennes ni de dtype général ;
  ce serait un refactoring majeur (types parallèles) au rendement discutable en Go.
