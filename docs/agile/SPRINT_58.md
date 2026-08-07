# Sprint 58 — geoapi : sous-échantillonnage géospatial

- **Période** : démarrage 2026-08-07.
- **Objectif** : briques de provider (subset bbox, position) + clarifier le
  positionnement vs gogeoapi (US-57).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-57 | `SubsetBBox` + `Position` (provider OGC API) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : subset bbox (2×2), position nearest (+ coin).
- [x] `CHANGELOG.md` / `docs/GEOAPI.md` / backlog à jour ; commit atomique.

## Décisions de conception / positionnement

- **gogeoapi est un projet distinct** (serveur OGC API en Go). xarray-go ne doit
  PAS le réimplémenter : son rôle est la **couche de données** (provider) —
  lecture, subset, extraction, CoverageJSON — comme xarray l'est pour pygeoapi.
- `SubsetBBox`/`Position` s'appuient directement sur `SelRange`/`SelNearest` de
  xarray-go ; sélection temps/z via les mêmes primitives sur les dimensions
  correspondantes.

## Rétrospective

- **Bien** : recadrage utile (provider vs serveur) évitant de réinventer gogeoapi ;
  briques minimales et testées.
- **Suite** : encoder les NaN en `null` (CoverageJSON), axes temps/z dans le
  domaine, exemple d'intégration comme provider d'un serveur OGC API Go.
