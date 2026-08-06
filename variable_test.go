package xarray

import (
	"reflect"
	"testing"
)

func TestNewVariable(t *testing.T) {
	v, err := NewVariable([]string{"x", "y"}, []int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if v.Ndim() != 2 {
		t.Errorf("Ndim = %d, attendu 2", v.Ndim())
	}
	if v.Size() != 6 {
		t.Errorf("Size = %d, attendu 6", v.Size())
	}
	if got := v.Shape(); !reflect.DeepEqual(got, []int{2, 3}) {
		t.Errorf("Shape = %v, attendu [2 3]", got)
	}
}

func TestNewVariableErreurs(t *testing.T) {
	cases := []struct {
		nom   string
		dims  []string
		shape []int
		data  []float64
	}{
		{"dims != shape", []string{"x"}, []int{2, 3}, []float64{1, 2}},
		{"data incohérent", []string{"x"}, []int{3}, []float64{1, 2}},
		{"dimension vide", []string{""}, []int{1}, []float64{1}},
		{"dimension dupliquée", []string{"x", "x"}, []int{1, 1}, []float64{1}},
		{"taille négative", []string{"x"}, []int{-1}, nil},
	}
	for _, c := range cases {
		t.Run(c.nom, func(t *testing.T) {
			if _, err := NewVariable(c.dims, c.shape, c.data); err == nil {
				t.Errorf("erreur attendue pour le cas %q", c.nom)
			}
		})
	}
}

func TestVariableAt(t *testing.T) {
	v, _ := NewVariable([]string{"x", "y"}, []int{2, 3}, []float64{1, 2, 3, 4, 5, 6})
	got, err := v.At(1, 2)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if got != 6 {
		t.Errorf("At(1,2) = %v, attendu 6", got)
	}
	if _, err := v.At(5, 0); err == nil {
		t.Error("erreur attendue pour indice hors bornes")
	}
}

func TestVariableIsel(t *testing.T) {
	v, _ := NewVariable([]string{"x", "y"}, []int{2, 3}, []float64{1, 2, 3, 4, 5, 6})

	// Sélection de la ligne x=1 -> [4 5 6]
	sub, err := v.Isel("x", 1)
	if err != nil {
		t.Fatalf("erreur inattendue : %v", err)
	}
	if !reflect.DeepEqual(sub.Dims(), []string{"y"}) {
		t.Errorf("dims = %v, attendu [y]", sub.Dims())
	}
	if !reflect.DeepEqual(sub.Data(), []float64{4, 5, 6}) {
		t.Errorf("data = %v, attendu [4 5 6]", sub.Data())
	}

	// Sélection de la colonne y=0 -> [1 4]
	col, _ := v.Isel("y", 0)
	if !reflect.DeepEqual(col.Data(), []float64{1, 4}) {
		t.Errorf("data = %v, attendu [1 4]", col.Data())
	}

	if _, err := v.Isel("z", 0); err == nil {
		t.Error("erreur attendue pour dimension inconnue")
	}
}

func TestVariableAttrs(t *testing.T) {
	v, _ := NewVariable([]string{"x"}, []int{1}, []float64{1})
	v.SetAttr("unité", "mètres")
	if v.Attrs()["unité"] != "mètres" {
		t.Errorf("attribut non conservé")
	}
}
