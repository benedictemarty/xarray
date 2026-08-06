# Sprint 25 — Dataset comme groupe Zarr

- **Période** : démarrage 2026-08-06.
- **Objectif** : stocker un `Dataset` entier au format Zarr (groupe).

## Périmètre

| Sujet | État |
|-------|------|
| `WriteDatasetZarr` / `ReadDatasetZarr` (groupe Zarr v2) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Aller-retour Dataset testé (variables + coordonnées réattachées).
- [x] **Interop groupe vérifiée** avec zarr-python (`zarr.open_group`).
- [x] `CHANGELOG.md` / `docs/ZARR.md` à jour ; commit atomique.

## Décisions de conception

- Groupe = répertoire + `.zgroup` (`{"zarr_format":2}`) ; un sous-array par
  variable et par coordonnée. Coordonnées = arrays 1D nommés comme leur dimension
  (convention xarray, `_ARRAY_DIMENSIONS`).
- Un seul chunk par array (chunk = forme) pour ce MVP ; chunking par variable =
  raffinement futur.
- Refactorisation : `writeZarrArrayInternal` / `readZarrArrayInternal` mutualisés
  entre `DataArray` et `Dataset`.

## Validation d'interopérabilité

Groupe écrit par Go (`temperature`, `pluie`, coords `temps`/`lieu`) ouvert par
`zarr.open_group` (zarr-python 3.3.0) : arrays et coordonnées **identiques**.

## Rétrospective

- **Bien** : réutilisation de la brique array ; interop groupe confirmée.
- **À étendre** : chunking par variable, Zarr v3, autres dtypes.
