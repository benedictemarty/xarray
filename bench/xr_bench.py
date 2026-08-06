#!/usr/bin/env python3
"""Benchmarks de référence pour la comparaison avec xarray-go.

Mesure, pour des opérations équivalentes, DEUX références Python :
  - NumPy **pur** (le moteur de calcul, C + SIMD) ;
  - **xarray** (la couche étiquetée au-dessus de NumPy).

Distinguer les deux est essentiel : comparer xarray-go à xarray flatte
artificiellement Go (overhead d'objets Python de xarray) ; la comparaison juste
du calcul brut est celle contre NumPy pur.

Usage : python3 bench/xr_bench.py
"""

import time
import numpy as np
import xarray as xr


def mesure(fn, cible_s=0.5, warmup=3):
    for _ in range(warmup):
        fn()
    n = 1
    while True:
        t0 = time.perf_counter()
        for _ in range(n):
            fn()
        dt = time.perf_counter() - t0
        if dt >= cible_s or n > 5_000_000:
            break
        n *= 2
    return dt / n * 1e9  # ns/op


def grille_np(n):
    return np.arange(n * n, dtype=np.float64).reshape(n, n)


def grille_xr(n):
    return xr.DataArray(grille_np(n), dims=("x", "y"),
                        coords={"x": np.arange(n, dtype=np.float64),
                                "y": np.arange(n, dtype=np.float64)})


# Chaque entrée : (numpy_pur, xarray). None si non applicable.
def benches():
    a100n, b100n = grille_np(100), grille_np(100)
    a100x, b100x = grille_xr(100), grille_xr(100)
    a1kn, b1kn = grille_np(1000), grille_np(1000)
    a1kx, b1kx = grille_xr(1000), grille_xr(1000)

    xn = np.array([1.0, 2.0, 3.0] * 66 + [1.0, 2.0])  # 200
    yn = np.arange(200, dtype=np.float64)
    xx = xr.DataArray(xn, dims=("x",))
    yx = xr.DataArray(yn, dims=("y",))

    xl = np.zeros(1000)
    yl = np.zeros(1000)
    xlx = xr.DataArray(xl, dims=("x",))
    ylx = xr.DataArray(yl, dims=("y",))

    gd = np.array([10.0, 20.0, 30.0, 40.0, 50.0])
    starts = [0, 2]  # groupes contigus [1,1] puis [2,2,2]
    gx = xr.DataArray(gd, dims=("t",), coords={"t": [1.0, 1.0, 2.0, 2.0, 2.0]})

    return {
        "Add":        (lambda: a100n + b100n,           lambda: a100x + b100x),
        "AddLarge":   (lambda: a1kn + b1kn,             lambda: a1kx + b1kx),
        "Broadcast":  (lambda: xn[:, None] + yn[None],  lambda: xx + yx),
        "BroadcastL": (lambda: xl[:, None] + yl[None],  lambda: xlx + ylx),
        "SumAxis":    (lambda: a100n.sum(axis=0),       lambda: a100x.sum(dim="x")),
        "MeanAxis":   (lambda: a100n.mean(axis=0),      lambda: a100x.mean(dim="x")),
        "GroupBySum": (lambda: np.add.reduceat(gd, starts), lambda: gx.groupby("t").sum()),
    }


def main():
    print(f"# numpy {np.__version__} | xarray {xr.__version__}")
    print(f"{'benchmark':<12}{'numpy_pur(ns)':>16}{'xarray(ns)':>16}")
    for nom, (fnp, fxr) in benches().items():
        np_ns = mesure(fnp) if fnp else float("nan")
        xr_ns = mesure(fxr) if fxr else float("nan")
        print(f"{nom:<12}{np_ns:>16.0f}{xr_ns:>16.0f}")


if __name__ == "__main__":
    main()
