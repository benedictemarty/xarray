# Sprint 3 — `Dataset`

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : regrouper plusieurs `DataArray` cohérents dans un
  `Dataset` et propager indexation et réductions.

## Périmètre engagé

| ID    | User story | État |
|-------|------------|------|
| US-11 | Regrouper des `DataArray` dans un `Dataset` | ✅ |
| US-12 | Indexation propagée (`Isel`/`Sel`) | ✅ |
| US-13 | Fusion de datasets (`Merge`) | ✅ |

## Critères d'acceptation (Definition of Done)

- [x] `gofmt`, `go vet` propres.
- [x] `go test ./...` : tous les tests passent.
- [x] Cohérence dimensions/coordonnées vérifiée et testée.
- [x] Propagation d'indexation et de réductions testée (dont dimension partielle).
- [x] `CHANGELOG.md`, backlog et roadmap à jour.
- [x] Commit atomique.

## Décisions de conception

- **Coordonnées partagées au niveau du dataset** : `NewDataset` agrège les
  coordonnées des variables et refuse toute incohérence (même dimension, tailles
  ou étiquettes différentes).
- **Propagation sélective** : `Isel`/`Sel` et les réductions ne s'appliquent
  qu'aux variables portant la dimension visée ; les autres sont conservées telles
  quelles. Le résultat est reconstruit via `NewDataset`, ce qui revalide la
  cohérence à chaque étape.
- **Immutabilité** : toutes les opérations renvoient un nouveau `Dataset` ; les
  variables sont clonées en profondeur.

## Rétrospective

- **Bien** : réutilisation directe des opérations `DataArray` (Sprint 2) ; la
  reconstruction systématique via `NewDataset` garantit un invariant fort.
- **À surveiller** : le clonage systématique est coûteux (dette T-03,
  performances) ; l'arithmétique entre `Dataset` n'est pas encore exposée.
- **Prochain sprint** : entrées/sorties (US-14 à US-16).
