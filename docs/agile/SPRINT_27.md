# Sprint 27 — GRIB2 complex packing + différenciation spatiale

- **Période** : démarrage 2026-08-07.
- **Objectif** : décoder le complex packing GRIB2 (templates 5.2/5.3), laissé de
  côté au Sprint 26.

## Périmètre

| Sujet | État |
|-------|------|
| Complex packing (5.2) et complex + différenciation spatiale (5.3) | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./...` passe.
- [x] **Validation contre ecCodes** (diff max = 0,0 sur 26 331 valeurs réelles).
- [x] Test unitaire versionné (`testdata/complex_synth.grib2`).
- [x] `docs/GRIB.md` / `CHANGELOG.md` à jour ; commit atomique.

## Démarche

Au Sprint 26, la tentative de complex packing échouait : les paramètres se
décodaient bien mais le flux de bits des données était désaligné, et je l'avais
retiré plutôt que livrer du faux. **Recherche documentaire** (code de référence
g2clib `comunpack`) : le point manquant était l'**alignement à l'octet après
chaque bloc** (références, largeurs, longueurs de groupe), et la lecture des
longueurs pour **tous** les groupes (dernière remplacée par « true length »).

Après correction, le décodage correspond **exactement** à ecCodes.

## Décisions de conception

- Algorithme aligné sur g2clib : blocs octet-alignés, formules
  `width = ref + lu`, `length = lu·incrément + ref` (dernier = trueLengthLastGroup),
  réversion spatiale d'ordre 1 (`v[i] += omin + v[i-1]`) et 2
  (`v[i] += omin + 2·v[i-1] − v[i-2]`).
- Templates **locaux** (50002…) toujours refusés explicitement → ecCodes.

## Rétrospective

- **Leçon** : ne pas livrer de code non validé (Sprint 26), puis **se documenter**
  sur la spec de référence a permis de résoudre proprement. La validation contre
  ecCodes sur données réelles est la garantie décisive.
