# Sprint 54 — ufuncs supplémentaires et binaires

- **Période** : démarrage 2026-08-07.
- **Objectif** : compléter les fonctions élément par élément (US-53).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-53 | `Round`/`Floor`/`Ceil`/`Sign`/`Sin`/`Cos`/`Tanh`, `Maximum`/`Minimum` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : arrondis/signe, trigonométrie, Maximum/Minimum (+ broadcasting).
- [x] `CHANGELOG.md` / backlog à jour ; commit atomique.

## Décisions de conception

- Unaires via `Apply` (float64 intermédiaire). `Sign` utilise une valeur runtime
  `-one` (le littéral `-1` ne compile pas pour la contrainte `Number` qui inclut
  les types non signés).
- Binaires via `binary` → alignement et broadcasting hérités gratuitement.

## Rétrospective

- **Bien** : ergonomie proche de NumPy ; `Maximum`/`Minimum` bénéficient de
  l'alignement/broadcasting existant.
