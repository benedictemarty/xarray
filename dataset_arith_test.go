package xarray

import (
	"reflect"
	"testing"
)

func TestDatasetArith(t *testing.T) {
	a1, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 2, 3},
		map[string][]float64{"x": {0, 1, 2}}, "a")
	b1, _ := NewDataArray([]string{"x"}, []int{3}, []float64{10, 20, 30},
		map[string][]float64{"x": {0, 1, 2}}, "b")
	ds1, _ := NewDataset(map[string]*DataArray[float64]{"a": a1, "b": b1})

	a2, _ := NewDataArray([]string{"x"}, []int{3}, []float64{1, 1, 1},
		map[string][]float64{"x": {0, 1, 2}}, "a")
	b2, _ := NewDataArray([]string{"x"}, []int{3}, []float64{2, 2, 2},
		map[string][]float64{"x": {0, 1, 2}}, "b")
	ds2, _ := NewDataset(map[string]*DataArray[float64]{"a": a2, "b": b2})

	sum, err := ds1.Add(ds2)
	if err != nil {
		t.Fatalf("Add : %v", err)
	}
	sa, _ := sum.Get("a")
	sb, _ := sum.Get("b")
	if !reflect.DeepEqual(sa.Data(), []float64{2, 3, 4}) {
		t.Errorf("a+a = %v", sa.Data())
	}
	if !reflect.DeepEqual(sb.Data(), []float64{12, 22, 32}) {
		t.Errorf("b+b = %v", sb.Data())
	}

	// Multiplication.
	prod, _ := ds1.Mul(ds2)
	pa, _ := prod.Get("a")
	if !reflect.DeepEqual(pa.Data(), []float64{1, 2, 3}) {
		t.Errorf("a*1 = %v", pa.Data())
	}
}

func TestDatasetAddScalar(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2}, nil, "a")
	b, _ := NewDataArray([]string{"x"}, []int{2}, []float64{3, 4}, nil, "b")
	ds, _ := NewDataset(map[string]*DataArray[float64]{"a": a, "b": b})
	r, _ := ds.AddScalar(100)
	ra, _ := r.Get("a")
	rb, _ := r.Get("b")
	if !reflect.DeepEqual(ra.Data(), []float64{101, 102}) || !reflect.DeepEqual(rb.Data(), []float64{103, 104}) {
		t.Errorf("AddScalar = %v / %v", ra.Data(), rb.Data())
	}
}

func TestDatasetArithVarManquante(t *testing.T) {
	a, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2}, nil, "a")
	ds1, _ := NewDataset(map[string]*DataArray[float64]{"a": a})
	c, _ := NewDataArray([]string{"x"}, []int{2}, []float64{1, 2}, nil, "c")
	ds2, _ := NewDataset(map[string]*DataArray[float64]{"c": c})
	if _, err := ds1.Add(ds2); err == nil {
		t.Error("erreur attendue : variable a absente de ds2")
	}
}
