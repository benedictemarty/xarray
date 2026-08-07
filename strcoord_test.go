package xarray

import (
	"reflect"
	"testing"
)

func TestStrCoordSelStr(t *testing.T) {
	// Données numériques, coordonnée textuelle « station ».
	da, _ := NewDataArray([]string{"station", "heure"}, []int{3, 2},
		[]float64{1, 2, 3, 4, 5, 6}, nil, "temp")
	da, err := da.WithStrCoord("station", []string{"Paris", "Lyon", "Nice"})
	if err != nil {
		t.Fatalf("WithStrCoord : %v", err)
	}
	if c, _ := da.StrCoord("station"); !reflect.DeepEqual(c, []string{"Paris", "Lyon", "Nice"}) {
		t.Errorf("StrCoord = %v", c)
	}

	// Sélection par étiquette textuelle.
	lyon, err := da.SelStr("station", "Lyon")
	if err != nil {
		t.Fatalf("SelStr : %v", err)
	}
	if !reflect.DeepEqual(lyon.Data(), []float64{3, 4}) {
		t.Errorf("SelStr(Lyon) = %v, attendu [3 4]", lyon.Data())
	}
	if !reflect.DeepEqual(lyon.Dims(), []string{"heure"}) {
		t.Errorf("Dims = %v", lyon.Dims())
	}

	// Étiquette absente.
	if _, err := da.SelStr("station", "Berlin"); err == nil {
		t.Error("erreur attendue : étiquette absente")
	}
}

func TestStrCoordPreserveeParIsel(t *testing.T) {
	da, _ := NewDataArray([]string{"station", "heure"}, []int{2, 3},
		[]float64{1, 2, 3, 4, 5, 6}, nil, "v")
	da, _ = da.WithStrCoord("station", []string{"A", "B"})

	// Isel sur heure : la coordonnée textuelle station doit subsister.
	sub, _ := da.Isel("heure", 0)
	if c, _ := sub.StrCoord("station"); !reflect.DeepEqual(c, []string{"A", "B"}) {
		t.Errorf("station après Isel = %v, attendu [A B]", c)
	}
}

func TestSelStrMany(t *testing.T) {
	da, _ := NewDataArray([]string{"station"}, []int{3}, []float64{10, 20, 30}, nil, "v")
	da, _ = da.WithStrCoord("station", []string{"Paris", "Lyon", "Nice"})
	r, err := da.SelStrMany("station", []string{"Nice", "Paris"})
	if err != nil {
		t.Fatalf("SelStrMany : %v", err)
	}
	if !reflect.DeepEqual(r.Data(), []float64{30, 10}) {
		t.Errorf("SelStrMany = %v, attendu [30 10]", r.Data())
	}
	// La coordonnée textuelle suit.
	if c, _ := r.StrCoord("station"); !reflect.DeepEqual(c, []string{"Nice", "Paris"}) {
		t.Errorf("StrCoord = %v", c)
	}
}

func TestWithStrCoordLongueurInvalide(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3}, nil, "v")
	if _, err := da.WithStrCoord("x", []string{"a", "b"}); err == nil {
		t.Error("erreur attendue : longueur incompatible")
	}
}
