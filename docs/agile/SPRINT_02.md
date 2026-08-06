# Sprint 2 — Opérations

- **Période** : démarrage 2026-08-06.
- **Objectif de sprint** : rendre les `DataArray` « calculables » — réductions
  par axe, arithmétique avec broadcasting et alignement.

## Périmètre engagé

| ID    | User story | État |
|-------|------------|------|
| US-06 | Réductions par axe (`SumAxis`, `MeanAxis`, `MinAxis`, `MaxAxis`) | ✅ |
| US-07 | Arithmétique élément par élément (`Add`, `Sub`, `Mul`, `Div`) | ✅ |
| US-08 | Broadcasting par nom de dimension | ✅ |
| US-09 | Alignement automatique sur les coordonnées | ✅ |
| US-10 | `Transpose` | ✅ |

## Critères d'acceptation (Definition of Done)

- [x] `gofmt`, `go vet` propres.
- [x] `go test ./...` : tous les tests passent.
- [x] Broadcasting et alignement couverts par des tests (cas nominaux + erreurs).
- [x] `CHANGELOG.md`, backlog et roadmap à jour.
- [x] Commits atomiques.

## Décisions de conception

- **Broadcasting par nom, pas par position** : c'est la sémantique de xarray et
  ce qui la distingue de NumPy. Deux tableaux de dimensions `{x}` et `{y}` se
  combinent en `{x, y}` ; les dimensions communes doivent avoir la même taille.
- **Alignement avant opération** : lorsqu'une dimension porte des coordonnées des
  deux côtés, seules les étiquettes communes sont conservées (jointure interne),
  dans l'ordre du premier opérande. Sans étiquette commune → erreur explicite.
- **Ordre des dimensions du résultat** : dimensions de l'opérande gauche, puis
  celles de l'opérande droit absentes à gauche (déterministe).

## Rétrospective

- **Bien** : le moteur générique `binaryOp` (compteur multi-dimensionnel + strides
  par nom) factorise tout le broadcasting ; `reduceAxis` est réutilisé par les
  quatre réductions.
- **À surveiller** : uniquement la jointure **interne** est gérée (pas encore
  `outer`/`left`) ; performances non profilées (dette T-03).
- **Prochain sprint** : `Dataset` (US-11 à US-13).
