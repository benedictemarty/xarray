# Sprint 44 — Groupby saisonnier

- **Période** : démarrage 2026-08-07.
- **Objectif** : agrégation par saison météorologique (US-43).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-43 | `CompSeason` (DJF/MAM/JJA/SON) + `SeasonName` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : groupby saisonnier (Mean), noms, extraction.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Saison = `(mois % 12) / 3` → DJF=0, MAM=1, JJA=2, SON=3 (convention
  météorologique). Se branche sur `GroupByTime`/`ExtractTime` sans nouvelle
  mécanique.

## Rétrospective

- **Bien** : ajout minimal (une composante) pour un cas d'usage climatologique
  central ; réutilisation totale de l'infrastructure temporelle.
