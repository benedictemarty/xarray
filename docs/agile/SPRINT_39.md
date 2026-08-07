# Sprint 39 — Graphe lazy multi-tableaux

- **Période** : démarrage 2026-08-07.
- **Objectif** : combiner deux LazyArray dans un graphe de calcul différé (US-40).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-40 | `Add`/`Sub`/`Mul`/`Div` entre deux LazyArray | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : combinaison mémoire, expression composée, réduction, incompat,
      out-of-core (deux fichiers).
- [x] `CHANGELOG.md` / backlog / `docs/LAZY.md` à jour ; commit atomique.

## Décisions de conception

- `binarySource` combine deux LazyArray : pour chaque chunk i, lit les deux
  sources, applique leurs opérations différées respectives, puis combine élément
  par élément. Résultat = un nouveau LazyArray (le graphe peut être poursuivi).
- Contrainte MVP : mêmes forme et même découpage (même NumChunks) ; pas
  d'alignement ni de broadcasting lazy.

## Rétrospective

- **Bien** : compose naturellement avec les sources existantes (mémoire, fichier,
  Zarr) → expressions entre gros tableaux hors-mémoire.
- **Suite possible** : alignement/broadcasting lazy, réductions par axe en lazy.
