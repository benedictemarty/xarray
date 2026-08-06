# Sprint 23 — ML classique (paquet `ml`)

- **Période** : démarrage 2026-08-06.
- **Objectif** : démontrer un usage ML classique au-dessus de `ndarray`.

## Périmètre

| Sujet | État |
|-------|------|
| Paquet `ml` : `Standardize`, `LinearRegression` (Fit/Predict), `MSE` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./ml/` passe.
- [x] Test de convergence sur données synthétiques.
- [x] `CHANGELOG.md` à jour ; commit atomique.

## Décisions de conception

- `LinearRegression` par **descente de gradient** (utilise `MatVec` et `T`),
  `Standardize` (centrage-réduction par colonne, aide la convergence).
- **Portée assumée** : ML classique/pédagogique. Pas d'autograd ni de GPU — pour
  l'apprentissage profond, Python (PyTorch/TF) reste la référence ; Go sert surtout
  à l'inférence en production.

## Résultat

Le modèle apprend `y = 2·x₁ + 3·x₂ + 1` : poids retrouvés (~[2, 3]), biais ~1,
MSE ≈ 0.

## Rétrospective

- **Bien** : `matmul`/`MatVec`/`T` du Sprint 22 réutilisés directement.
- **Positionnement** : xarray-go + ndarray + ml = préparation de données et ML
  classique en Go pur ; le calcul lourd/DL relève d'autres outils.
