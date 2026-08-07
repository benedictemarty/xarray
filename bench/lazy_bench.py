#!/usr/bin/env python3
"""Génère un store Zarr v2 et mesure la moyenne out-of-core via dask.

Le même store est ensuite lu par xarray-go (cmd/benchzarr) pour comparer les deux
approches lazy/hors-mémoire.

Usage : python3 bench/lazy_bench.py /tmp/lazy.zarr 4000 1000 400
"""

import sys
import time
import numpy as np
import zarr
import numcodecs
import dask.array as darr


def genstore(path, N, M, chunkRows):
    a = np.arange(N * M, dtype="<f8").reshape(N, M)
    try:
        z = zarr.create(shape=(N, M), chunks=(chunkRows, M), dtype="<f8",
                        store=path, zarr_format=2,
                        compressor=numcodecs.Zlib(level=1), overwrite=True)
    except TypeError:
        z = zarr.create(shape=(N, M), chunks=(chunkRows, M), dtype="<f8",
                        store=path, zarr_format=2,
                        compressors=numcodecs.Zlib(level=1), overwrite=True)
    z[:] = a
    return z


def mesure(fn, cible=0.5, warmup=2):
    for _ in range(warmup):
        fn()
    n = 1
    while True:
        t0 = time.perf_counter()
        for _ in range(n):
            fn()
        dt = time.perf_counter() - t0
        if dt >= cible or n > 100000:
            break
        n *= 2
    return dt / n * 1000  # ms


def main():
    path = sys.argv[1] if len(sys.argv) > 1 else "/tmp/lazy.zarr"
    N = int(sys.argv[2]) if len(sys.argv) > 2 else 4000
    M = int(sys.argv[3]) if len(sys.argv) > 3 else 1000
    chunkRows = int(sys.argv[4]) if len(sys.argv) > 4 else 400

    z = genstore(path, N, M, chunkRows)
    print(f"store {N}x{M} chunks {chunkRows}x{M} zlib -> {path}")

    d = darr.from_zarr(z).rechunk((chunkRows, M))

    def run():
        return float(d.mean().compute())

    ms = mesure(run)
    print(f"dask from_zarr.mean().compute() : {ms:.2f} ms (moyenne={run():.4f})")


if __name__ == "__main__":
    main()
