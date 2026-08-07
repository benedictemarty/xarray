package xarray

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/klauspost/compress/zstd"
)

// zstdDec est un décodeur zstd partagé ; DecodeAll est sûr en usage concurrent.
var zstdDec, _ = zstd.NewReader(nil)

func zstdDecode(src, dst []byte) error {
	res, err := zstdDec.DecodeAll(src, dst[:0])
	if err != nil {
		return err
	}
	if len(res) != len(dst) {
		return fmt.Errorf("zstd: %d octets décodés, %d attendus", len(res), len(dst))
	}
	copy(dst, res)
	return nil
}

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
		cname := m.Cname
		switch cname {
		case "", "lz4", "blosclz", "zstd":
		default:
			return nil, fmt.Errorf("xarray: codec Blosc %q non pris en charge (lz4/blosclz/zstd)", cname)
		}
		return func(src []byte) ([]byte, error) { return bloscDecompress(src, cname) }, nil
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
func bloscDecompress(src []byte, cname string) ([]byte, error) {
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
	// Blosc applique le filtre (shuffle/bitshuffle) PAR BLOC : on décode puis
	// dé-filtre chaque bloc indépendamment avant de l'assembler.
	for b := 0; b < nblocks; b++ {
		off := int(int32(binary.LittleEndian.Uint32(src[16+4*b : 20+4*b])))
		if off < 0 || off > len(src) {
			return nil, fmt.Errorf("xarray: offset de bloc Blosc invalide")
		}
		blkLen := blocksize
		if b == nblocks-1 {
			blkLen = nbytes - b*blocksize
		}
		block := make([]byte, blkLen)
		if err := decodeBloscBlock(src[off:], block, typesize, blocksize, cname); err != nil {
			return nil, err
		}
		switch {
		case bitshuffle:
			ub, err := bitUnshuffle(block, typesize)
			if err != nil {
				return nil, err
			}
			block = ub
		case doShuffle && typesize > 1:
			block = unshuffle(block, typesize)
		}
		copy(out[b*blocksize:], block)
	}
	return out, nil
}

// decodeBloscBlock décode un bloc dans dst (dimensionné à sa taille réelle).
// Blosc découpe un bloc en sous-flux de `neblock = blocksize/typesize` octets
// (règle BLOSC_MIN_BUFFERSIZE=128, basée sur blocksize et non sur la taille du
// bloc), le dernier sous-flux portant le reliquat ; sinon un seul flux. Chaque
// sous-flux est préfixé de sa taille compressée (int32) ; s'il vaut sa taille
// décompressée, il est stocké brut, sinon codé en LZ4.
func decodeBloscBlock(src, dst []byte, typesize, blocksize int, cname string) error {
	blkLen := len(dst)
	const bloscMinBuffer = 128
	// Seuls les codecs LZ4/BLOSCLZ découpent (et seulement les blocs PLEINS en
	// `typesize` sous-flux de blocksize/typesize octets). zstd ne découpe pas ;
	// un bloc partiel (dernier bloc) reste un flux unique. Vérifié sur stores réels.
	splittable := cname == "" || cname == "lz4" || cname == "blosclz"
	neblock := blkLen // un seul flux par défaut
	if splittable && blkLen == blocksize && typesize >= 2 && typesize <= 16 && blocksize/typesize >= bloscMinBuffer {
		neblock = blocksize / typesize
	}
	sp := 0
	for pos := 0; pos < blkLen; pos += neblock {
		outLen := neblock
		if pos+outLen > blkLen {
			outLen = blkLen - pos // sécurité (ne devrait pas arriver)
		}
		if sp+4 > len(src) {
			return fmt.Errorf("xarray: préfixe de sous-flux Blosc tronqué")
		}
		clen := int(int32(binary.LittleEndian.Uint32(src[sp : sp+4])))
		sp += 4
		if clen < 0 || sp+clen > len(src) {
			return fmt.Errorf("xarray: sous-flux Blosc hors bornes (clen=%d)", clen)
		}
		out := dst[pos : pos+outLen]
		var err error
		switch {
		case clen == outLen:
			copy(out, src[sp:sp+clen]) // sous-flux incompressible (brut)
		case cname == "zstd":
			err = zstdDecode(src[sp:sp+clen], out)
		default:
			err = lz4Decompress(src[sp:sp+clen], out)
		}
		if err != nil {
			return fmt.Errorf("xarray: codec %q (sous-flux @%d): %w", cname, pos, err)
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

// bitUnshuffle inverse le bit-shuffle de Blosc : le bloc est vu comme une
// matrice de bits (nelem × typesize*8) transposée ; on la re-transpose. Les bits
// sont numérotés LSB-first à l'intérieur de chaque octet (convention bitshuffle).
func bitUnshuffle(src []byte, typesize int) ([]byte, error) {
	n := len(src)
	if typesize < 1 || n%typesize != 0 {
		return src, nil
	}
	nelem := n / typesize
	// Le transpose de bits « plein » n'est correct que si nelem est un multiple
	// de 8 ; Blosc traite le reliquat (nelem%8) par un chemin distinct que l'on
	// ne décode pas encore. Erreur explicite plutôt que données fausses.
	if nelem%8 != 0 {
		return nil, fmt.Errorf("xarray: Blosc bitshuffle avec nelem=%d non multiple de 8 non pris en charge", nelem)
	}
	elemBits := typesize * 8
	total := n * 8
	out := make([]byte, n)
	getBit := func(buf []byte, pos int) byte { return (buf[pos>>3] >> uint(pos&7)) & 1 }
	for p := 0; p < total; p++ {
		// Disposition shufflée : tous les bits d'indice `bitpos` des nelem
		// éléments, consécutifs, pour bitpos = 0..elemBits-1.
		bitpos := p / nelem
		elem := p % nelem
		if getBit(src, p) != 0 {
			dst := elem*elemBits + bitpos
			out[dst>>3] |= 1 << uint(dst&7)
		}
	}
	return out, nil
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
