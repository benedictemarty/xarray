package xarray

import (
	"reflect"
	"testing"
	"time"
)

func TestDatasetResampleCalendar(t *testing.T) {
	times := []time.Time{
		date(2020, 1, 10), date(2020, 1, 20), date(2020, 2, 5),
	}
	coord := EpochCoord(times)
	temp, _ := NewDataArray([]string{"temps"}, []int{3}, []float64{10, 20, 30},
		map[string][]float64{"temps": coord}, "temperature")
	pluie, _ := NewDataArray([]string{"temps"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"temps": coord}, "pluie")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"temperature": temp, "pluie": pluie})

	g, err := ds.ResampleCalendar("temps", PeriodMonth)
	if err != nil {
		t.Fatalf("ResampleCalendar : %v", err)
	}
	if g.Groups() != 2 {
		t.Errorf("Groups = %d, attendu 2", g.Groups())
	}
	m, _ := g.Mean()
	// janvier temperature (10+20)/2=15, février 30 ; pluie jan (1+2)/2=1.5, fév 3
	tp, _ := m.Get("temperature")
	if !reflect.DeepEqual(tp.Data(), []float64{15, 30}) {
		t.Errorf("temperature mensuelle = %v, attendu [15 30]", tp.Data())
	}
	pl, _ := m.Get("pluie")
	if !reflect.DeepEqual(pl.Data(), []float64{1.5, 3}) {
		t.Errorf("pluie mensuelle = %v, attendu [1.5 3]", pl.Data())
	}
}

func TestDatasetGroupByTimeSeason(t *testing.T) {
	times := []time.Time{
		date(2020, 1, 1), date(2020, 12, 1), date(2020, 7, 1),
	}
	coord := EpochCoord(times)
	a, _ := NewDataArray([]string{"t"}, []int{3}, []float64{2, 4, 100},
		map[string][]float64{"t": coord}, "a")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a})

	g, _ := ds.GroupByTime("t", CompSeason)
	s, _ := g.Sum()
	// DJF : jan+déc = 2+4 = 6 ; JJA : juil = 100
	av, _ := s.Get("a")
	if !reflect.DeepEqual(av.Data(), []float64{6, 100}) {
		t.Errorf("saisonnier = %v, attendu [6 100]", av.Data())
	}
}

func TestDatasetResampleNumerique(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{4}, []float64{10, 20, 30, 40},
		map[string][]float64{"x": {0, 1, 2, 3}}, "a")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a})
	g, _ := ds.Resample("x", 2)
	s, _ := g.Sum()
	av, _ := s.Get("a")
	// bins {0,1}->30, {2,3}->70
	if !reflect.DeepEqual(av.Data(), []float64{30, 70}) {
		t.Errorf("resample numérique = %v, attendu [30 70]", av.Data())
	}
}
