# Sprint 49 — Fonctions universelles (ufuncs)

- **Période** : démarrage 2026-08-07.
- **Objectif** : appliquer des fonctions élément par élément (US-48).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-48 | `Apply` + ufuncs (`Abs`/`Clip`/`Sqrt`/`Exp`/`Log`/`Pow`) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : Apply/Abs, Clip, Sqrt/Pow, Exp∘Log ≈ id, cas entier.
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- `Apply` s'appuie sur `Variable.mapScalar` + `cloneCoords` (préservation).
- `Abs`/`Clip` restent en `T` ; `Sqrt`/`Exp`/`Log`/`Pow` calculent en float64 puis
  reconvertissent en `T` (troncature documentée pour les entiers).

## Rétrospective

- **Bien** : `Apply` couvre le cas général (équivalent d'appliquer une ufunc) ;
  ufuncs courantes fournies pour l'ergonomie.
- **Suite possible** : ufuncs binaires nommées (`maximum`/`minimum`), `Round`,
  trig.
