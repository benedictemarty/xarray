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
| US-17 | En tant qu'utilisateur, je veux choisir la stratégie de jointure (inner/outer/left/right) lors des opérations. | P1 | ✅ Fait (S6) |
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
| US-14 | En tant qu'utilisateur, je veux charger/sauver en CSV. | P1 | ✅ Fait (S4) |
| US-15 | En tant qu'utilisateur, je veux charger/sauver en JSON. | P1 | ✅ Fait (S4) |
| US-16 | En tant qu'utilisateur, je veux lire/écrire du netCDF. | P2 | ✅ Fait (S8, sous-ensemble CDF-1) |

## Épopée 5 — Analyse

| ID    | User story | Prio | État |
|-------|------------|------|------|
| US-18 | En tant qu'utilisateur, je veux regrouper (`groupby`) et agréger par valeur de coordonnée. | P1 | ✅ Fait (S9) |
| US-19 | En tant qu'utilisateur, je veux comparer les performances xarray-go vs xarray (Python). | P2 | ✅ Fait (S10) |
| US-20 | En tant qu'utilisateur, je veux des jointures externes (inner/outer/left/right). | P1 | ✅ Fait (S6) |
| US-21 | En tant qu'utilisateur, je veux `concat` et `stack`. | P1 | ✅ Fait (S12) |
| US-22 | En tant qu'utilisateur, je veux des réductions ignorant les NaN (`skipna`). | P1 | ✅ Fait (S17) |
| US-23 | En tant qu'utilisateur, je veux `Dataset.GroupBy`. | P1 | ✅ Fait (S18) |
| US-24 | En tant qu'utilisateur, je veux un moteur de calcul dense performant (`ndarray`). | P2 | ✅ Fait (S15-16) |
| US-25 | En tant qu'utilisateur, je veux lire/écrire du Zarr (v2). | P1 | ✅ Fait (S24) |
| US-26 | En tant qu'utilisateur, je veux de l'algèbre linéaire et du ML classique. | P2 | ✅ Fait (S22-23) |
| US-27 | En tant qu'utilisateur, je veux lire du GRIB2 (grille lat/lon, simple packing). | P2 | ✅ Fait (S26, sous-ensemble) |
| US-28 | En tant qu'utilisateur, je veux lire le GRIB2 complex packing (templates 5.2/5.3). | P2 | ✅ Fait (S27, validé ecCodes) |
| US-29 | En tant qu'utilisateur, je veux lire les templates GRIB locaux (ex. 50002). | P2 | ✅ Fait (S28, backend ecCodes) |
| US-30 | En tant qu'utilisateur, je veux `rolling` (fenêtre glissante) et `resample`. | P1 | ✅ Fait (S29) |
| US-31 | En tant qu'utilisateur, je veux des coordonnées temporelles et un resample calendaire. | P1 | ✅ Fait (S30) |
| US-32 | En tant qu'utilisateur, je veux gérer les données manquantes (fillna/dropna/ffill). | P1 | ✅ Fait (S31) |
| US-33 | En tant qu'utilisateur, je veux une indexation par label avancée (nearest/plage/liste). | P1 | ✅ Fait (S32) |
| US-34 | En tant qu'utilisateur, je veux `where` et l'interpolation des NaN (`interpolate_na`). | P1 | ✅ Fait (S33) |
| US-35 | En tant qu'utilisateur, je veux des réductions statistiques (var/std/median) et cumulatives (cumsum/diff). | P1 | ✅ Fait (S34) |
| US-36 | En tant qu'utilisateur, je veux argmin/argmax, quantile et cumprod. | P2 | ✅ Fait (S35) |
| US-37 | En tant qu'utilisateur, je veux idxmin/idxmax (étiquette à l'extremum). | P2 | ✅ Fait (S36) |
| US-38 | En tant qu'utilisateur, je veux une évaluation paresseuse par chunks (out-of-core, esprit dask). | P1 | ✅ Fait (S37, MVP) |

## Dette / transverse

| ID    | Sujet | Prio |
|-------|-------|------|
| T-01  | Généraliser le type de données (generics) au-delà de `float64`. | P2 — ✅ Fait (S5) |
| T-02  | Mesure de couverture de tests en intégration continue. | P1 |
| T-03  | Benchmarks de performance. | P2 — ✅ Fait (S7) |
