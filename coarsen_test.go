package xarray

import (
	"reflect"
	"testing"
)

func TestCoarsen1D(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{6}, []float64{1, 2, 3, 4, 5, 6},
		map[string][]float64{"x": {0, 1, 2, 3, 4, 5}}, "v")
	c, err := da.Coarsen("x", 2)
	if err != nil {
		t.Fatalf("Coarsen : %v", err)
	}
	if c.Groups() != 3 {
		t.Errorf("Groups = %d, attendu 3", c.Groups())
	}
	m, _ := c.Mean()
	// blocs {1,2},{3,4},{5,6} -> [1.5 3.5 5.5]
	if !reflect.DeepEqual(m.Data(), []float64{1.5, 3.5, 5.5}) {
		t.Errorf("Coarsen mean = %v, attendu [1.5 3.5 5.5]", m.Data())
	}
	// étiquettes = bornes gauches des blocs
	if crd, _ := m.Coord("x"); !reflect.DeepEqual(crd, []float64{0, 2, 4}) {
		t.Errorf("coord = %v, attendu [0 2 4]", crd)
	}
}

func TestCoarsenTrim(t *testing.T) {
	// Taille 7, facteur 2 -> 3 blocs, dernier élément ignoré.
	da, _ := NewDataArray([]string{"x"}, []int{7}, []float64{1, 2, 3, 4, 5, 6, 7}, nil, "v")
	c, _ := da.Coarsen("x", 2)
	s, _ := c.Sum()
	// {1,2},{3,4},{5,6} -> [3 7 11] (le 7 est ignoré)
	if !reflect.DeepEqual(s.Data(), []float64{3, 7, 11}) {
		t.Errorf("Coarsen trim = %v, attendu [3 7 11]", s.Data())
	}
}

func TestCoarsen2D(t *testing.T) {
	// dims [t, x], coarsen sur t (facteur 2). 4×2 -> 2×2.
	da, _ := NewDataArray([]string{"t", "x"}, []int{4, 2},
		[]float64{1, 2, 3, 4, 5, 6, 7, 8},
		map[string][]float64{"x": {10, 20}}, "v")
	c, _ := da.Coarsen("t", 2)
	s, _ := c.Sum()
	if !reflect.DeepEqual(s.Shape(), []int{2, 2}) {
		t.Errorf("Shape = %v", s.Shape())
	}
	// bloc0 (t=0,1) : [1+3, 2+4]=[4 6] ; bloc1 (t=2,3) : [5+7, 6+8]=[12 14]
	if !reflect.DeepEqual(s.Data(), []float64{4, 6, 12, 14}) {
		t.Errorf("Coarsen 2D = %v", s.Data())
	}
}

func TestCoarsenDataset(t *testing.T) {
	a, _ := NewDataArray([]string{"t"}, []int{4}, []float64{1, 2, 3, 4},
		map[string][]float64{"t": {0, 1, 2, 3}}, "a")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a})
	c, _ := ds.Coarsen("t", 2)
	m, _ := c.Mean()
	av, _ := m.Get("a")
	if !reflect.DeepEqual(av.Data(), []float64{1.5, 3.5}) {
		t.Errorf("Coarsen dataset = %v, attendu [1.5 3.5]", av.Data())
	}
}

func TestCoarsenFacteurTropGrand(t *testing.T) {
	da, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3}, nil, "v")
	if _, err := da.Coarsen("x", 5); err == nil {
		t.Error("erreur attendue : facteur trop grand")
	}
}
