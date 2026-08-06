# Sprint 26 — Lecture GRIB2 (sous-ensemble)

- **Période** : démarrage 2026-08-06.
- **Objectif** : lire des messages GRIB2 en grille lat/lon (US-27).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-27 | Lecture GRIB2 `regular_ll` + simple packing | ✅ (sous-ensemble) |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Test unitaire autonome (message minimal) + **validation contre ecCodes**.
- [x] Périmètre et limites documentés (`docs/GRIB.md`).
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Périmètre : GRIB **édition 2**, grille **regular_ll** (template 3.0), **simple
  packing** (template 5.0), sans bitmap. Formule `Y = (R + X·2^E) / 10^D`,
  facteurs d'échelle en signe-magnitude, lecteur de bits big-endian.
- **Choix assumé** : ne PAS implémenter le complex/second-order packing (celui de
  la plupart des fichiers opérationnels) — décodage bit-level trop délicat à
  garantir sans la maturité d'ecCodes. Livrer un décodeur fiable sur un
  sous-ensemble plutôt qu'un décodeur douteux sur tout.

## Validation (point fort)

Environnement disposant d'ecCodes/cfgrib et de vrais fichiers GRIB2. Un champ réel
(vent `u` 850 hPa, 201×131) réencodé en simple packing par ecCodes puis décodé par
`ReadGrib` : **26 331 valeurs identiques** (diff max = 0,0).

Le test unitaire a par ailleurs révélé une **inversion `Di`/`Dj`** (masquée sur
les grilles à pas égal), corrigée.

## Rétrospective

- **Bien** : validation contre la référence ecCodes ; le test minimal a attrapé un
  vrai bug.
- **Limite** : les fichiers opérationnels (second-order) restent illisibles →
  ecCodes requis. Extension : templates 5.2/5.3 ou binding ecCodes cgo.
