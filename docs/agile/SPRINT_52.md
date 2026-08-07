# Sprint 52 — GroupByBins (intervalles arbitraires)

- **Période** : démarrage 2026-08-07.
- **Objectif** : regroupement par intervalles définis explicitement (US-51).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-51 | `DataArray.GroupByBins` + `Dataset.GroupByBins` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : bins, valeurs hors bornes ignorées, Dataset, bornes insuffisantes.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Intervalles `[edges[k], edges[k+1])`, dernier fermé à droite (proche de
  pandas.cut) ; valeurs hors bornes ignorées. Étiquette = borne gauche.
- Helper `binEdgesGroups` réutilisé DataArray/Dataset, comme les autres
  regroupements ; construit un `Resample`/`DatasetGroupBy`.

## Rétrospective

- **Bien** : complète la famille des regroupements (résample régulier, calendaire,
  composante, coarsen, bins) sur la même infrastructure.
