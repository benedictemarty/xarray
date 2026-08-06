# Sprint 16 — API in-place `ndarray`

- **Période** : démarrage 2026-08-06.
- **Objectif** : supprimer le coût d'allocation, identifié comme le vrai goulot
  (démo cgo).

## Périmètre

| Sujet | État |
|-------|------|
| `AddInto`/`SubInto`/`MulInto`/`DivInto`/`AddInPlace`/`EmptyLike` | ✅ |

## Definition of Done

- [x] `gofmt`, `go vet` propres ; `go test ./ndarray/` passe.
- [x] Gain mesuré (0 allocation) ; `CHANGELOG.md`/`docs/NDARRAY.md` à jour.
- [x] Commit atomique.

## Décision de conception

Le résultat est écrit dans un `dst` fourni par l'appelant → **zéro allocation**.
En réutilisant un buffer (ou via pool), on atteint quasiment le débit NumPy sur
le memory-bound.

## Résultat

`Add` 1 M : **1632 µs (8 Mo alloués) → 856 µs (0 alloc)** = **1,17× de NumPy pur**
(733 µs) au lieu de 2,2×. En Go pur, sans cgo.

## Rétrospective

- **Bien** : confirme que l'allocation, pas le calcul, était le goulot.
- **Suite** : `skipna` (Sprint 17).
