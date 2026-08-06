#!/usr/bin/env python3
"""Benchmarks xarray (Python) miroir des benchmarks Go de xarray-go.

Mesure le temps moyen par opération (en nanosecondes) pour des opérations
équivalentes et des tailles identiques à celles de bench_test.go, afin de
produire une comparaison honnête xarray-go vs xarray (Python/NumPy).

Usage : python3 bench/xr_bench.py
"""

import time
import numpy as np
import xarray as xr


def mesure(fn, cible_s=0.5, warmup=3):
    """Renvoie le temps moyen par appel de fn, en nanosecondes."""
    for _ in range(warmup):
        fn()
    # Calibrage du nombre d'itérations pour atteindre ~cible_s.
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


def grille(n):
    data = np.arange(n * n, dtype=np.float64).reshape(n, n)
    return xr.DataArray(
        data, dims=("x", "y"),
        coords={"x": np.arange(n, dtype=np.float64),
                "y": np.arange(n, dtype=np.float64)},
        name="grille",
    )


def bench_add():
    a = grille(100)
    c = grille(100)
    return mesure(lambda: (a + c))


def bench_broadcast():
    x = xr.DataArray(np.zeros(200), dims=("x",))
    y = xr.DataArray(np.zeros(200), dims=("y",))
    return mesure(lambda: (x + y))


def bench_broadcast_large():
    x = xr.DataArray(np.zeros(1000), dims=("x",))
    y = xr.DataArray(np.zeros(1000), dims=("y",))
    return mesure(lambda: (x + y))


def bench_sumaxis():
    a = grille(100)
    return mesure(lambda: a.sum(dim="x"))


def bench_meanaxis():
    a = grille(100)
    return mesure(lambda: a.mean(dim="x"))


def bench_groupby_sum():
    n, groupes = 1000, 10
    data = np.arange(n * 10, dtype=np.float64).reshape(n, 10)
    t = (np.arange(n) % groupes).astype(np.float64)
    a = xr.DataArray(
        data, dims=("t", "x"),
        coords={"t": t, "x": np.arange(10, dtype=np.float64)}, name="v",
    )
    return mesure(lambda: a.groupby("t").sum())


BENCHES = {
    "Add": bench_add,
    "Broadcast": bench_broadcast,
    "BroadcastLarge": bench_broadcast_large,
    "SumAxis": bench_sumaxis,
    "MeanAxis": bench_meanaxis,
    "GroupBySum": bench_groupby_sum,
}


def main():
    print(f"# xarray {xr.__version__} | numpy {np.__version__}")
    print(f"{'benchmark':<14}{'ns/op':>16}")
    for nom, fn in BENCHES.items():
        ns = fn()
        print(f"{nom:<14}{ns:>16.0f}")


if __name__ == "__main__":
    main()
