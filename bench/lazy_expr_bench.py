#!/usr/bin/env python3
"""Génère deux stores Zarr et mesure l'expression composée mean(a*b) via dask.

Usage : python3 bench/lazy_expr_bench.py /tmp/a.zarr /tmp/b.zarr 8000 2000 800
"""

import sys
import time
import numpy as np
import zarr
import numcodecs
import dask.array as darr


def genstore(path, arr, chunkRows):
    N, M = arr.shape
    try:
        z = zarr.create(shape=(N, M), chunks=(chunkRows, M), dtype="<f8",
                        store=path, zarr_format=2,
                        compressor=numcodecs.Zlib(level=1), overwrite=True)
    except TypeError:
        z = zarr.create(shape=(N, M), chunks=(chunkRows, M), dtype="<f8",
                        store=path, zarr_format=2,
                        compressors=numcodecs.Zlib(level=1), overwrite=True)
    z[:] = arr
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
    return dt / n * 1000


def main():
    pa, pb = sys.argv[1], sys.argv[2]
    N = int(sys.argv[3]) if len(sys.argv) > 3 else 8000
    M = int(sys.argv[4]) if len(sys.argv) > 4 else 2000
    cr = int(sys.argv[5]) if len(sys.argv) > 5 else 800

    a = np.arange(N * M, dtype="<f8").reshape(N, M)
    b = (a * 0.5).astype("<f8")
    za = genstore(pa, a, cr)
    zb = genstore(pb, b, cr)
    print(f"stores {N}x{M} chunks {cr}x{M} zlib")

    da_a = darr.from_zarr(za).rechunk((cr, M))
    da_b = darr.from_zarr(zb).rechunk((cr, M))

    def run():
        return float((da_a * da_b).mean().compute())

    ms = mesure(run)
    print(f"dask mean(a*b) : {ms:.2f} ms (résultat={run():.4f})")


if __name__ == "__main__":
    main()
