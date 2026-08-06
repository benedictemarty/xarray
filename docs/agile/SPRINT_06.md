# Sprint 6 — Jointures externes

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : étendre l'alignement au-delà de la jointure interne
  (US-17), avec remplissage des valeurs manquantes.

## Périmètre engagé

| ID    | User story | État |
|-------|------------|------|
| US-17 | Choisir la stratégie de jointure (inner/outer/left/right) | ✅ |

## Critères d'acceptation (Definition of Done)

- [x] `gofmt`, `go vet` propres.
- [x] `go test ./...` : tous les tests passent.
- [x] Les quatre stratégies testées (1D) + un cas 2D.
- [x] `CHANGELOG.md`, backlog à jour.
- [x] Commit atomique.

## Décisions de conception

- **`JoinType`** : `inner` (défaut historique, `Add`/`Sub`… inchangés), `outer`,
  `left`, `right`.
- **Valeur de remplissage explicite** : `AddJoin(other, join, fill)`. Contrairement
  à xarray/Python qui remplit par `NaN`, on exige une valeur car les types
  entiers n'ont pas de NaN. L'appelant passe `0`, `NaN` (pour les flottants), etc.
- **Mécanique** : `Variable.takeFill` accepte l'indice sentinelle `-1` pour
  produire une tranche remplie ; `DataArray.reindex` calcule les indices depuis
  les étiquettes cibles et met à jour la coordonnée.
- **Rétrocompatibilité** : les opérations existantes (`Add`, etc.) conservent la
  jointure interne.

## Rétrospective

- **Bien** : `reindex` est une brique réutilisable (utile aussi pour un futur
  `Dataset.reindex`).
- **À surveiller** : seules les dimensions de l'opérande gauche portant des
  coordonnées des deux côtés sont jointes (cohérent avec l'inner existant).
- **Prochain sprint** : benchmarks et performances (T-03).
