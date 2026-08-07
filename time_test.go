package xarray

import (
	"reflect"
	"testing"
	"time"
)

func TestEpochAllerRetour(t *testing.T) {
	orig := time.Date(2026, 8, 7, 12, 30, 0, 0, time.UTC)
	sec := EpochSeconds(orig)
	got := TimeFromEpoch(sec)
	if !got.Equal(orig) {
		t.Errorf("aller-retour epoch : %v != %v", got, orig)
	}
}

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func TestResampleCalendarMonth(t *testing.T) {
	times := []time.Time{
		date(2020, 1, 15),
		date(2020, 2, 10),
		date(2020, 2, 20),
		date(2021, 1, 5),
	}
	da, _ := NewDataArray([]string{"temps"}, []int{4}, []float64{1, 2, 3, 4},
		map[string][]float64{"temps": EpochCoord(times)}, "v")

	r, err := ResampleCalendar(da, "temps", PeriodMonth)
	if err != nil {
		t.Fatalf("ResampleCalendar : %v", err)
	}
	if r.Groups() != 3 {
		t.Errorf("Groups = %d, attendu 3 (2020-01, 2020-02, 2021-01)", r.Groups())
	}
	s, _ := r.Sum()
	// 2020-01 : {1} ; 2020-02 : {2+3=5} ; 2021-01 : {4}
	if !reflect.DeepEqual(s.Data(), []float64{1, 5, 4}) {
		t.Errorf("Sum mensuelle = %v, attendu [1 5 4]", s.Data())
	}
	// Les étiquettes sont les débuts de mois (epoch).
	c, _ := s.Coord("temps")
	attendu := []float64{
		EpochSeconds(date(2020, 1, 1)),
		EpochSeconds(date(2020, 2, 1)),
		EpochSeconds(date(2021, 1, 1)),
	}
	if !reflect.DeepEqual(c, attendu) {
		t.Errorf("coord = %v, attendu %v", c, attendu)
	}
}

func TestResampleCalendarYear(t *testing.T) {
	times := []time.Time{
		date(2020, 1, 15), date(2020, 6, 10), date(2020, 12, 20), date(2021, 3, 5),
	}
	da, _ := NewDataArray([]string{"temps"}, []int{4}, []float64{10, 20, 30, 40},
		map[string][]float64{"temps": EpochCoord(times)}, "v")
	r, _ := ResampleCalendar(da, "temps", PeriodYear)
	m, _ := r.Mean()
	// 2020 : (10+20+30)/3 = 20 ; 2021 : 40
	if !reflect.DeepEqual(m.Data(), []float64{20, 40}) {
		t.Errorf("Mean annuelle = %v, attendu [20 40]", m.Data())
	}
}

func TestResampleCalendarSansCoord(t *testing.T) {
	da, _ := NewDataArray([]string{"temps"}, []int{2}, []float64{1, 2}, nil, "v")
	if _, err := ResampleCalendar(da, "temps", PeriodMonth); err == nil {
		t.Error("erreur attendue : pas de coordonnée")
	}
}
