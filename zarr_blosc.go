package xarray

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

// Décompression des chunks Zarr. Prend en charge, en pur Go et sans dépendance :
//   - aucune compression ;
//   - zlib ;
//   - Blosc (conteneur v1) avec codec LZ4 ou données non compressées (memcpy),
//     filtre byte-shuffle. Le bitshuffle et les autres codecs (zstd…) ne sont
//     pas gérés. C'est le compresseur par défaut de zarr-python (Blosc/LZ4).
//
// La correction des données est validée par aller-retour contre des stores Zarr
// réels produits par zarr-python (voir tests).

// decompressor décompresse les octets bruts d'un chunk vers les octets f8.
type decompressor func([]byte) ([]byte, error)

// newDecompressor construit le décompresseur adapté aux métadonnées .zarray.
// Renvoie nil si aucune compression.
func newDecompressor(m *zarrCompressorMeta) (decompressor, error) {
	if m == nil {
		return nil, nil
	}
	switch m.ID {
	case "zlib":
		return zlibDecompress, nil
	case "blosc":
		if m.Shuffle == 2 {
			return nil, fmt.Errorf("xarray: Blosc bitshuffle (shuffle=2) non pris en charge")
		}
		if m.Cname != "" && m.Cname != "lz4" && m.Cname != "blosclz" {
			return nil, fmt.Errorf("xarray: codec Blosc %q non pris en charge (lz4/blosclz)", m.Cname)
		}
		return bloscDecompress, nil
	default:
		return nil, fmt.Errorf("xarray: compresseur %q non pris en charge (aucun, zlib, blosc)", m.ID)
	}
}

func zlibDecompress(src []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

// bloscDecompress décode un buffer Blosc (conteneur v1). En-tête (16 octets) :
// version, versionlz, flags, typesize, nbytes(u32), blocksize(u32), cbytes(u32).
func bloscDecompress(src []byte) ([]byte, error) {
	if len(src) < 16 {
		return nil, fmt.Errorf("xarray: en-tête Blosc tronqué (%d octets)", len(src))
	}
	flags := src[2]
	typesize := int(src[3])
	nbytes := int(binary.LittleEndian.Uint32(src[4:8]))
	blocksize := int(binary.LittleEndian.Uint32(src[8:12]))
	doShuffle := flags&0x01 != 0
	memcpyed := flags&0x02 != 0
	bitshuffle := flags&0x04 != 0
	if bitshuffle {
		return nil, fmt.Errorf("xarray: Blosc bitshuffle non pris en charge")
	}

	// Cas memcpy : le buffer contient les données ORIGINALES (non filtrées, non
	// compressées) après l'en-tête. Pas d'unshuffle dans ce cas.
	if memcpyed {
		if len(src) < 16+nbytes {
			return nil, fmt.Errorf("xarray: bloc Blosc memcpy tronqué")
		}
		out := make([]byte, nbytes)
		copy(out, src[16:16+nbytes])
		return out, nil
	}

	if blocksize <= 0 {
		return nil, fmt.Errorf("xarray: blocksize Blosc invalide")
	}
	out := make([]byte, nbytes)
	nblocks := (nbytes + blocksize - 1) / blocksize
	if len(src) < 16+4*nblocks {
		return nil, fmt.Errorf("xarray: table d'offsets Blosc tronquée")
	}
	for b := 0; b < nblocks; b++ {
		off := int(int32(binary.LittleEndian.Uint32(src[16+4*b : 20+4*b])))
		if off < 0 || off > len(src) {
			return nil, fmt.Errorf("xarray: offset de bloc Blosc invalide")
		}
		blkLen := blocksize
		if b == nblocks-1 {
			blkLen = nbytes - b*blocksize
		}
		if err := decodeBloscBlock(src[off:], out[b*blocksize:b*blocksize+blkLen], typesize); err != nil {
			return nil, err
		}
	}

	if doShuffle && typesize > 1 {
		out = unshuffle(out, typesize)
	}
	return out, nil
}

// decodeBloscBlock décode un bloc dans dst (dimensionné à sa taille
// décompressée). Un bloc est fait de `nstreams` sous-flux consécutifs, chacun
// préfixé de sa taille compressée (int32) puis des données codec (LZ4). Blosc
// découpe en `typesize` sous-flux quand la taille le permet, sinon 1.
func decodeBloscBlock(src, dst []byte, typesize int) error {
	blkLen := len(dst)
	// Règle de découpage de Blosc (BLOSC_MIN_BUFFERSIZE = 128) : un bloc est
	// découpé en `typesize` sous-flux (codec LZ4/BLOSCLZ) si typesize ∈ [2,16] et
	// blkLen/typesize ≥ 128 ; sinon un seul flux. Vérifié sur des stores réels
	// (neblock 100 → non découpé, 200/20000 → découpé).
	const bloscMinBuffer = 128
	nstreams := 1
	if typesize >= 2 && typesize <= 16 && blkLen%typesize == 0 && blkLen/typesize >= bloscMinBuffer {
		nstreams = typesize
	}
	neblock := blkLen / nstreams
	sp := 0
	for s := 0; s < nstreams; s++ {
		if sp+4 > len(src) {
			return fmt.Errorf("xarray: préfixe de sous-flux Blosc tronqué")
		}
		clen := int(int32(binary.LittleEndian.Uint32(src[sp : sp+4])))
		sp += 4
		if clen < 0 || sp+clen > len(src) {
			return fmt.Errorf("xarray: sous-flux Blosc %d hors bornes (clen=%d)", s, clen)
		}
		out := dst[s*neblock : (s+1)*neblock]
		if clen == neblock {
			// Sous-flux incompressible : stocké brut par Blosc.
			copy(out, src[sp:sp+clen])
		} else if err := lz4Decompress(src[sp:sp+clen], out); err != nil {
			return fmt.Errorf("xarray: LZ4 (sous-flux %d): %w", s, err)
		}
		sp += clen
	}
	return nil
}

// lz4Decompress décode un bloc LZ4 (format « block », sans en-tête de frame)
// depuis src (exactement la taille compressée) jusqu'à remplir dst.
func lz4Decompress(src, dst []byte) error {
	sp, dp := 0, 0
	n := len(dst)
	for dp < n {
		if sp >= len(src) {
			return fmt.Errorf("entrée LZ4 tronquée (token)")
		}
		token := src[sp]
		sp++
		litLen := int(token >> 4)
		if litLen == 15 {
			for {
				if sp >= len(src) {
					return fmt.Errorf("entrée LZ4 tronquée (litLen)")
				}
				b := src[sp]
				sp++
				litLen += int(b)
				if b != 255 {
					break
				}
			}
		}
		if sp+litLen > len(src) || dp+litLen > n {
			return fmt.Errorf("littéraux LZ4 hors bornes")
		}
		copy(dst[dp:dp+litLen], src[sp:sp+litLen])
		sp += litLen
		dp += litLen
		if dp == n {
			break // dernière séquence : littéraux seuls
		}
		if sp+2 > len(src) {
			return fmt.Errorf("entrée LZ4 tronquée (offset)")
		}
		offset := int(src[sp]) | int(src[sp+1])<<8
		sp += 2
		if offset == 0 || offset > dp {
			return fmt.Errorf("offset LZ4 invalide (%d)", offset)
		}
		matchLen := int(token & 0x0f)
		if matchLen == 15 {
			for {
				if sp >= len(src) {
					return fmt.Errorf("entrée LZ4 tronquée (matchLen)")
				}
				b := src[sp]
				sp++
				matchLen += int(b)
				if b != 255 {
					break
				}
			}
		}
		matchLen += 4 // minmatch
		if dp+matchLen > n {
			return fmt.Errorf("match LZ4 hors bornes")
		}
		// Copie octet par octet (les matchs peuvent se chevaucher).
		for i := 0; i < matchLen; i++ {
			dst[dp] = dst[dp-offset]
			dp++
		}
	}
	return nil
}

// unshuffle inverse le byte-shuffle de Blosc : les octets sont regroupés par
// position (tous les octets 0 des éléments, puis les octets 1, …).
func unshuffle(src []byte, typesize int) []byte {
	n := len(src)
	if typesize <= 1 || n%typesize != 0 {
		return src
	}
	count := n / typesize
	out := make([]byte, n)
	for i := 0; i < typesize; i++ {
		base := i * count
		for j := 0; j < count; j++ {
			out[j*typesize+i] = src[base+j]
		}
	}
	return out
}
