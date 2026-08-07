# Sprint 28 — Backend GRIB via ecCodes (cas particuliers / templates locaux)

- **Période** : démarrage 2026-08-07.
- **Objectif** : lire les fichiers GRIB que le décodeur pur-Go ne couvre pas —
  en particulier le template **local 50002 de Météo-France** — en s'appuyant sur
  ecCodes.

## Périmètre

| Sujet | État |
|-------|------|
| Backend cgo ecCodes (`ReadFile`), opt-in `-tags eccodes` | ✅ |
| Lecture des templates locaux (50002) | ✅ (via ecCodes) |
| Cœur du projet inchangé (100 % Go pur par défaut) | ✅ |

## Definition of Done

- [x] `go build ./...` et `go test ./...` **sans** cgo : cœur intact.
- [x] Backend compile avec `-tags eccodes` (libeccodes + eckit).
- [x] **Validé** contre ecCodes Python sur un vrai fichier 50002 (diff = 0,0).
- [x] `docs/GRIB.md`, README du backend, `CHANGELOG.md` à jour ; commit atomique.

## Différence 50002 vs standard

- **5.2/5.3** : templates **WMO standard**, documentés → décodables à la main.
- **50002** : template **local Météo-France** (≥ 50000 = plage centres), portage
  du « second-order packing » historique. Format binaire **non publié** dans la
  spec WMO → seul ecCodes (qui embarque les définitions des centres) le décode
  correctement. Le réimplémenter en Go reviendrait à l'inventer.

## Décisions de conception

- **Stratégie hybride** : décodeur pur-Go par défaut (portable, sans dépendance) ;
  backend ecCodes en secours pour les templates locaux / formats exotiques.
- Isolation par **build tag** + fichier stub → le cœur reste 100 % Go pur et
  cross-compilable ; le backend cgo n'est compilé que sur demande.
- Chemins cgo spécifiques à la machine (installation `eccodeslib`/`eckitlib`),
  documentés comme à adapter (idéalement `pkg-config`).

## Rétrospective

- **Leçon** : pour un cas particulier propriétaire non documenté, la bonne
  réponse n'est pas de le réimplémenter (on l'inventerait) mais de **déléguer à
  l'outil de référence** — tout en gardant un cœur pur-Go pour le cas général.
