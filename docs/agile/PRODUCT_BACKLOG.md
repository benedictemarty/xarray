# Backlog produit — xarray-go

Vision : offrir en Go l'ergonomie de xarray (Python) — des tableaux
N-dimensionnels étiquetés — avec une API idiomatique et testée, livrée par
incréments.

Priorités : `P0` (indispensable), `P1` (important), `P2` (souhaitable).

## Épopée 1 — Cœur du modèle de données

| ID    | User story | Prio | État |
|-------|------------|------|------|
| US-01 | En tant qu'utilisateur, je veux créer un tableau N-D avec des dimensions nommées afin de donner un sens à chaque axe. | P0 | ✅ Fait (S1) |
| US-02 | En tant qu'utilisateur, je veux attacher des coordonnées à mes dimensions pour indexer par label. | P0 | ✅ Fait (S1) |
| US-03 | En tant qu'utilisateur, je veux sélectionner par position (`isel`) et par label (`sel`). | P0 | ✅ Fait (S1) |
| US-04 | En tant qu'utilisateur, je veux une représentation textuelle lisible de mes tableaux. | P1 | ✅ Fait (S1) |
| US-05 | En tant qu'utilisateur, je veux des attributs (métadonnées) sur mes tableaux. | P1 | ✅ Fait (S1) |

## Épopée 2 — Opérations

| ID    | User story | Prio | État |
|-------|------------|------|------|
| US-06 | En tant qu'utilisateur, je veux des réductions par axe (`sum`, `mean`… le long d'une dimension). | P0 | ✅ Fait (S2) |
| US-07 | En tant qu'utilisateur, je veux des opérations arithmétiques élément par élément entre tableaux. | P0 | ✅ Fait (S2) |
| US-08 | En tant qu'utilisateur, je veux le broadcasting par nom de dimension. | P0 | ✅ Fait (S2) |
| US-09 | En tant qu'utilisateur, je veux l'alignement automatique sur les coordonnées avant opération. | P1 | ✅ Fait (S2) |
| US-10 | En tant qu'utilisateur, je veux `transpose` / réordonner les dimensions. | P1 | ✅ Fait (S2) |

## Épopée 3 — Dataset

| ID    | User story | Prio | État |
|-------|------------|------|------|
| US-11 | En tant qu'utilisateur, je veux regrouper plusieurs `DataArray` dans un `Dataset`. | P0 | ✅ Fait (S3) |
| US-12 | En tant qu'utilisateur, je veux indexer un `Dataset` entier (`sel`/`isel` propagé). | P0 | ✅ Fait (S3) |
| US-13 | En tant qu'utilisateur, je veux fusionner des `Dataset`. | P2 | ✅ Fait (S3) |

## Épopée 4 — Entrées / sorties

| ID    | User story | Prio | État |
|-------|------------|------|------|
| US-14 | En tant qu'utilisateur, je veux charger/sauver en CSV. | P1 | 🔜 S4 |
| US-15 | En tant qu'utilisateur, je veux charger/sauver en JSON. | P1 | 🔜 S4 |
| US-16 | En tant qu'utilisateur, je veux lire/écrire du netCDF. | P2 | 🔜 S4 |

## Dette / transverse

| ID    | Sujet | Prio |
|-------|-------|------|
| T-01  | Généraliser le type de données (generics) au-delà de `float64`. | P2 |
| T-02  | Mesure de couverture de tests en intégration continue. | P1 |
| T-03  | Benchmarks de performance. | P2 |
