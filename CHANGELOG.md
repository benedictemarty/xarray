# Changelog

Toutes les modifications notables de ce projet sont documentées dans ce fichier.

Le format s'appuie sur [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/)
et le projet suit le [versionnage sémantique](https://semver.org/lang/fr/).

## [Non publié]

### Ajouté (Sprint 74 — démo runnable NDVI + docs/SATELLITE.md)

- **`cmd/ndvi`** : démonstration exécutable (`go run ./cmd/ndvi`) de la chaîne
  raster complète : scène 2 bandes → **NDVI = (NIR−ROUGE)/(NIR+ROUGE)** (via
  l'arithmétique DataArray) → **géoréférencement** (affine + CRS) → **carte
  ASCII** + stats → **découpe par emprise** (coords monde) → **export Zarr**
  (groupe, zstd). L'export est relu **exactement par zarr-python** (aller-retour
  validé).
- **`docs/SATELLITE.md`** : documentation de bout en bout du pipeline imagerie
  satellite (lire → DecodeCF → géoréférencer → découper → exporter), matrice de
  capacités validées, limites assumées, note d'accès MTG/EUMETSAT.

### Ajouté (Sprint 73 — géoréférencement automatique depuis les métadonnées CF)

- **`ParseGDALGeoTransform`** : lit une chaîne GeoTransform GDAL
  « x0 dx rx y0 ry dy » en `Affine`.
- **`Dataset.GeoRefFromCF(varName)`** : extrait automatiquement le
  géoréférencement d'une variable via la convention CF/rioxarray — l'attribut
  `grid_mapping` désigne une variable de CRS portant `crs_wkt`/`spatial_ref` et
  `GeoTransform`. Renvoie un `GeoRef` prêt pour `DataArray.Georeference`.
- Ferme la boucle satellite : le CRS et l'affine sont désormais **lus depuis le
  fichier** (plus besoin de les fournir à la main). Validé sur une fixture Zarr à
  la convention rioxarray (`testdata/zarr_georef`). Test `TestGeoRefFromCF`.

### Ajouté (Sprint 72 — géoréférencement raster : affine + CRS)

- **`Affine`** (géotransformation 2D, convention `affine`/GDAL) : `Apply`
  (pixel→monde), `Inverse` (monde→pixel), `FromGDAL`/`GDAL` (GeoTransform à
  6 coefficients). Validé contre la bibliothèque `affine` (rasterio).
- **`GeoRef{Transform, CRS}`** et **`DataArray.Georeference(gr, xDim, yDim)`** :
  attache des coordonnées monde (**centres de pixels**) aux axes x/y d'un raster
  2D et transporte le **CRS** (identifiant opaque : `EPSG:xxxx`/WKT/proj) dans
  l'attribut `crs`. **`GeoCoords`** génère les axes 1D (grille sans rotation ;
  rotation refusée par erreur explicite).
- Périmètre assumé : mapping pixel↔coordonnées + transport du CRS ; **pas de
  reprojection** (nécessiterait PROJ). Complète le décodage CF pour l'imagerie
  satellite (localiser les pixels après dépacking). Tests `geoaffine_test.go`.

### Ajouté (Sprint 71 — attributs CF à la lecture Zarr → dépacking satellite)

- **Capture des attributs `.zattrs` (v2) et `attributes` (v3)** dans
  `Variable.attrs` à la lecture Zarr : `units`, `long_name`, `scale_factor`,
  `add_offset`, `_FillValue`… (les clés structurelles `_ARRAY_DIMENSIONS`/`name`/
  `coords` restent traitées à part). Les valeurs numériques sont converties en
  chaînes, comme côté netCDF.
- Effet : **`DecodeCF` fonctionne désormais sur les données lues depuis Zarr** —
  cas typique des produits satellite **int16 + `scale_factor`/`add_offset`**
  (ex. réflectance MTG/Sentinel). Auparavant les attributs étaient ignorés et
  `DecodeCF` était un no-op sur Zarr.
- Validé : fixture `testdata/zarr_cf_packed` (int16, scale/offset) → lecture des
  attributs + dépacking correct. Test `TestZarrReadCFAttrs`.

### Ajouté (Sprint 70 — découpage configurable à l'écriture Zarr)

- **`WriteDatasetZarrChunked`** (v2) et **`WriteDatasetZarrV3Chunked`** (v3) :
  écriture avec un découpage paramétrable via `chunks map[string]int`
  (dimension → taille de chunk, façon `ds.chunk({...})` de xarray). Les
  dimensions absentes de la spec ne sont pas découpées ; tailles bornées à
  `[1, taille de la dimension]`. Permet des accès partiels efficaces sur de
  grands tableaux. `WriteDatasetZarr`/`WriteDatasetZarrV3` délèguent avec `nil`
  (un seul chunk, comportement inchangé).
- Validé : une grille 7×10 découpée `{y:3, x:4}` produit **9 chunks** (3×3),
  relue à l'identique en Go **et** par zarr-python 3.3.0 (`chunks=(3,4)`), en v2
  comme en v3. Test `TestZarrWriteChunked`.

### Ajouté (Sprint 69 — écriture Zarr v3)

- **`WriteDataArrayZarrV3`/`WriteDatasetZarrV3`** : xarray-go produit désormais
  des stores **Zarr v3** (`zarr.json` par nœud, `node_type` group/array,
  `data_type: float64`, `dimension_names`, encodage de clés `default` → `c/0/0`,
  pipeline de codecs `bytes` + compression). Compressions : aucune, `gzip`
  (via `ZarrZlib`), `zstd` (`ZarrZstd`).
- Validé : aller-retour Go (les trois compressions) **et** relecture **exacte par
  zarr-python 3.3.0** (`xr.open_zarr`) — valeurs, coordonnées et
  `dimension_names`. Test `TestZarrWriteV3Roundtrip`.

### Ajouté (Sprint 68 — écriture zstd & lecture Zarr v3)

- **Écriture Zarr compressée zstd** : nouvelle option `ZarrZstd` pour
  `WriteDataArrayZarr`/`WriteDatasetZarr` (codec numcodecs `zstd`,
  `{"id":"zstd","level":5}`). Complète `ZarrNone`/`ZarrZlib`. Symétrie avec la
  lecture ; validé en aller-retour Go **et** relu exact par zarr-python 3.3.0.
  Lecture du codec `zstd` « simple » (id `zstd`) ajoutée en regard.
- **Lecture Zarr v3** (`zarr_v3.go`) : détection automatique (`zarr.json`) et
  routage v2/v3. Gère le pipeline de codecs (`bytes` + `zstd`/`blosc`/`gzip`),
  `data_type` (float/int/uint + endianité), `dimension_names`, l'encodage de clés
  de chunk `default` (`c/0/0`), les coordonnées et les groupes. Validé sur stores
  v3 réels (xarray/zarr-python) : Dataset zstd, array blosc, coordonnées int64.
  Non gérés : `crc32c`, encodage de clé non standard.

### Ajouté (Sprint 67 — codec Blosc zstd)

- **Codec Blosc `zstd`** pris en charge à la lecture, via
  `github.com/klauspost/compress/zstd` (pur Go). Contrairement à LZ4/BLOSCLZ,
  zstd ne découpe pas les blocs (un flux zstd par bloc). Première dépendance
  externe du module (zstd est trop complexe pour un décodeur maison fiable).
  Validé contre un store réel zarr-python. Test `TestZarrReadBloscZstd`.
- **Reliquat bitshuffle non traité (assumé).** Le bitshuffle avec `nelem` non
  multiple de 8 utilise un agencement spécifique (traitement par blocs + tail)
  de la bibliothèque *bitshuffle* que l'on n'a pas pu reproduire de façon
  certaine ; plutôt que de renvoyer des données fausses, ce cas échoue par une
  **erreur explicite** (le cas `nelem` multiple de 8 reste géré). Test
  `TestBitUnshuffleRemainder`.

### Corrigé/Ajouté (Sprint 66 — Blosc multi-blocs & bitshuffle)

- **Correctif : décodage Blosc par bloc.** Le filtre (byte-shuffle) et le
  découpage en sous-flux sont appliqués **par bloc**, pas sur le buffer entier.
  Les stores dont le chunk s'étale sur **plusieurs blocs** (grands tableaux)
  étaient mal lus (le Sprint 65 n'avait été validé que sur des chunks mono-bloc).
  Détails validés sur stores réels : seuls les blocs **pleins** sont découpés en
  `typesize` sous-flux ; le **dernier bloc partiel** reste un flux unique.
- **Bitshuffle** (`shuffle=2`) désormais pris en charge (transpose de bits par
  bloc) quand `nelem` est multiple de 8 ; sinon erreur explicite (reliquat non
  géré). Complète le byte-shuffle.
- Validé contre zarr-python 3.3.0 (multi-blocs 1M valeurs, bitshuffle). Tests
  hermétiques `TestZarrReadBloscMultiblock` et `TestZarrReadBloscBitshuffle`.

### Ajouté (Sprint 65 — lecture de Zarr réels : Blosc/LZ4 + dtypes)

- **Décodeur Blosc/LZ4 pur Go** (`zarr_blosc.go`) : xarray-go lit désormais les
  stores Zarr produits par **zarr-python avec le compresseur par défaut**
  (Blosc, codec LZ4, byte-shuffle) — auparavant seuls `none`/`zlib` étaient gérés.
  Gère le conteneur Blosc v1 (memcpy + blocs), le découpage en `typesize`
  sous-flux (règle `BLOSC_MIN_BUFFERSIZE=128`), les sous-flux non compressés
  (`clen==neblock`), le décodage LZ4 *block* et le byte-unshuffle. Bitshuffle et
  autres codecs (zstd) non gérés (erreur explicite).
- **Dtypes numériques** : lecture de `<f8/<f4`, `<i8/<i4/<i2/<i1`, `<u*` (et
  boutisme `>`), tous convertis en float64. Les **coordonnées int64** (`<i8`,
  courantes chez zarr-python) et variables entières se lisent enfin.
- **`fill_value` non numérique** : `"NaN"`/`"Infinity"`/`"-Infinity"` gérés.
- Validé contre des stores réels **zarr-python 3.3.0** (memcpy, LZ4 découpé,
  sous-flux brut, coords int64). Tests hermétiques `TestZarrReadBloscLZ4` et
  `TestZarrReadIntDtypes` (+ fixtures `testdata/`).

### Ajouté (Sprint 64 — geoapi : domaine PointSeries)

- **`geoapi.ToCoverageJSON` émet `domainType: "PointSeries"`** lorsque la grille
  est réduite à un point (1×1), au lieu de « Grid » figé — conforme à la
  spécification CoverageJSON. Aligne l'encodeur mono-`DataArray` de `geoapi` sur
  le comportement multi-paramètres de gocoverage. Test `TestToCoverageJSONPointSeries`.

### Corrigé (Sprint 62 bis — lecture des `fill_value` non numériques)

- **`readZarrArrayInternal` décode le `fill_value` via `parseZarrFill`**, gérant
  les jetons JSON `"NaN"`/`"Infinity"`/`"-Infinity"` (le champ `fill_value` du
  `.zarray` est un `json.RawMessage`). Complète le câblage du consommateur.

### Ajouté (Sprint 63 — métadonnées Zarr consolidées)

- **`WriteDatasetZarr` écrit désormais un `.zmetadata` consolidé** (Zarr v2,
  `zarr_consolidated_format: 1`) agrégeant tous les `.zgroup`/`.zarray`/`.zattrs`
  du store. Permet à zarr-python/xarray d'ouvrir avec `consolidated=True` (une
  seule lecture, sans parcourir l'arborescence) et supprime le `RuntimeWarning`
  « consolidated metadata not found » observé auparavant.
- Validé : `xr.open_zarr(dir, consolidated=True)` (zarr-python 3.3.0), warnings en
  erreurs, valeurs exactes. Test `TestZarrDatasetConsolidatedMetadata`.

### Corrigé (Sprint 62 — Zarr `fill_value`, interop écriture zarr-python)

- **`fill_value` du `.zarray` passé de `0` à `null`.** Un `fill_value` numérique
  est interprété par xarray/zarr-python comme `_FillValue` : à la relecture, toutes
  les valeurs égales (les **0 légitimes**) étaient **masquées en NaN**. L'écriture
  xarray-go produisait donc du Zarr lu à tort par zarr-python (zéros → NaN).
- Validé empiriquement : `WriteDatasetZarr`/`WriteDataArrayZarr` → relecture par
  **zarr-python 3.3.0** (xarray Python) → **valeurs exactes**, zéros préservés,
  sans et avec compression zlib.
- Test de non-régression `TestZarrFillValueNull` (vérifie `fill_value: null` dans
  `.zarray` et la préservation des zéros à l'aller-retour interne).

### Ajouté (Sprint 61 — ouverture NetCDF-4/HDF5 & CDF-2/5 par conversion)

- **`OpenNetCDFFile(path, conv)`** : ouvre un fichier netCDF quel que soit son
  format. CDF-1 lu directement ; **NetCDF-4/HDF5, CDF-2/5** délégués à un
  **convertisseur externe** (`nccopy -k classic` ou `cdo -f nc copy`) qui les
  réécrit en CDF-1, puis lecture normale. Pas de cgo, pas de lecteur HDF5 en Go :
  un pont pragmatique, dans l'esprit du backend eccodes optionnel pour le GRIB.
- **`SniffNetCDFFormat`** (détection par octets de signature : CDF-1/2/5, HDF5) et
  **`FindNetCDFConverter`** (détection `nccopy`/`cdo` dans le PATH).
- Robustesse : sans convertisseur disponible, un HDF5/CDF-2 échoue par une
  **erreur explicite** (jamais de panic ni de lecture erronée).
- Validé de bout en bout sur un vrai fichier NetCDF-4/HDF5 (superblock v2) avec
  le **chemin de production réel** : `nccopy` (netCDF 4.9.3) détecté
  automatiquement par `FindNetCDFConverter`, conversion → lecture → valeurs
  correctes (`OpenNetCDFFile(path, nil)`). Un second test couvre un convertisseur
  stand-in (xarray Python) pour les environnements sans `nccopy`/`cdo`.

### Ajouté (Sprint 60 — attributs netCDF & décodage CF minimal)

- **Lecture/écriture des attributs netCDF** (globaux et de variable) dans le
  lecteur CDF-1 : types `NC_CHAR` et numériques. Les attributs de variable
  alimentent désormais `Variable.attrs` (`units`, `long_name`, `scale_factor`…).
  Auparavant les listes d'attributs étaient supposées ABSENT, ce qui
  désynchronisait l'analyse de tout fichier réel portant des attributs.
- **`DecodeCF(*Dataset[float64])`** — décodage du *packing* CF façon
  `xarray.decode_cf` : valeur = brut × `scale_factor` + `add_offset`, et
  `_FillValue`/`missing_value` → NaN ; les attributs consommés sont retirés.
- **`DecodeTime(*Dataset[float64], dim)`** et **`DecodeCFTime([]float64, units)`**
  — décodage de l'axe temporel CF « `<unité> since <date>` » (seconds/minutes/
  hours/days) vers des secondes depuis l'epoch Unix.
- **Dimension d'enregistrement illimitée** (`numrecs` > 0) : désormais **lue**
  (variables d'enregistrement entrelacées, désentrelacées à la lecture), au lieu
  d'être refusée. C'est le cas le plus fréquent des fichiers climato réels
  (axe `time` illimité). L'écriture reste en dimensions fixes.
- Motivation : rendre `gocoverage` capable de charger des netCDF portant des
  métadonnées CF réelles (unités pour `get_fields`, valeurs dépackées, temps)
  et un axe temporel illimité.
- Validé empiriquement sur des fichiers écrits par Python xarray
  (`NETCDF3_CLASSIC`, packing int16, `time` illimité).

### Ajouté (Sprint 59 — sélection *nearest* conservant la dimension)

- **`DataArray.SelNearestMany(dim, labels)`** et **`DataArray.SelNearestKeep(dim, label)`** :
  sélection au plus proche voisin qui **conserve la dimension** (taille 1),
  reproduisant fidèlement xarray `sel(dim=[...], method="nearest")`. À l'inverse,
  `SelNearest` (label scalaire) réduit la dimension comme `sel(dim=l)`.
- Motivation : les exports **CoverageJSON / EDR** exigent des axes explicites ;
  une dimension supprimée casse le domaine (Grid/PointSeries). `gocoverage`
  s'appuie désormais sur `SelNearestKeep` au lieu d'un contournement `SelRange(n,n)`.
- Refactorisation interne : helper `nearestIndex` partagé par `SelNearest`,
  `SelNearestKeep` et `SelNearestMany`.

### Ajouté (Sprint 58 — geoapi : sous-échantillonnage géospatial)

- **`geoapi.SubsetBBox`** (sous-cube dans une emprise, via `SelRange`) et
  **`geoapi.Position`** (valeur au point le plus proche, requête EDR *position*,
  via `SelNearest`).
- Clarification de positionnement (`docs/GEOAPI.md`) : **gogeoapi** (serveur OGC
  API en Go) est un projet distinct ; `xarray-go` + `geoapi` jouent le rôle de
  **provider de données** (comme xarray pour pygeoapi), pas de serveur.

### Ajouté (Sprint 57 — Paquet geoapi : CoverageJSON)

- **Nouveau paquet `geoapi`** (briques pour un service OGC API en Go, façon
  pygeoapi) : `ToCoverageJSON` sérialise un `DataArray[float64]` 2D (grille
  latitude × longitude) au format **CoverageJSON** (domaine Grid, CRS84).
- Documentation `docs/GEOAPI.md` : positionnement (pygeoapi = Python ; ce paquet
  = reconstruction en Go), ce que xarray-go couvre déjà (subset/extraction) et ce
  qui reste (endpoints OGC, CRS).

### Ajouté (Sprint 56 — Masques de nullité)

- **`IsNull`/`NotNull`** : masques (1/0) des valeurs manquantes/présentes.
- **`Count`** (nombre de valeurs non-NaN) et **`CountAxis(dim)`** (comptage le
  long d'une dimension).

## [0.14.0] — 2026-08-07

Quatorzième version : arithmétique entre Datasets (Add/Sub/Mul/Div + scalaires).

### Ajouté (Sprint 55 — Arithmétique entre Datasets)

- **`Dataset.Add`/`Sub`/`Mul`/`Div`** : opérations variable par variable (sur les
  variables de même nom, via l'arithmétique DataArray). **`AddScalar`/`MulScalar`**
  s'appliquent à toutes les variables.

## [0.13.0] — 2026-08-07

Treizième version : coordonnées textuelles (stations/catégories), `groupby_bins`
(intervalles arbitraires), `Dot` (contraction tensorielle nommée) et fonctions
universelles supplémentaires (round/sign/trigonométrie, maximum/minimum).

### Ajouté (Sprint 54 — ufuncs supplémentaires et binaires)

- **Arrondis / signe / trigonométrie** : `Round`, `Floor`, `Ceil`, `Sign`, `Sin`,
  `Cos`, `Tanh`.
- **Binaires élément par élément** : `Maximum(other)`, `Minimum(other)` (avec
  alignement et broadcasting).

### Ajouté (Sprint 53 — Dot / contraction tensorielle)

- **`Dot(a, b, dim)`** : contraction de deux DataArrays sur une dimension commune
  nommée (`xr.dot`), dimensions restantes et coordonnées conservées. Couvre
  produit matriciel, matrice-vecteur et produit scalaire.

### Ajouté (Sprint 52 — GroupByBins)

- **`DataArray.GroupByBins(dim, edges)`** et **`Dataset.GroupByBins`** :
  regroupement par intervalles arbitraires (`groupby_bins`) ; intervalles
  `[edges[k], edges[k+1])`, dernier fermé à droite, valeurs hors bornes ignorées.

### Ajouté (Sprint 51 — Coordonnées textuelles)

- **Coordonnées non numériques (string)** sur les dimensions : `WithStrCoord`,
  `StrCoord`, `SelStr`, `SelStrMany`. Données numériques (T), étiquettes texte
  (stations, catégories) ; préservées par l'indexation.

## [0.12.0] — 2026-08-07

Douzième version : propagation d'opérations au Dataset (Rolling, resample/temporel),
restructuration de dimensions (squeeze/expand/rename), fonctions universelles
(Apply/Abs/Sqrt/Clip…), et coarsen (downsampling par blocs).

### Ajouté (Sprint 50 — Coarsen)

- **`DataArray.Coarsen(dim, factor)`** et **`Dataset.Coarsen(dim, factor)`** :
  agrégation par blocs **non chevauchants** (downsampling), boundary "trim".
  Agrégations Sum/Mean/Min/Max. Utile pour réduire la résolution de grilles.

### Ajouté (Sprint 49 — Fonctions universelles)

- **`Apply(fn)`** : fonction arbitraire élément par élément (dimensions et
  coordonnées préservées).
- **ufuncs mathématiques** : `Abs`, `Clip(lo, hi)`, `Sqrt`, `Exp`, `Log`,
  `Pow(p)`. Les transcendantes passent par float64 (troncature vers T pour les
  entiers).

### Ajouté (Sprint 48 — Restructuration de dimensions)

- **`Squeeze(dim)`** : supprime une dimension de taille 1 (et sa coordonnée).
- **`ExpandDims(dim)`** : insère une nouvelle dimension de taille 1 en tête.
- **`RenameDim(old, new)`** : renomme une dimension et sa coordonnée.

### Ajouté (Sprint 47 — Resample/temporel au Dataset)

- **`Dataset.Resample(dim, freq)`**, **`Dataset.ResampleCalendar(dim, period)`**,
  **`Dataset.GroupByTime(dim, comp)`** : rééchantillonnage et regroupement
  temporel propagés à toutes les variables (via `DatasetGroupBy` : Sum/Mean/Min/Max).
- Refactorisation : helpers `binGroups`/`calendarGroups`/`componentGroups`
  mutualisés entre les niveaux DataArray et Dataset (suppression de la
  duplication).

### Ajouté (Sprint 46 — Dataset.Rolling)

- **`Dataset.Rolling(dim, window)`** : fenêtre glissante propagée aux variables
  portant la dimension (agrégations `Mean`/`Sum`/`Min`/`Max` → `Dataset[float64]`) ;
  les variables sans la dimension sont conservées (converties).

## [0.11.0] — 2026-08-07

Onzième version : comparatif lazy vs **dask** (compétitif à supérieur en
out-of-core), réduction lazy déterministe, approfondissement du **temps**
(composantes, climatologie mensuelle et saisonnière), et propagation d'opérations
au **Dataset**.

### Ajouté (Sprint 45 — Propagation au Dataset)

- **`Dataset.VarAxis`/`StdAxis`/`MedianAxis`** : réductions statistiques par axe
  propagées aux variables portant la dimension (résultat `Dataset[float64]`).
- **`Dataset.FillNA(value)`** : remplacement des NaN sur toutes les variables.
- **`Dataset.Cumsum(dim)`** : somme cumulée propagée aux variables concernées
  (les autres conservées).

### Ajouté (Sprint 44 — Groupby saisonnier)

- **`CompSeason`** : composante saison météorologique (0=DJF, 1=MAM, 2=JJA,
  3=SON), utilisable avec `GroupByTime`/`ExtractTime`. `SeasonName(int)` donne le
  nom court. Couvre l'analyse climatologique **saisonnière** (`groupby("time.season")`).

### Ajouté (Sprint 43 — Composantes temporelles et groupby par composante)

- **`ExtractTime(coord, comp)`** : extrait une composante calendaire
  (`CompYear`/`CompMonth`/`CompDay`/`CompHour`/`CompMinute`/`CompWeekday`/
  `CompDayOfYear`) d'une coordonnée en secondes epoch.
- **`GroupByTime(da, dim, comp)`** : regroupement par composante temporelle —
  équivalent de `groupby("time.month")` de xarray (climatologie). Réunit p. ex.
  tous les mois de janvier, quelle que soit l'année ; agrégations Sum/Mean/Min/Max.

### Performances (Sprint 42 — Expression composée vs dask + réduction déterministe)

- **Comparatif expression multi-tableaux** `mean(a*b)` (out-of-core, 2 stores
  Zarr) : xarray-go ~220 ms vs dask ~281 ms (~1,3× plus rapide). Le graphe lazy Go
  n'a pas d'overhead de planification.
- **Réduction lazy rendue déterministe** : les agrégats partiels par chunk sont
  combinés dans l'ordre des chunks (plus de mutex), indépendamment de
  l'ordonnancement des goroutines. Les écarts avec dask relèvent de la précision
  flottante (ordre d'accumulation, non-associativité).
- Harnais : `bench/lazy_expr_bench.py`, `cmd/benchexpr`.

### Ajouté (Sprint 41 — Comparatif lazy vs dask)

- **Comparaison de performance lazy/out-of-core** (`docs/BENCHMARKS.md`) :
  `ChunkZarr(...).Mean()` vs dask `from_zarr().mean().compute()` sur un même store
  Zarr. Résultats identiques ; xarray-go compétitif (32 vs 28 ms à 4 M) et **plus
  rapide sur gros volume** (119 vs 140 ms à 16 M).
- Harnais : `bench/lazy_bench.py` (dask) et `cmd/benchzarr` (Go).

## [0.10.0] — 2026-08-07

Dixième version : **moteur d'évaluation paresseuse par chunks (esprit dask)** —
opérations différées, `Compute` parallèle, réductions en streaming (globales et
par axe), graphe multi-tableaux, et sources hors-mémoire (fichier binaire et
store Zarr v2).

### Ajouté (Sprint 40 — Réductions par axe en lazy)

- **`SumAxis`/`MeanAxis`/`MinAxis`/`MaxAxis` sur `LazyArray`** (tableaux 1D/2D),
  en streaming : réduire l'axe 0 (celui du découpage) accumule entre chunks ;
  réduire l'axe 1 réduit à l'intérieur de chaque bloc. Le résultat (plus petit)
  est matérialisé en `DataArray[float64]`, coordonnées préservées.

### Ajouté (Sprint 39 — Graphe lazy multi-tableaux)

- **Combinaison paresseuse de deux `LazyArray`** : `Add`/`Sub`/`Mul`/`Div`
  élément par élément, chunk par chunk (source `binarySource`). Formes et
  découpages doivent coïncider.
- Permet des expressions différées entre gros tableaux hors-mémoire (fichier,
  Zarr) sans les charger entièrement (ex. `la.MulScalar(2).Sub(lb).Sum()`).

### Ajouté (Sprint 38 — Lazy adossé à Zarr, out-of-core standard)

- **`ChunkZarr(dir, chunkSize)`** : `LazyArray` hors-mémoire adossé à un store
  Zarr v2 (tableaux 1D/2D, `<f8`). Chaque bloc de lignes est reconstruit en ne
  lisant que les chunks Zarr qui le recouvrent → agrégation de tableaux plus
  grands que la RAM sur un format **standard et interopérable**.
- Source `zarrRowSource` (implémente `ChunkSource`), gère chunks non alignés et
  compression zlib.

### Ajouté (Sprint 37 — Évaluation paresseuse par chunks, esprit dask)

- **Moteur lazy `LazyArray`** : opérations différées sur un tableau float64
  découpé en chunks (le long de l'axe 0).
  - `Chunk(da, size)` (source mémoire), `ChunkFile(...)` (source **hors-mémoire**,
    lecture par blocs), `WriteRawF64` ; interface publique `ChunkSource`.
  - Opérations différées `Map`/`AddScalar`/`MulScalar` ; `Compute()` matérialise
    en parallèle (goroutines).
  - Réductions **en streaming** `Sum`/`Mean`/`Min`/`Max` (un chunk à la fois) →
    traitement de données **plus grandes que la RAM**.
- Documentation : `docs/LAZY.md` (modèle, out-of-core, limites vs dask).

## [0.9.0] — 2026-08-07

Neuvième version : `where`/`interpolate_na`, et un ensemble complet de réductions
statistiques et cumulatives (`var`/`std`/`median`/`quantile`,
`argmin`/`argmax`/`idxmin`/`idxmax`, `cumsum`/`cumprod`/`diff`).

### Ajouté (Sprint 36 — IdxMin/IdxMax)

- **`IdxMinAxis`/`IdxMaxAxis`** : étiquette de coordonnée à l'extremum le long
  d'une dimension (équivalent de `idxmin`/`idxmax` de xarray). Nécessite une
  coordonnée.

### Ajouté (Sprint 35 — ArgMin/ArgMax, Quantile, Cumprod)

- **`ArgMinAxis`/`ArgMaxAxis`** : indice (float64) de l'extremum le long d'une
  dimension.
- **`Quantile(q)`** et **`QuantileAxis(dim, q)`** : quantile q ∈ [0,1] par
  interpolation linéaire (méthode « linear » de NumPy).
- **`Cumprod(dim)`** : produit cumulé le long d'une dimension.

### Ajouté (Sprint 34 — Réductions statistiques et cumulatives)

- **Statistiques** : `Var`/`Std`/`Median` (globales) et `VarAxis`/`StdAxis`/
  `MedianAxis` (par axe, `float64`). Variance de population (ddof=0), comme xarray.
- **Cumulatives** : `Cumsum(dim)` (somme cumulée, même forme) et `Diff(dim)`
  (différences successives, dimension réduite de 1, coordonnée = positions 1..n-1).

### Ajouté (Sprint 33 — Where et InterpolateNA)

- **`WhereFunc(keep, other)`** : masquage conditionnel élément par élément.
- **`Where(mask, other)`** : masquage par un DataArray de même forme (non-zéro =
  conserver, sinon `other`).
- **`InterpolateNA(dim)`** : interpolation linéaire des NaN le long de dim, selon
  la coordonnée si présente (sinon la position) ; NaN de bord conservés.

## [0.8.0] — 2026-08-07

Huitième version : **étage analyse** — `rolling`/`resample`, coordonnées
temporelles et resample calendaire, gestion des données manquantes
(`fillna`/`dropna`/`ffill`), et indexation `sel` avancée (nearest/plage/liste).

### Ajouté (Sprint 32 — Indexation `sel` avancée)

- **`SelNearest(dim, label)`** : sélection par plus proche voisin (dimension
  réduite).
- **`SelRange(dim, lo, hi)`** : sélection par plage d'étiquettes [lo, hi] (bornes
  incluses, ordre tolérant).
- **`SelMany(dim, labels)`** : sélection de plusieurs étiquettes (ordre respecté).

### Ajouté (Sprint 31 — Données manquantes)

- **Gestion des NaN** sur `DataArray` : `CountNA`, `FillNA(value)`,
  `DropNA(dim)` (supprime le long de dim les tranches contenant un NaN, how=any),
  `FFill(dim)`/`BFill(dim)` (propagation avant/arrière). Sans effet pour les types
  entiers.
- Helper interne `forEachLine` (itération par ligne le long d'un axe).

### Ajouté (Sprint 30 — Gestion du temps + resample calendaire)

- **Coordonnées temporelles** : `EpochSeconds`/`TimeFromEpoch`/`EpochCoord`
  (temps ↔ secondes epoch UTC, `float64`).
- **`ResampleCalendar(da, dim, period)`** : rééchantillonnage par période
  **civile** (`PeriodHour`/`PeriodDay`/`PeriodMonth`/`PeriodYear`) — équivalent
  de `resample('1M'/'1Y'…).mean()` de xarray ; les étiquettes sont les débuts de
  période. Réutilise `Resample`.
- Note : précision ~microseconde (float64), pas de calendriers non standard.

### Ajouté (Sprint 29 — Rolling et Resample)

- **`DataArray.Rolling(dim, window)`** : fenêtre glissante « trailing » le long
  d'une dimension, agrégations `Mean`/`Sum`/`Min`/`Max` (résultat `float64`, même
  forme, NaN aux bords incomplets — comme xarray).
- **`DataArray.Resample(dim, freq)`** : rééchantillonnage par intervalles
  réguliers d'une coordonnée numérique (binning `floor((l-origine)/freq)`),
  agrégations `Sum`/`Mean`/`Min`/`Max` ; la dimension est réduite aux bins non
  vides (coordonnée = borne gauche). Réutilise `groupReduceOn`.
- Note : `Resample` opère sur une coordonnée numérique (pas encore de gestion du
  temps/`datetime`).

## [0.7.0] — 2026-08-07

Septième version : **backend GRIB via ecCodes** (cgo, opt-in) pour couvrir les
templates locaux non standard (ex. 50002 Météo-France), en complément du
décodeur pur-Go. Le cœur du projet reste 100 % Go pur.

### Ajouté (Sprint 28 — Backend GRIB via ecCodes, cas particuliers)

- **Backend `experimental/eccodesgrib`** (cgo, opt-in `-tags eccodes`) :
  `ReadFile` délègue à la bibliothèque C ecCodes → gère **tout** le GRIB, dont les
  **templates locaux** (ex. **50002 Météo-France**) que le décodeur pur-Go ne
  couvre pas, sans code spécifique par template.
- Stub pur-Go (`//go:build !eccodes`) : le cœur du projet **reste 100 % Go pur**
  et compile sans cgo par défaut.
- **Validé** : lecture d'un vrai fichier 50002 Météo-France → valeurs identiques à
  ecCodes Python (diff max = 0,0). Utilitaire `cmd/readgrib-ec`.
- Documentation : différence 50002 vs standard 5.3, stratégie hybride pur-Go /
  ecCodes (`docs/GRIB.md`, `experimental/eccodesgrib/README.md`).

## [0.6.0] — 2026-08-07

Sixième version : **lecture GRIB2** (grille lat/lon), en simple packing **et**
complex packing / différenciation spatiale (templates 5.0/5.2/5.3), validée au
bit près contre ecCodes.

### Ajouté (Sprint 27 — GRIB2 complex packing + différenciation spatiale)

- **`ReadGrib` gère désormais le complex packing** (templates 5.2 et 5.3) en plus
  du simple packing : références/largeurs/longueurs de groupe avec **alignement à
  l'octet** entre blocs (conforme à g2clib `comunpack`), et **différenciation
  spatiale** d'ordre 1 et 2 (ajout du minimum global + sommation récursive).
- **Validé contre ecCodes** : un vrai champ 201×131 réencodé en
  `grid_complex_spatial_differencing` → **26 331 valeurs identiques** (diff max
  = 0,0). Test unitaire versionné (`testdata/complex_synth.grib2`, données
  synthétiques).
- Les templates de packing **locaux** (ex. 50002 Météo-France) restent non gérés
  (erreur explicite → ecCodes requis).

### Ajouté (Sprint 26 — Lecture GRIB2, sous-ensemble)

- **`ReadGrib`** : lecture GRIB2 pour grille **régulière lat/lon** (`regular_ll`)
  en **simple packing** (template 5.0), sans bitmap. `GribMessage.ToDataArray`
  produit un `DataArray[float64]` (latitude, longitude) avec coordonnées.
- Décodage signe-magnitude des facteurs d'échelle, lecteur de bits, formule
  `Y = (R + X·2^E) / 10^D`.
- **Validé contre ecCodes** : un vrai champ (201×131) réencodé en simple packing
  et décodé par Go → 26 331 valeurs **identiques** (diff max = 0,0). Test unitaire
  autonome sur message minimal. Utilitaire `cmd/readgrib`.
- **Non géré (documenté)** : complex/second-order packing (fichiers opérationnels),
  GRIB1, autres grilles, bitmaps → ecCodes/cfgrib requis. Voir `docs/GRIB.md`.
- Correction : inversion `Di`/`Dj` dans le template de grille (révélée par le test
  unitaire ; sans effet sur les grilles à pas égal).

## [0.5.0] — 2026-08-06

Cinquième version : prise en charge du format **Zarr v2** (arrays et Datasets en
groupes), avec interopérabilité **vérifiée** avec zarr-python dans les deux sens.

### Ajouté (Sprint 25 — Dataset comme groupe Zarr)

- **`WriteDatasetZarr` / `ReadDatasetZarr`** : un `Dataset[float64]` est stocké
  comme **groupe Zarr v2** (`.zgroup` + un sous-array par variable et par
  coordonnée ; coordonnées = arrays 1D nommés comme leur dimension).
- Refactorisation : lecture/écriture d'array extraites en
  `writeZarrArrayInternal` / `readZarrArrayInternal`, partagées entre l'API
  `DataArray` et l'API `Dataset`.
- **Interop groupe vérifiée** : groupe Go relu par `zarr.open_group`
  (zarr-python) — arrays et coordonnées identiques. Utilitaire `cmd/genzarrds`.

### Ajouté (Sprint 24 — Prise en charge de Zarr v2)

- **Lecture/écriture Zarr v2** (`WriteDataArrayZarr`, `ReadDataArrayZarr`) sur
  système de fichiers : `DataArray[float64]`, dtype `<f8`, ordre C, chunking (avec
  `fill_value` pour les bords), compression **none** ou **zlib** (stdlib).
- Dimensions/nom/coordonnées dans `.zattrs` (`_ARRAY_DIMENSIONS`, convention
  xarray).
- **Interopérabilité vérifiée dans les deux sens** avec zarr-python 3.3.0 :
  Go→Python et Python→Go donnent des données identiques (chunks non alignés +
  zlib inclus). Utilitaires `cmd/genzarr` et `cmd/readzarr`.
- Documentation : `docs/ZARR.md`. Périmètre non géré documenté (Zarr v3, autres
  dtypes, blosc/zstd, groupes/Dataset).

## [0.4.0] — 2026-08-06

Quatrième version : accélération de l'arithmétique float64 (noyaux directs sans
closure sur les quatre opérations, broadcast par lignes ~2×), algèbre linéaire
(`Matmul`/`MatVec`/`T`) et paquet de ML classique (`ml` : régression linéaire,
standardisation).

### Ajouté (Sprint 23 — ML classique : paquet `ml`)

- **Nouveau paquet `ml`** (sur `ndarray`) :
  - `Standardize` : centrage-réduction par colonne (moyenne/écart-type).
  - `LinearRegression` : régression linéaire par descente de gradient
    (`Fit`/`Predict`), `MSE`.
- Test de convergence : apprend `y = 2·x₁ + 3·x₂ + 1` (poids/biais retrouvés,
  MSE ≈ 0).
- Portée assumée : ML classique/pédagogique (pas d'autograd ni de GPU ; Go sert
  surtout à l'inférence en production).

### Ajouté (Sprint 22 — Algèbre linéaire `ndarray`)

- **`Matmul`** (produit matriciel 2D, ordre ikj cache-friendly), **`MatVec`**
  (matrice × vecteur), **`T`** (transposée 2D). Base pour le ML classique.
- Perf documentée : `Matmul` naïf ~2,5× plus lent que NumPy/BLAS (256×256) — un
  BLAS serait requis pour de grosses matrices (voir `docs/NDARRAY.md`).

### Performances (Sprint 21 — Broadcast par lignes, itération strided optimisée)

- **Réécriture du broadcast float64 en itération par lignes** : boucle interne
  **contiguë** sur le dernier axe, avec scalaire hissé quand un stride interne est
  nul (cas typique). Noyaux contigus spécialisés `fillScalarVec`/`fillVecScalar`/
  `fillVecVec` ; dispatcher `parallelLines` (parallélise selon le volume total,
  même quand les lignes sont peu nombreuses mais longues).
- **Gain net** : `Broadcast` 1 M passe de 2650 µs à **1350 µs** (~2×) ; écart avec
  NumPy réduit de 2,5× à **1,6×**. Le vrai goulot était la structure d'itération,
  pas la closure — confirmé.

### Performances (Sprint 20 — Broadcasting float64 spécialisé)

- **`broadcastFloat64`** : broadcasting par nom spécialisé float64 (sélection de
  l'opération par switch, sans closure), branché sur `Add`/`Sub`/`Mul`/`Div`.
- Refactorisation utile : `broadcastLayout` (préparation du layout de
  broadcasting) et `parallelFill` (dispatcher de parallélisation) extraits et
  mutualisés.
- **Résultat honnête** : gain **marginal** (~10 % sur 1 M, nul sur 40 k). Pour le
  broadcast, le goulot est l'itération strided et les accès mémoire non contigus,
  pas la closure — enseignement documenté dans `docs/BENCHMARKS.md`.

### Performances (Sprint 19 — Noyaux directs float64 sur toute l'arithmétique)

- Le chemin rapide float64 sans closure (auparavant limité à `Add`) est **branché
  sur `Sub`, `Mul`, `Div`** via `binaryFloat64Fast(kernel)` mutualisé, avec les
  noyaux `subFloat64`/`mulFloat64`/`divFloat64`.
- `Mul`/`Sub`/`Div` sur 1 M (mêmes coordonnées) : ~3218 µs (closure générique) →
  **~1545 µs** (2,1×). Équivalence avec le chemin générique vérifiée par test.

## [0.3.0] — 2026-08-06

Troisième version : jointures externes, `concat`/`stack`, `groupby`
(DataArray et Dataset), `skipna`, moteur de calcul `ndarray` (avec API in-place
rejoignant NumPy), volet performance honnête (comparaison NumPy pur / xarray,
expérience SIMD et cgo).

### Ajouté (Sprint 17 — `skipna`)

- **Réductions ignorant les NaN** : globales (`SumSkipNA`, `MeanSkipNA`,
  `MinSkipNA`, `MaxSkipNA`) et par axe (`SumAxisSkipNA`, `MeanAxisSkipNA`,
  `MinAxisSkipNA`, `MaxAxisSkipNA`), conformes au comportement par défaut de
  xarray. Sans effet pour les types entiers (aucun NaN possible).

### Ajouté (Sprint 18 — `Dataset.GroupBy`)

- **`Dataset.GroupBy(dim)`** : regroupement propagé à toutes les variables
  portant la dimension, avec agrégations `Sum`, `Mean`, `Min`, `Max` ; les
  variables sans la dimension sont conservées.
- Refactorisation : `groupReduceOn` (réduction de groupe découplée de la
  coordonnée propre) réutilisée par `DataArray.GroupBy` et `Dataset.GroupBy`.

### Performances (Sprint 16 — API in-place `ndarray`)

- **Opérations in-place** (`AddInto`, `SubInto`, `MulInto`, `DivInto`,
  `AddInPlace`, `EmptyLike`) : le résultat est écrit dans un `dst` fourni →
  **zéro allocation**.
- Résout le vrai goulot identifié au Sprint 15/cgo (l'allocation, pas le calcul) :
  `Add` 1000×1000 passe de 1632 µs (8 Mo alloués) à **856 µs (0 alloc)**, soit
  **1,17× de NumPy pur** (733 µs) au lieu de 2,2×. Obtenu en **Go pur**, sans cgo.
- Tests : correctness des `*Into`, immutabilité des opérandes, accumulation,
  forme de destination invalide.

### Ajouté (Sprint 15 — Paquet `ndarray`, moteur « mini-NumPy »)

- **Nouveau paquet `ndarray`** : tableau dense N-D `float64` spécialisé (sans
  générique/closure sur le chemin chaud), avec broadcasting **positionnel façon
  NumPy** (aligné à droite).
  - Construction (`New`, `Zeros`, `Arange`), accès (`At`, `Shape`, `Data`…) ;
  - arithmétique `Add`/`Sub`/`Mul`/`Div` (même forme + broadcasting), scalaires ;
  - réductions `Sum`, `Mean`, `SumAxis`, `MeanAxis` ; tests complets.
- **Conclusion mesurée et documentée** (`docs/NDARRAY.md`) : même un moteur Go nu
  reste 2,3×–4,5× plus lent que NumPy pur. Rattraper NumPy exige C+SIMD+BLAS ;
  ce n'est pas atteignable en Go idiomatique. Le paquet a une valeur
  architecturale (moteur propre, socle potentiel de `Variable[float64]`), pas de
  supériorité de débit.

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

- Initialisation du projet : module Go `github.com/benedictemarty/xarray`, dépôt git,
  cadre de gestion agile.
