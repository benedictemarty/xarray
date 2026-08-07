# Sprint 47 — Resample / temporel au niveau Dataset

- **Période** : démarrage 2026-08-07.
- **Objectif** : propager Resample/ResampleCalendar/GroupByTime au Dataset (US-46).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-46 | `Dataset.Resample`/`ResampleCalendar`/`GroupByTime` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : resample mensuel/saisonnier/numérique multi-variables.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- **Refactorisation** : extraction des calculs de bins en helpers réutilisables
  (`binGroups`, `calendarGroups`, `componentGroups`), utilisés par les
  constructeurs DataArray *et* Dataset → suppression de la duplication.
- Les versions Dataset construisent un `DatasetGroupBy` (agrégations propagées via
  `dsGroupReduce`, variables sans la dimension conservées/converties).

## Rétrospective

- **Bien** : la mutualisation via helpers a rendu les versions Dataset triviales ;
  cohérence complète DataArray/Dataset pour le temporel.
