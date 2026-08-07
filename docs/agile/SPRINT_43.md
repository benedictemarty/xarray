# Sprint 43 — Composantes temporelles et groupby par composante

- **Période** : démarrage 2026-08-07.
- **Objectif** : approfondir la gestion du temps (accessors + climatologie),
  point faible de la couverture (US-42).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-42 | `ExtractTime` (composantes) + `GroupByTime` (groupby par composante) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : extraction mois/année, climatologie mensuelle (Mean/Sum), erreurs.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Composantes extraites via `time` (stdlib) depuis la coordonnée epoch.
- `GroupByTime` regroupe par valeur de composante (mois 1..12, etc.) et réutilise
  la structure `Resample` (agrégations `groupReduceOn`) — étiquettes = valeurs de
  composante.
- Distinction avec `ResampleCalendar` : ce dernier garde chaque période séparée
  (jan 2020 ≠ jan 2021) ; `GroupByTime(CompMonth)` les réunit (climatologie).

## Rétrospective

- **Bien** : mutualisation avec `Resample`/`groupReduceOn` ; couvre le cas
  climatologique très courant en géosciences.
- **Suite possible** : accessor `.dt` sous forme de DataArray, groupby saisonnier
  (DJF/MAM/…), fréquences composées.
