# Sprint 33 — Where et InterpolateNA

- **Période** : démarrage 2026-08-07.
- **Objectif** : masquage conditionnel et interpolation des valeurs manquantes
  (US-34), dernières briques « données manquantes » les plus utilisées.

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-34 | `Where`/`WhereFunc`/`InterpolateNA` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests WhereFunc, Where (masque + forme invalide), InterpolateNA
      (linéaire, bords, coordonnée non uniforme).
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- `WhereFunc(keep, other)` : prédicat élément par élément (le plus idiomatique en
  Go, faute de type booléen dédié).
- `Where(mask, other)` : masque = DataArray de **même forme**, non-zéro conserve.
- `InterpolateNA(dim)` : interpolation **linéaire** entre valeurs valides
  encadrantes, pondérée par la **coordonnée** de dim si disponible (sinon la
  position). NaN de bord non comblés (pas d'extrapolation).

## Rétrospective

- **Bien** : réutilisation de `forEachLine` ; complète `fillna`/`ffill`/`bfill`.
- **Suite possible** : `where` avec broadcasting du masque, interpolation d'ordre
  supérieur, extrapolation optionnelle.
