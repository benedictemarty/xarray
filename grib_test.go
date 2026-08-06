package xarray

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"testing"
)

// buildMinimalGrib2 assemble à la main un message GRIB2 minimal : grille lat/lon
// 2×2, simple packing, valeurs [10 20 30 40] (bitsPerValue=8, R=0, E=0, D=0),
// La1=45°, Lo1=0°, Di=1°, Dj=2°, mode de balayage 0 (j du nord vers le sud).
func buildMinimalGrib2() []byte {
	be := binary.BigEndian

	sec1 := make([]byte, 21)
	be.PutUint32(sec1[0:4], 21)
	sec1[4] = 1

	sec3 := make([]byte, 72)
	be.PutUint32(sec3[0:4], 72)
	sec3[4] = 3
	be.PutUint32(sec3[6:10], 4) // nombre de points
	be.PutUint16(sec3[12:14], 0)
	sec3[14] = 6                          // forme de la Terre
	be.PutUint32(sec3[30:34], 2)          // Ni
	be.PutUint32(sec3[34:38], 2)          // Nj
	be.PutUint32(sec3[46:50], 45_000_000) // La1 = 45°
	be.PutUint32(sec3[50:54], 0)          // Lo1 = 0°
	be.PutUint32(sec3[54:58], 44_000_000) // La2 (indicatif)
	be.PutUint32(sec3[58:62], 1_000_000)  // Lo2
	be.PutUint32(sec3[63:67], 1_000_000)  // Di = 1°
	be.PutUint32(sec3[67:71], 2_000_000)  // Dj = 2°
	sec3[71] = 0                          // scanning mode : j négatif

	sec5 := make([]byte, 21)
	be.PutUint32(sec5[0:4], 21)
	sec5[4] = 5
	be.PutUint32(sec5[5:9], 4) // nombre de valeurs
	be.PutUint16(sec5[9:11], 0)
	be.PutUint32(sec5[11:15], math.Float32bits(0)) // R = 0
	be.PutUint16(sec5[15:17], 0)                   // E = 0
	be.PutUint16(sec5[17:19], 0)                   // D = 0
	sec5[19] = 8                                   // bitsPerValue

	sec6 := make([]byte, 6)
	be.PutUint32(sec6[0:4], 6)
	sec6[4] = 6
	sec6[5] = 255 // pas de bitmap

	data := []byte{10, 20, 30, 40} // 4 valeurs 8 bits
	sec7 := make([]byte, 5+len(data))
	be.PutUint32(sec7[0:4], uint32(5+len(data)))
	sec7[4] = 7
	copy(sec7[5:], data)

	sec8 := []byte{'7', '7', '7', '7'}

	body := bytes.Join([][]byte{sec1, sec3, sec5, sec6, sec7, sec8}, nil)
	total := 16 + len(body)

	sec0 := make([]byte, 16)
	copy(sec0[0:4], "GRIB")
	sec0[6] = 0 // discipline
	sec0[7] = 2 // édition
	be.PutUint64(sec0[8:16], uint64(total))

	return append(sec0, body...)
}

func TestGribSimplePacking(t *testing.T) {
	msgs, err := ReadGrib(bytes.NewReader(buildMinimalGrib2()))
	if err != nil {
		t.Fatalf("ReadGrib : %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("%d message(s), attendu 1", len(msgs))
	}
	m := msgs[0]
	if m.Ni != 2 || m.Nj != 2 {
		t.Errorf("Ni/Nj = %d/%d", m.Ni, m.Nj)
	}
	if !reflect.DeepEqual(m.Values, []float64{10, 20, 30, 40}) {
		t.Errorf("valeurs = %v, attendu [10 20 30 40]", m.Values)
	}

	da, err := m.ToDataArray("champ")
	if err != nil {
		t.Fatalf("ToDataArray : %v", err)
	}
	if !reflect.DeepEqual(da.Shape(), []int{2, 2}) {
		t.Errorf("Shape = %v", da.Shape())
	}
	// latitude : 45, 43 (j vers le sud, Dj=2) ; longitude : 0, 1 (Di=1)
	if lat, _ := da.Coord("latitude"); !reflect.DeepEqual(lat, []float64{45, 43}) {
		t.Errorf("latitude = %v, attendu [45 43]", lat)
	}
	if lon, _ := da.Coord("longitude"); !reflect.DeepEqual(lon, []float64{0, 1}) {
		t.Errorf("longitude = %v, attendu [0 1]", lon)
	}
}

func TestGribEdition1Refusee(t *testing.T) {
	msg := buildMinimalGrib2()
	msg[7] = 1 // édition 1
	if _, err := ReadGrib(bytes.NewReader(msg)); err == nil {
		t.Error("erreur attendue : édition 1 non gérée")
	}
}

func TestGribSignMag16(t *testing.T) {
	if signMag16(0x8007) != -7 {
		t.Errorf("signMag16(0x8007) = %d, attendu -7", signMag16(0x8007))
	}
	if signMag16(0x0005) != 5 {
		t.Errorf("signMag16(0x0005) = %d, attendu 5", signMag16(0x0005))
	}
}
