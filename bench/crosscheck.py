#!/usr/bin/env python3
"""Vérification d'équivalence : calcule des opérations avec xarray (Python) et
écrit les résultats dans bench/expected.json.

Le test Go `crosscheck_test.go` recalcule les mêmes opérations et compare à ce
fichier, garantissant que xarray-go produit des valeurs identiques à xarray.

Usage : python3 bench/crosscheck.py
"""

import json
import os
import numpy as np
import xarray as xr


def flat(da):
    """Renvoie les données aplaties en ordre C (comme xarray-go)."""
    return np.asarray(da.values, dtype=np.float64).reshape(-1).tolist()


def main():
    results = {}

    # add : deux tableaux 2D de mêmes coordonnées, addition élément par élément.
    a = xr.DataArray(np.arange(6, dtype=np.float64).reshape(2, 3),
                     dims=("x", "y"),
                     coords={"x": [0.0, 1.0], "y": [10.0, 20.0, 30.0]})
    b = xr.DataArray(np.arange(6, 12, dtype=np.float64).reshape(2, 3),
                     dims=("x", "y"),
                     coords={"x": [0.0, 1.0], "y": [10.0, 20.0, 30.0]})
    results["add"] = flat(a + b)

    # broadcast : x(3) + y(2) -> (3,2) par nom de dimension.
    x = xr.DataArray(np.array([1.0, 2.0, 3.0]), dims=("x",))
    y = xr.DataArray(np.array([10.0, 20.0]), dims=("y",))
    results["broadcast"] = flat(x + y)

    # sum_axis / mean_axis le long de x.
    results["sum_axis_x"] = flat(a.sum(dim="x"))
    results["mean_axis_x"] = flat(a.mean(dim="x"))

    # outer join : a2 sur [0,1,2], b2 sur [1,2,3], remplissage 0.
    a2 = xr.DataArray(np.array([10.0, 20.0, 30.0]), dims=("k",),
                      coords={"k": [0.0, 1.0, 2.0]})
    b2 = xr.DataArray(np.array([1.0, 2.0, 3.0]), dims=("k",),
                      coords={"k": [1.0, 2.0, 3.0]})
    joined = a2 + b2.reindex(k=[0.0, 1.0, 2.0, 3.0], fill_value=0.0)
    joined = a2.reindex(k=[0.0, 1.0, 2.0, 3.0], fill_value=0.0) + \
        b2.reindex(k=[0.0, 1.0, 2.0, 3.0], fill_value=0.0)
    results["outer_join"] = flat(joined)

    # groupby sum : coordonnée t répétée.
    g = xr.DataArray(np.array([10.0, 20.0, 30.0, 40.0, 50.0]), dims=("t",),
                     coords={"t": [1.0, 1.0, 2.0, 2.0, 2.0]})
    results["groupby_sum"] = flat(g.groupby("t").sum())

    out = os.path.join(os.path.dirname(__file__), "expected.json")
    with open(out, "w") as f:
        json.dump(results, f, indent=2)
    print(f"écrit {out} (xarray {xr.__version__})")
    for k, v in results.items():
        print(f"  {k}: {v}")


if __name__ == "__main__":
    main()
