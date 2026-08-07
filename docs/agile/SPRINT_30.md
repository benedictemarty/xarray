# Sprint 30 — Gestion du temps + resample calendaire

- **Période** : démarrage 2026-08-07.
- **Objectif** : introduire les coordonnées temporelles et un rééchantillonnage
  par période civile (US-31), prochain manque majeur face à xarray.

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-31 | Coordonnées temporelles + `ResampleCalendar` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Aller-retour epoch, resample mensuel et annuel testés.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- **Temps = secondes epoch Unix (UTC), en `float64`.** S'intègre au modèle
  générique sans type dédié ; conversions `EpochSeconds`/`TimeFromEpoch`/
  `EpochCoord` via `time` (stdlib).
- **`ResampleCalendar`** tronque chaque instant au début de sa période civile
  (heure/jour/mois/année) et regroupe ; réutilise `Resample` (donc `groupReduceOn`).
  Étiquettes = débuts de période (epoch).

## Limites (assumées)

- Précision ~microseconde (mantisse float64), pas la nanoseconde.
- Périodes calendaires standard uniquement (pas de fréquences « 15min », « 3M »
  arbitraires, ni calendriers 360-day/noleap des géosciences).
- Pas d'accessor `.dt` complet (year/month/... exposés indirectement via la
  troncature de période).

## Rétrospective

- **Bien** : `ResampleCalendar` = un `Resample` dont les bins sont des périodes
  civiles → réutilisation directe de l'infra des Sprints 9/29.
- **Suite possible** : accessors temporels (`.dt.month` pour groupby saisonnier),
  fréquences composées, calendriers géophysiques.
