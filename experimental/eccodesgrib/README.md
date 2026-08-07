# experimental/eccodesgrib — backend GRIB via ecCodes (cgo)

Backend GRIB s'appuyant sur la bibliothèque C **ecCodes** (ECMWF) via cgo, pour
gérer **tout** le format GRIB, y compris les cas que le décodeur pur-Go du paquet
racine ne couvre pas.

## Deux backends complémentaires

| | Décodeur pur-Go (`xarray.ReadGrib`) | Backend ecCodes (`eccodesgrib.ReadFile`) |
|---|---|---|
| Dépendance | **aucune** (100 % Go) | libeccodes (C) + cgo |
| Portabilité | binaire statique, cross-compile | dépend de la lib système |
| Simple packing (5.0) | ✅ | ✅ |
| Complex packing standard (5.2/5.3) | ✅ | ✅ |
| **Templates locaux (ex. 50002 Météo-France)** | ❌ | ✅ |
| GRIB1, grilles gaussiennes, bitmaps, JPEG2000… | ❌ | ✅ |

**Règle** : utiliser le décodeur pur-Go par défaut (portable, sans dépendance).
Basculer sur le backend ecCodes uniquement pour les fichiers qu'il ne sait pas
lire (templates locaux, formats exotiques).

## Pourquoi ecCodes gère les cas particuliers « gratuitement »

ecCodes embarque les **fichiers de définition** de chaque centre météo (dont
Météo-France). Décoder un template local ne demande donc **aucun code spécifique** :
on appelle `codes_get_values` et ecCodes applique la bonne définition. À l'inverse,
réimplémenter un template local en Go supposerait de connaître son format binaire
exact — non publié dans la spec WMO.

## Compilation et exécution

Le backend est **opt-in** (build tag `eccodes`) ; sans lui, le paquet fournit un
stub renvoyant une erreur, et le cœur du projet reste 100 % Go pur.

```bash
# Adapter les chemins cgo (CFLAGS/LDFLAGS) dans grib_eccodes.go à votre
# installation de libeccodes, puis :
go build -tags eccodes ./cmd/readgrib-ec

# À l'exécution, la lib doit être trouvable :
LD_LIBRARY_PATH=/chemin/eccodeslib/lib64:/chemin/eckitlib/lib64 \
    ./readgrib-ec fichier.grib
```

## Validation

Lecture d'un fichier **Météo-France en template local 50002** (que le décodeur
pur-Go refuse) : les valeurs obtenues sont **identiques** à celles d'ecCodes en
Python (diff max = 0,0 sur 26 331 points).

## Limites

- Les chemins cgo sont **spécifiques à la machine** (ici, l'installation
  `eccodeslib`/`eckitlib` du paquet Python) — à adapter (idéalement via
  `pkg-config eccodes` quand disponible).
- Perd les avantages du pur-Go : compilateur C requis, binaire non statique,
  cross-compilation compliquée.
