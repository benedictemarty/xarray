# Sprint 4 — Entrées / sorties

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : permettre la persistance et l'échange des données via
  des formats de fichiers courants.

## Périmètre engagé

| ID    | User story | État |
|-------|------------|------|
| US-14 | Charger/sauver en CSV | ✅ |
| US-15 | Charger/sauver en JSON | ✅ |
| US-16 | Lire/écrire du netCDF | ⏸️ Reporté |

## Critères d'acceptation (Definition of Done)

- [x] `gofmt`, `go vet` propres.
- [x] `go test ./...` : tous les tests passent.
- [x] Allers-retours JSON et CSV testés (contenu exact, cas d'erreur).
- [x] `CHANGELOG.md`, backlog et roadmap à jour.
- [x] Commit atomique.

## Décisions de conception

- **JSON** : forme sérialisable explicite (`dims`, `shape`, `data`, `coords`,
  `name`). La relecture passe par `NewDataArray`/`NewDataset`, donc toute donnée
  invalide est rejetée. Pour le `Dataset`, les coordonnées partagées sont
  réinjectées dans chaque variable selon ses dimensions.
- **CSV « tidy »** : une ligne par cellule (une colonne par dimension + la
  valeur). Choisi plutôt qu'un tableau 2D « large » car :
  - il gère un nombre **quelconque** de dimensions sans ambiguïté ;
  - l'en-tête porte les **noms** des dimensions et de la variable ;
  - la relecture reconstruit les coordonnées et détecte les **grilles
    incomplètes**.
- **netCDF reporté** : format binaire (basé HDF5/format classique) dont une
  implémentation correcte exigerait une dépendance externe. Plutôt que de livrer
  un sous-ensemble fragile, la story reste au backlog (P2). *Je préfère ne pas
  inventer un format que je ne maîtrise pas entièrement.*

## Rétrospective

- **Bien** : le format tidy s'est révélé simple et robuste ; réutilisation des
  constructeurs pour la validation à la lecture.
- **À surveiller** : le CSV impose des étiquettes numériques (cohérent avec le
  choix « float64 partout » — dette T-01) ; pas encore de lecture/écriture vers
  fichiers nommés (seulement `io.Reader`/`io.Writer`, ce qui est plus flexible).
- **Bilan de la roadmap initiale** : les épopées 1 à 4 sont couvertes (hors
  netCDF). Les évolutions futures relèvent surtout de la dette identifiée
  (generics, performances, jointures externes).
