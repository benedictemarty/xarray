# Sprint 37 — Évaluation paresseuse par chunks (esprit dask)

- **Période** : démarrage 2026-08-07.
- **Objectif** : explorer le pilier « lazy / hors-mémoire » de xarray/dask (US-38).

## Périmètre

| ID | User story | État |
|----|------------|------|
| US-38 | Moteur lazy chunké (Compute parallèle, réductions streaming, out-of-core) | ✅ (MVP) |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] Tests : compute mémoire, réductions, parallélisme (chunks non divisibles),
      out-of-core sur fichier.
- [x] `CHANGELOG.md` / backlog / `docs/LAZY.md` à jour ; commit atomique.

## Décisions de conception

- **Interface `ChunkSource`** : découple le moteur de la source (mémoire, fichier
  binaire, extensible à Zarr). Chunking 1D le long de l'axe 0 (blocs de lignes
  contigus en ordre C → copies/lectures simples).
- **Graphe différé** minimal : liste d'opérations élément par élément appliquées
  par chunk. Pas de graphe multi-tableaux (choix MVP).
- **`Compute` parallèle** : un pool de goroutines traite les chunks ; plages de
  sortie disjointes → aucune synchronisation sur l'écriture.
- **Réductions en streaming** : chaque chunk est lu/transformé/agrégé puis libéré
  → empreinte mémoire ~1 chunk, d'où l'out-of-core (source fichier).

## Limites (assumées, vs dask)

- Pas de chunking multi-dim ni rechunk, pas de graphe entre plusieurs LazyArray,
  pas de réduction par axe en lazy, pas de spilling/cluster, float64 uniquement.

## Rétrospective

- **Bien** : le MVP démontre fidèlement le modèle (différé + chunké + parallèle +
  streaming + out-of-core) tout en restant petit et testé ; `ChunkSource` ouvre
  l'extension (Zarr).
