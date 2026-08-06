# Changelog

Toutes les modifications notables de ce projet sont documentées dans ce fichier.

Le format s'appuie sur [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/)
et le projet suit le [versionnage sémantique](https://semver.org/lang/fr/).

## [Non publié]

### Ajouté (Sprint 1 — Cœur `Variable` / `DataArray`)

- **Type `Variable`** : tableau N-dimensionnel bas niveau (données `float64` à plat
  en ordre C, dimensions nommées, attributs).
  - Construction validée (`NewVariable`), propriétés (`Dims`, `Shape`, `Ndim`,
    `Size`, `Data`, `Attrs`).
  - Indexation positionnelle : `At`, `Isel` (réduction d'une dimension).
  - Représentation lisible via `String`.
- **Type `DataArray`** : `Variable` + coordonnées étiquetées + nom.
  - Construction validée (`NewDataArray`) avec coordonnées de dimension.
  - Indexation par position (`Isel`) et par label (`Sel`).
  - Réductions globales : `Sum`, `Mean`, `Min`, `Max`.
  - Copie profonde (`clone`), `Rename`, accès aux coordonnées (`Coord`).
- **Tests** : couverture des cas nominaux et d'erreur pour `Variable` et `DataArray`.
- **Documentation projet** : README, backlog produit, roadmap, note de sprint,
  document d'architecture.

## [0.0.0] — 2026-08-06

### Ajouté

- Initialisation du projet : module Go `github.com/bmarty/xarray`, dépôt git,
  cadre de gestion agile.
