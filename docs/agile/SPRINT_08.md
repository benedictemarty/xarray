# Sprint 8 — netCDF (sous-ensemble CDF-1)

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : lever le report de la story netCDF (US-16) en
  implémentant un sous-ensemble auto-cohérent du format classique.

## Périmètre engagé

| ID    | User story | État |
|-------|------------|------|
| US-16 | Lire/écrire du netCDF | ✅ (sous-ensemble) |

## Critères d'acceptation (Definition of Done)

- [x] `gofmt`, `go vet` propres.
- [x] `go test ./...` : tous les tests passent.
- [x] Aller-retour testé pour plusieurs types numériques.
- [x] Périmètre et limites documentés.
- [x] `CHANGELOG.md`, backlog à jour.
- [x] Commit atomique.

## Décisions de conception

- **Format ciblé : netCDF classique CDF-1**, format binaire big-endian
  publiquement documenté, implémentable sans dépendance externe (aucun accès
  réseau requis pour `go get`).
- **Calcul des offsets (`begin`)** : l'en-tête est d'abord écrit dans un tampon
  pour en mesurer la taille (les `begin` ont une taille fixe), puis les offsets
  de données sont calculés et l'en-tête est réécrit avec les bonnes valeurs.
- **Type uniforme** : toutes les variables d'un dataset partagent le type `T`,
  mappé vers un type netCDF. À la lecture, les valeurs stockées sont converties
  vers le `T` demandé.
- **Coordonnées** : représentées comme en netCDF, par des variables 1D nommées
  comme leur dimension ; réattachées aux variables de données à la lecture.

## Limites assumées (documentées)

- Pas de NetCDF-4/HDF5, pas de CDF-5, pas de dimension d'enregistrement
  illimitée, pas d'attributs (listes ABSENT).
- Types exportables : `float64`, `float32`, `int32`, `int16`, `int8`. Les autres
  (`int`/`uint` 64 bits…) renvoient une erreur explicite plutôt qu'un résultat
  incorrect.
- **Validation** : l'aller-retour est vérifié par les tests internes ; il n'a
  **pas** encore été confronté à un outil de référence (`ncdump`), indisponible
  dans l'environnement. À faire avant de garantir l'interopérabilité externe.

## Rétrospective

- **Bien** : implémentation autonome, testée, avec des limites clairement
  énoncées — conforme au principe « ne pas prétendre ce qui n'est pas vérifié ».
- **À surveiller** : interopérabilité réelle à confirmer ; attributs et records
  non gérés.
- **Suite** : publication v0.2.0 ; interopérabilité netCDF, attributs.
