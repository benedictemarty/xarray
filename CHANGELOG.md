# Changelog

Toutes les modifications notables de ce projet sont documentées dans ce fichier.

Le format s'appuie sur [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/)
et le projet suit le [versionnage sémantique](https://semver.org/lang/fr/).

## [Non publié]

### Performances (Sprint 14 — SIMD et chemin direct float64)

- **Noyau SIMD AVX** en assembleur (`simd_amd64.s`) : addition `float64` via
  `VADDPD`/YMM déroulé ×4, détection AVX au runtime (`CPUID`/`XGETBV`), repli
  pur-Go. **Conclusion mesurée : le compilateur Go bat ce noyau** (77 vs 30 Go/s)
  car l'opération est memory-bound — le SIMD n'est donc pas branché sur `Add`.
- **Chemin direct `Add` (float64, formes identiques)** sans closure générique :
  la closure `func(T,T) T` du chemin générique n'étant pas inlinée (un appel par
  élément), l'éviter accélère `Add` d'un ordre de grandeur.
  - `Add` 100×100 : 48 µs → **18 µs** (13× plus rapide que NumPy).
  - `Add` 1000×1000 : 3,58 ms → **1,63 ms** (NumPy reste 1,4× devant :
    zéro-initialisation de l'allocateur Go + SIMD).
- `docs/BENCHMARKS.md` : section « Bat-on NumPy ? » et expérience SIMD chiffrée.

### Performances & vérification (Sprint 13 — Calcul vectoriel)

- **Itération incrémentale** dans `binaryOp` : `flatA`/`flatB` maintenus par pas
  (O(1) amorti) au lieu d'un recalcul O(ndim) par élément.
- **Parallélisation multi-cœurs** du calcul élément par élément au-delà de 32 768
  éléments (plages disjointes, sans course de données).
- Effet sur `Broadcast` : 284 µs → **160 µs** (40 k éléments) ; l'écart avec
  NumPy tombe de 2,7× à 1,3×. NumPy reste devant sur le très gros débit (SIMD).
- **Vérification d'équivalence des résultats** : `bench/crosscheck.py` (génère
  `bench/expected.json` via xarray) + test `TestEquivalenceAvecXarray`. Confirme
  que xarray-go et xarray produisent des valeurs identiques (tolérance 1e-9) pour
  add, broadcast, réductions, jointure externe, groupby.
- `docs/BENCHMARKS.md` : section « Pourquoi Go n'a pas d'auto-vectorisation SIMD ».

### Ajouté (Sprint 12 — concat / stack)

- **`Concat(arrays, dim)`** : concaténation le long d'une dimension existante ;
  coordonnée de la dimension concaténée = concaténation des coordonnées.
- **`Stack(arrays, newDim, labels)`** : empilement sur une nouvelle dimension en
  tête (expose la primitive `stackDim`).
- Tests : concat 1D, 2D (axe 0 et axe 1 avec entrelacement), erreurs ; stack et
  erreurs.

### Performances (Sprint 11 — Alignement)

- **`align` — chemin rapide « coordonnées identiques »** : lorsqu'une dimension
  porte des coordonnées identiques des deux côtés, aucune réindexation n'est
  effectuée (suppression de deux copies via `takeAlong`). Gain majeur sur `Add` :
  272 µs → **48 µs**, 135 → **18 allocations**. xarray-go passe devant NumPy sur
  l'arithmétique alignée (voir `docs/BENCHMARKS.md`).

### Ajouté (Sprint 10 — Volet performance xarray-go vs xarray Python)

- **Comparaison de performance** documentée (`docs/BENCHMARKS.md`) contre
  xarray 2026.4.0 / NumPy 2.2.4, sur opérations et tailles identiques.
- **Harnais Python** `bench/xr_bench.py` (miroir des benchmarks Go, calibrage
  adaptatif).
- Benchmark Go `BenchmarkGroupBySum`.
- **Optimisation `binaryOp`** : chemin rapide « dimensions identiques » (boucle
  directe sans calcul d'indices), qui rapproche `Add` des performances NumPy
  (~348 µs → ~250 µs) via `sameDimsShape`.
- Bilan : Go domine réductions et `groupby` (5×–14×) ; NumPy domine le calcul
  élément par élément à grande taille (1,2×–2,1×).

### Ajouté (Sprint 9 — GroupBy)

- **Regroupement `DataArray.GroupBy(dim)`** par les valeurs (répétées) de la
  coordonnée d'une dimension, avec agrégations `Sum`, `Mean`, `Min`, `Max`.
  - `Mean` renvoie du `float64` ; les autres conservent le type `T`.
  - La dimension groupée est remplacée par ses étiquettes uniques triées ; les
    autres coordonnées sont conservées.
- Primitive d'empilement `stackDim` (nouvelle dimension en tête).
- Accès : `GroupBy.Groups`, `GroupBy.Labels`.
- Tests : regroupement 1D/2D, min/max, type entier, cas d'erreur.

## [0.2.0] — 2026-08-06

Deuxième version : jointures externes, optimisations de performance, et support
d'un sous-ensemble du format netCDF classique.

### Ajouté (Sprint 8 — netCDF, US-16)

- **Support d'un sous-ensemble du format netCDF classique (CDF-1)** :
  - `Dataset.WriteNetCDF` / `ReadDatasetNetCDF[T]` ;
  - `DataArray.WriteNetCDF` / `ReadDataArrayNetCDF[T]` (dataset à une variable).
- Périmètre : dimensions fixes, variables numériques
  (`NC_DOUBLE`/`NC_FLOAT`/`NC_INT`/`NC_SHORT`/`NC_BYTE`), coordonnées de
  dimension. Types Go supportés à l'export : `float64`, `float32`, `int32`,
  `int16`, `int8` (les autres renvoient une erreur explicite).
- Non couvert (documenté) : NetCDF-4/HDF5, CDF-5, dimensions d'enregistrement
  illimitées, attributs. L'aller-retour est validé en interne (auto-cohérent),
  pas encore contre un outil netCDF de référence.
- Tests : allers-retours `DataArray`/`Dataset` en float64, float32 et int32 ;
  rejet d'un type non supporté ; signature invalide.

### Performances (Sprint 7 — T-03)

- **Optimisation de `binaryOp`** (arithmétique/broadcasting) : pré-calcul des
  strides par position, supprimant les accès à une map dans la boucle interne.
  Gains mesurés : `Add` ×2,2, `Broadcast` ×6,2, `OuterJoin` ×2,0.
- **Clonage sans revalidation** : nouveau `Variable.cloneVar` (copie profonde
  directe) utilisé sur les chemins internes (`clone`, `Isel`, `reindex`,
  coordonnées…) à la place de `NewVariable`, évitant une double copie et la
  revalidation de données déjà valides.
- **Benchmarks** : suite `bench_test.go` (Add, SumAxis, MeanAxis, Broadcast,
  Clone, WriteCSV, OuterJoin).

### Ajouté (Sprint 6 — Jointures externes)

- **Stratégies de jointure** pour l'alignement des coordonnées avant opération :
  type `JoinType` (`JoinInner`, `JoinOuter`, `JoinLeft`, `JoinRight`).
- **Opérations avec jointure et remplissage** : `AddJoin`, `SubJoin`, `MulJoin`,
  `DivJoin(other, join, fill)`. Les étiquettes manquantes sont comblées par la
  valeur `fill` fournie (nécessaire car il n'existe pas de NaN universel pour les
  entiers).
- Primitives : `Variable.takeFill` (sélection avec indice -1 → remplissage),
  `DataArray.reindex` (réalignement sur des étiquettes cibles).
- Tests : jointures inner/outer/left/right, remplissage personnalisé, cas 2D.

## [0.1.0] — 2026-08-06

Première version publiée : modèle de données étiqueté complet (DataArray,
Dataset), opérations (broadcasting, alignement interne, réductions), I/O
JSON/CSV, et types génériques.

### Modifié (Sprint 5 — Généralisation des types / dette T-01)

- **BREAKING** : `Variable`, `DataArray` et `Dataset` sont désormais **génériques**
  sur un type numérique (nouvelle contrainte `Number` : int/uint/float). Les
  constructeurs infèrent le type depuis les données (`NewDataArray(dims, shape,
  []int{…}, …)` donne un `DataArray[int]`).
  - Les fonctions de lecture requièrent un paramètre de type explicite :
    `ReadDataArrayJSON[float64]`, `ReadDatasetJSON[T]`, `ReadDataArrayCSV[T]`.
- **Réductions** :
  - `Mean`/`MeanAxis` renvoient toujours du `float64` (moyenne d'entiers →
    flottant, comme xarray) ; `Sum`/`Min`/`Max`/`*Axis` conservent le type `T`.
  - `Min`/`Max` d'un tableau vide renvoient la **zéro-valeur** de `T` (il n'existe
    pas de NaN universel pour les entiers) ; `Mean` d'un vide reste `NaN`.
- Helpers internes génériques : `convertNum`, `convertDataArray`, `reduceAxisVar`,
  `reduceAxisDA`, `reduceDatasetAxis`.
- Tests : validation avec `int`, `float32` et `float64` (arithmétique, division
  entière, réductions, I/O JSON/CSV, `Dataset`).

### Ajouté (Sprint 4 — Entrées/sorties)

- **JSON (aller-retour)** :
  - `DataArray` : `WriteJSON`/`ReadDataArrayJSON`, plus `MarshalJSON`/`UnmarshalJSON`
    (validation à la relecture via `NewDataArray`).
  - `Dataset` : `WriteJSON`/`ReadDatasetJSON` ; les coordonnées partagées sont
    réinjectées dans chaque variable à la lecture pour garantir la cohérence.
- **CSV format « tidy »** (une ligne par cellule : une colonne par dimension puis
  la valeur) : `DataArray.WriteCSV`/`ReadDataArrayCSV`. Format général (N-D) et
  sans ambiguïté ; coordonnées reconstruites dans l'ordre d'apparition ; détection
  des grilles incomplètes.
- Tests : allers-retours JSON (`DataArray`, `Dataset`) et CSV, contenu exact,
  cas sans coordonnées, cas d'erreur (CSV vide, en-tête invalide, grille incomplète).

### Reporté

- **netCDF (US-16)** : non implémenté dans cet incrément. Format binaire complexe
  qui nécessiterait une dépendance externe ; conservé au backlog en priorité P2.

### Ajouté (Sprint 3 — `Dataset`)

- **Type `Dataset`** : collection de `DataArray` (« variables de données »)
  partageant un système commun de dimensions et de coordonnées.
  - Construction validée (`NewDataset`) : vérifie la cohérence des tailles de
    dimensions et l'identité des coordonnées partagées entre variables.
  - Accès : `VarNames`, `Get`, `Dims`, `Coord`.
  - Indexation propagée : `Isel` (position) et `Sel` (label via coordonnée
    partagée) appliquées à toutes les variables portant la dimension visée.
  - Réductions propagées : `SumAxis`, `MeanAxis`, `MinAxis`, `MaxAxis` (les
    variables sans la dimension restent inchangées).
  - Gestion des variables : `WithVar`, `DropVars`, `Merge`.
  - Représentation lisible via `String`.
- Helper `DataArray.HasDim`.
- Tests : cohérence, indexation et réductions propagées (y compris dimension
  partielle), fusion, cas d'erreur.

### Ajouté (Sprint 2 — Opérations)

- **`Transpose`** (sur `Variable` et `DataArray`) : réordonnancement des axes par
  permutation des noms de dimensions ; coordonnées conservées.
- **Réductions par axe** (`SumAxis`, `MeanAxis`, `MinAxis`, `MaxAxis`) : réduisent
  une dimension nommée et retirent sa coordonnée du résultat.
- **Arithmétique entre `DataArray`** (`Add`, `Sub`, `Mul`, `Div`) avec :
  - **broadcasting par nom de dimension** (et non par position) ;
  - **alignement automatique** sur les coordonnées (jointure interne sur les
    étiquettes communes) avant l'opération.
- **Opérations scalaires** (`AddScalar`, `MulScalar`) préservant les coordonnées.
- Primitives bas niveau : `Variable.take` (sélection multi-positions),
  `binaryOp` (broadcasting), `reduceAxis`, `mapScalar`.
- Tests : broadcasting, alignement, réductions par axe (2D et 3D), scalaires,
  cas d'erreur (tailles incompatibles, absence d'étiquette commune).

### Ajouté (Sprint 1 — Cœur `Variable` / `DataArray`)

- **Type `Variable`** : tableau N-dimensionnel bas niveau (données `float64` à plat
  en ordre C, dimensions nommées, attributs).
  - Construction validée (`NewVariable`), propriétés (`Dims`, `Shape`, `Ndim`,
    `Size`, `Data`, `Attrs`).
  - Indexation positionnelle : `At`, `Isel` (réduction d'une dimension).
  - Représentation lisible via `String`.
- **Type `DataArray`** : `Variable` + coordonnées étiquetées + nom.
  - Construction validée (`NewDataArray`) avec coordonnées de dimension.
  - Indexation par position (`Isel`) et par label (`Sel`).
  - Réductions globales : `Sum`, `Mean`, `Min`, `Max`.
  - Copie profonde (`clone`), `Rename`, accès aux coordonnées (`Coord`).
- **Tests** : couverture des cas nominaux et d'erreur pour `Variable` et `DataArray`.
- **Documentation projet** : README, backlog produit, roadmap, note de sprint,
  document d'architecture.

## [0.0.0] — 2026-08-06

### Ajouté

- Initialisation du projet : module Go `github.com/bmarty/xarray`, dépôt git,
  cadre de gestion agile.
