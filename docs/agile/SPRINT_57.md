# Sprint 57 — Paquet geoapi : CoverageJSON

- **Période** : démarrage 2026-08-07.
- **Objectif** : première brique d'un service géo en Go (OGC API) — export
  CoverageJSON (US-56).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-56 | Paquet `geoapi` + `ToCoverageJSON` (grille lat/lon) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./geoapi/` passe.
- [x] Tests : structure CoverageJSON (domaine/axes/ranges), cas d'erreur.
- [x] `CHANGELOG.md` / `docs/GEOAPI.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Paquet séparé `geoapi`, dépendant uniquement de l'API publique de xarray-go.
- CoverageJSON domaine **Grid** 2D (latitude=y, longitude=x), CRS84, NdArray en
  ordre C (y outer, x inner). Le tableau doit être ordonné (latitude, longitude).
- **Clarification importante** : pygeoapi (Python) utilise le vrai xarray ; ce
  paquet vise un service équivalent **en Go**, pas une intégration à pygeoapi.

## Rétrospective

- **Bien** : brique de sortie standard livrée et testée ; positionnement clarifié.
- **Suite** : subset géospatial de haut niveau (bbox+temps+z), endpoints OGC API
  (HTTP), axes temps/z dans le domaine CoverageJSON.
