package ndarray

import (
	"reflect"
	"testing"
)

func TestInto(t *testing.T) {
	a, _ := New([]int{2, 2}, []float64{1, 2, 3, 4})
	b, _ := New([]int{2, 2}, []float64{10, 20, 30, 40})
	dst := Zeros(2, 2)

	if err := AddInto(dst, a, b); err != nil {
		t.Fatalf("AddInto : %v", err)
	}
	if !reflect.DeepEqual(dst.Data(), []float64{11, 22, 33, 44}) {
		t.Errorf("AddInto = %v", dst.Data())
	}
	if err := MulInto(dst, a, b); err != nil {
		t.Fatalf("MulInto : %v", err)
	}
	if !reflect.DeepEqual(dst.Data(), []float64{10, 40, 90, 160}) {
		t.Errorf("MulInto = %v", dst.Data())
	}

	// a et b ne doivent pas avoir été modifiés.
	if !reflect.DeepEqual(a.Data(), []float64{1, 2, 3, 4}) {
		t.Errorf("a modifié : %v", a.Data())
	}
}

func TestIntoFormeInvalide(t *testing.T) {
	a, _ := New([]int{2, 2}, []float64{1, 2, 3, 4})
	b, _ := New([]int{2, 2}, []float64{1, 2, 3, 4})
	dst := Zeros(3, 3)
	if err := AddInto(dst, a, b); err == nil {
		t.Error("erreur attendue : forme de destination incompatible")
	}
}

func TestAddInPlace(t *testing.T) {
	a, _ := New([]int{3}, []float64{1, 2, 3})
	b, _ := New([]int{3}, []float64{10, 10, 10})
	if err := a.AddInPlace(b); err != nil {
		t.Fatalf("AddInPlace : %v", err)
	}
	if !reflect.DeepEqual(a.Data(), []float64{11, 12, 13}) {
		t.Errorf("AddInPlace = %v", a.Data())
	}
}

func TestReutilisationDestination(t *testing.T) {
	// Accumulation de plusieurs tableaux dans un seul buffer, sans allocation.
	acc := Zeros(4)
	for k := 0; k < 5; k++ {
		v, _ := New([]int{4}, []float64{1, 1, 1, 1})
		if err := acc.AddInPlace(v); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(acc.Data(), []float64{5, 5, 5, 5}) {
		t.Errorf("accumulation = %v", acc.Data())
	}
}
