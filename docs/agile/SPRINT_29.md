# Sprint 29 — Rolling et Resample

- **Période** : démarrage 2026-08-07.
- **Objectif** : combler deux manques majeurs identifiés dans l'analyse de
  couverture face à xarray (US-30).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-30 | `Rolling` (fenêtre glissante) et `Resample` (rééchantillonnage) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Rolling 1D/2D (Mean/Sum/Min/Max), NaN de bord, erreurs.
- [x] Resample Mean/Sum 1D/2D, bins irréguliers, erreurs.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- **Rolling** : fenêtre « trailing » (valeur i = agrégat de [i-window+1 … i]),
  NaN pour les positions incomplètes, résultat `float64` de même forme — comme
  xarray. Coordonnées conservées (converties en float64).
- **Resample** : faute de gestion du temps, rééchantillonnage sur une coordonnée
  **numérique** par bins réguliers `floor((l-origine)/freq)`. La dimension est
  réduite aux bins non vides (coordonnée = borne gauche). Réutilise
  `groupReduceOn` (mutualisé avec `GroupBy`).

## Limites (assumées)

- Pas encore de `center=`, `min_periods=`, `window` par plusieurs dimensions pour
  Rolling.
- Resample non temporel (pas de `datetime`, pas de fréquences calendaires « 1D »,
  « 1M »…). Les bins vides ne sont pas matérialisés.

## Rétrospective

- **Bien** : `Resample` = un `groupby` par bins → réutilisation directe de
  l'infrastructure existante.
- **Suite possible** : gestion du temps (pour un vrai resample calendaire),
  `min_periods`/`center` pour Rolling.
