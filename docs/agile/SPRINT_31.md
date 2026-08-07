# Sprint 31 — Données manquantes (NaN)

- **Période** : démarrage 2026-08-07.
- **Objectif** : combler le manque « données manquantes » de l'analyse de couverture xarray (US-32).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-32 | `FillNA`/`DropNA`/`FFill`/`BFill`/`CountNA` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests FillNA, DropNA 1D/2D, FFill/BFill, cas entier.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Détection NaN via `math.IsNaN(float64(x))` → sans effet sur les entiers.
- `DropNA(dim)` : how = "any" (défaut xarray) ; supprime une tranche de dim dès
  qu'elle contient un NaN ; réutilise `takeAlong` (coordonnées réindexées).
- `FFill`/`BFill` : propagation le long de dim ; NaN de bord non comblés
  conservés. Helper `forEachLine` factorise le parcours par ligne.

## Rétrospective

- **Bien** : opérations génériques cohérentes avec le reste ; `forEachLine`
  réutilisable (rolling, futures ops par axe).
- **Suite possible** : `interpolate_na` (interpolation linéaire), `how="all"`
  pour DropNA, `where`.
