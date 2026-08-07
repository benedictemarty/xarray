package xarray

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Lecture d'un sous-ensemble du format GRIB2 (OMM) : messages en grille
// régulière lat/lon (« regular_ll », template 3.0) avec **simple packing**
// (template 5.0), sans bitmap.
//
// Non géré (documenté) : GRIB édition 1, autres grilles (gaussienne, Lambert…),
// et surtout le **complex/second-order packing** (`grid_second_order`), qui
// domine les fichiers opérationnels. Pour ces cas, la bibliothèque de référence
// est ecCodes (ECMWF), via cgo/cfgrib. Voir docs/GRIB.md.

// GribMessage est un message GRIB2 décodé : les valeurs sur une grille lat/lon.
type GribMessage struct {
	Ni, Nj         int
	La1, Lo1       float64 // premier point (degrés)
	Di, Dj         float64 // incréments (degrés)
	JScansPositive bool    // false = du nord vers le sud (cas courant)
	Values         []float64
}

// ToDataArray convertit le message en DataArray[float64] de dimensions
// (latitude, longitude), avec les coordonnées correspondantes.
func (m *GribMessage) ToDataArray(name string) (*DataArray[float64], error) {
	lat := make([]float64, m.Nj)
	for j := 0; j < m.Nj; j++ {
		if m.JScansPositive {
			lat[j] = m.La1 + float64(j)*m.Dj
		} else {
			lat[j] = m.La1 - float64(j)*m.Dj
		}
	}
	lon := make([]float64, m.Ni)
	for i := 0; i < m.Ni; i++ {
		lon[i] = m.Lo1 + float64(i)*m.Di
	}
	return NewDataArray(
		[]string{"latitude", "longitude"}, []int{m.Nj, m.Ni}, m.Values,
		map[string][]float64{"latitude": lat, "longitude": lon}, name,
	)
}

// ReadGrib lit tous les messages GRIB2 (grille lat/lon, simple packing) d'un flux.
func ReadGrib(r io.Reader) ([]*GribMessage, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var msgs []*GribMessage
	pos := 0
	for pos+16 <= len(raw) {
		// Recherche du marqueur "GRIB".
		if !(raw[pos] == 'G' && raw[pos+1] == 'R' && raw[pos+2] == 'I' && raw[pos+3] == 'B') {
			pos++
			continue
		}
		edition := raw[pos+7]
		if edition != 2 {
			return nil, fmt.Errorf("xarray: seul GRIB édition 2 est pris en charge (édition %d)", edition)
		}
		total := int(binary.BigEndian.Uint64(raw[pos+8 : pos+16]))
		if pos+total > len(raw) {
			return nil, fmt.Errorf("xarray: message GRIB tronqué")
		}
		msg, err := decodeGrib2Message(raw[pos : pos+total])
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
		pos += total
	}
	if len(msgs) == 0 {
		return nil, fmt.Errorf("xarray: aucun message GRIB2 trouvé")
	}
	return msgs, nil
}

func decodeGrib2Message(msg []byte) (*GribMessage, error) {
	out := &GribMessage{}
	var (
		refValue     float32
		binaryScale  int
		decimalScale int
		bitsPerValue int
		drt          int
		sec5         []byte
		dataBytes    []byte
		haveGrid     bool
		haveDRS      bool
		haveData     bool
	)

	p := 16 // après la section 0
	for p+4 <= len(msg) {
		// Fin de message : "7777".
		if msg[p] == '7' && msg[p+1] == '7' && msg[p+2] == '7' && msg[p+3] == '7' {
			break
		}
		secLen := int(binary.BigEndian.Uint32(msg[p : p+4]))
		if secLen < 5 || p+secLen > len(msg) {
			return nil, fmt.Errorf("xarray: section GRIB invalide (longueur %d)", secLen)
		}
		secNum := msg[p+4]
		sec := msg[p : p+secLen]

		switch secNum {
		case 3: // Grid Definition Section
			tmpl := binary.BigEndian.Uint16(sec[12:14])
			if tmpl != 0 {
				return nil, fmt.Errorf("xarray: grille GRIB template %d non gérée (seul regular_ll=0)", tmpl)
			}
			out.Ni = int(binary.BigEndian.Uint32(sec[30:34]))
			out.Nj = int(binary.BigEndian.Uint32(sec[34:38]))
			out.La1 = float64(int32(binary.BigEndian.Uint32(sec[46:50]))) * 1e-6
			out.Lo1 = float64(int32(binary.BigEndian.Uint32(sec[50:54]))) * 1e-6
			// Template 3.0 : octets 64-67 = Di (incrément i), 68-71 = Dj.
			out.Di = float64(int32(binary.BigEndian.Uint32(sec[63:67]))) * 1e-6
			out.Dj = float64(int32(binary.BigEndian.Uint32(sec[67:71]))) * 1e-6
			scanning := sec[71]
			out.JScansPositive = scanning&0x40 != 0
			haveGrid = true

		case 5: // Data Representation Section
			drt = int(binary.BigEndian.Uint16(sec[9:11]))
			if drt != 0 && drt != 2 && drt != 3 {
				return nil, fmt.Errorf("xarray: packing GRIB template %d non géré (0=simple, 2/3=complex ; les templates locaux comme 50002 requièrent ecCodes)", drt)
			}
			sec5 = append([]byte(nil), sec...)
			refValue = math.Float32frombits(binary.BigEndian.Uint32(sec[11:15]))
			binaryScale = signMag16(binary.BigEndian.Uint16(sec[15:17]))
			decimalScale = signMag16(binary.BigEndian.Uint16(sec[17:19]))
			bitsPerValue = int(sec[19])
			haveDRS = true

		case 6: // Bit-Map Section
			if sec[5] != 255 {
				return nil, fmt.Errorf("xarray: bitmap GRIB présent (indicateur %d) non géré", sec[5])
			}

		case 7: // Data Section
			dataBytes = sec[5:]
			haveData = true
		}
		p += secLen
	}

	if !haveGrid || !haveDRS || !haveData {
		return nil, fmt.Errorf("xarray: sections GRIB manquantes (grid=%v drs=%v data=%v)", haveGrid, haveDRS, haveData)
	}

	n := out.Ni * out.Nj
	pow2 := math.Pow(2, float64(binaryScale))
	pow10 := math.Pow(10, float64(decimalScale))

	var scaled []int64
	if drt == 0 {
		scaled = make([]int64, n)
		br := &bitReader{data: dataBytes}
		for i := 0; i < n; i++ {
			x, err := br.read(bitsPerValue)
			if err != nil {
				return nil, err
			}
			scaled[i] = int64(x)
		}
	} else {
		var err error
		scaled, err = unpackComplex(sec5, dataBytes, n, bitsPerValue, drt)
		if err != nil {
			return nil, err
		}
	}

	values := make([]float64, n)
	for i := 0; i < n; i++ {
		// Y = (R + X·2^E) / 10^D.
		values[i] = (float64(refValue) + float64(scaled[i])*pow2) / pow10
	}
	out.Values = values
	return out, nil
}

// unpackComplex décode la section 7 en complex packing (templates WMO 5.2 et
// 5.3). Algorithme conforme à g2clib (comunpack) : alignement à l'octet après
// chaque bloc (références, largeurs, longueurs de groupe).
func unpackComplex(sec5, data []byte, n, bitsPerValue, drt int) ([]int64, error) {
	if sec5[22] != 0 {
		return nil, fmt.Errorf("xarray: gestion des valeurs manquantes GRIB non gérée (%d)", sec5[22])
	}
	ng := int(binary.BigEndian.Uint32(sec5[31:35]))
	refGroupWidths := int(sec5[35])
	nbitsGroupWidths := int(sec5[36])
	refGroupLengths := int(binary.BigEndian.Uint32(sec5[37:41]))
	lengthIncrement := int(sec5[41])
	trueLengthLastGroup := int(binary.BigEndian.Uint32(sec5[42:46]))
	nbitsGroupLengths := int(sec5[46])

	order, nds := 0, 0
	if drt == 3 {
		order = int(sec5[47])
		nds = int(sec5[48])
	}

	br := &bitReader{data: data}

	// Différenciation spatiale : valeurs initiales + minimum global (signé).
	initials := make([]int64, order)
	var omin int64
	if order > 0 {
		nbitsd := nds * 8
		for g := 0; g < order; g++ {
			v, err := br.read(nbitsd)
			if err != nil {
				return nil, err
			}
			initials[g] = int64(v)
		}
		sign, err := br.read(1)
		if err != nil {
			return nil, err
		}
		mag, err := br.read(nbitsd - 1)
		if err != nil {
			return nil, err
		}
		omin = int64(mag)
		if sign == 1 {
			omin = -omin
		}
	}

	// Références de groupe (bitsPerValue bits) + alignement octet.
	refs := make([]int64, ng)
	if bitsPerValue != 0 {
		for k := 0; k < ng; k++ {
			v, err := br.read(bitsPerValue)
			if err != nil {
				return nil, err
			}
			refs[k] = int64(v)
		}
		br.align()
	}

	// Largeurs de groupe (+ référence) + alignement octet.
	widths := make([]int, ng)
	if nbitsGroupWidths != 0 {
		for k := 0; k < ng; k++ {
			v, err := br.read(nbitsGroupWidths)
			if err != nil {
				return nil, err
			}
			widths[k] = int(v)
		}
		br.align()
	}
	for k := 0; k < ng; k++ {
		widths[k] += refGroupWidths
	}

	// Longueurs de groupe (échelle + référence, dernier = trueLength) + alignement.
	lengths := make([]int, ng)
	if nbitsGroupLengths != 0 {
		for k := 0; k < ng; k++ {
			v, err := br.read(nbitsGroupLengths)
			if err != nil {
				return nil, err
			}
			lengths[k] = int(v)
		}
		br.align()
	}
	total := 0
	for k := 0; k < ng; k++ {
		lengths[k] = lengths[k]*lengthIncrement + refGroupLengths
	}
	lengths[ng-1] = trueLengthLastGroup
	for k := 0; k < ng; k++ {
		total += lengths[k]
	}
	if total != n {
		return nil, fmt.Errorf("xarray: somme des longueurs de groupe (%d) != nombre de points (%d)", total, n)
	}

	// Valeurs : pour chaque groupe, length valeurs de width bits + réf du groupe.
	vals := make([]int64, n)
	idx := 0
	for k := 0; k < ng; k++ {
		w := widths[k]
		if w == 0 {
			for j := 0; j < lengths[k]; j++ {
				vals[idx] = refs[k]
				idx++
			}
			continue
		}
		for j := 0; j < lengths[k]; j++ {
			v, err := br.read(w)
			if err != nil {
				return nil, err
			}
			vals[idx] = refs[k] + int64(v)
			idx++
		}
	}

	// Réversion de la différenciation spatiale.
	switch order {
	case 0:
	case 1:
		vals[0] = initials[0]
		for i := 1; i < n; i++ {
			vals[i] += omin
			vals[i] += vals[i-1]
		}
	case 2:
		vals[0] = initials[0]
		vals[1] = initials[1]
		for i := 2; i < n; i++ {
			vals[i] += omin
			vals[i] += 2*vals[i-1] - vals[i-2]
		}
	default:
		return nil, fmt.Errorf("xarray: ordre de différenciation spatiale %d non géré", order)
	}
	return vals, nil
}

// signMag16 décode un entier 16 bits en représentation signe-magnitude (GRIB).
func signMag16(u uint16) int {
	if u&0x8000 != 0 {
		return -int(u & 0x7fff)
	}
	return int(u)
}

// bitReader lit des entiers de largeur arbitraire dans un flux de bits big-endian.
type bitReader struct {
	data   []byte
	bitpos int
}

// align avance jusqu'à la prochaine frontière d'octet.
func (b *bitReader) align() {
	if b.bitpos%8 != 0 {
		b.bitpos += 8 - (b.bitpos % 8)
	}
}

func (b *bitReader) read(nbits int) (uint64, error) {
	if nbits == 0 {
		return 0, nil
	}
	var v uint64
	for i := 0; i < nbits; i++ {
		byteIdx := b.bitpos >> 3
		if byteIdx >= len(b.data) {
			return 0, fmt.Errorf("xarray: flux de données GRIB trop court")
		}
		bit := (b.data[byteIdx] >> (7 - (b.bitpos & 7))) & 1
		v = (v << 1) | uint64(bit)
		b.bitpos++
	}
	return v, nil
}
