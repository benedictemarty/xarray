package ml

import (
	"math"
	"testing"

	nd "github.com/bmarty/xarray/ndarray"
)

func TestStandardize(t *testing.T) {
	// Colonne 0 : [1 3] -> moyenne 2, écart-type 1 -> [-1 1]
	// Colonne 1 : [10 30] -> moyenne 20, écart-type 10 -> [-1 1]
	x, _ := nd.New([]int{2, 2}, []float64{1, 10, 3, 30})
	out, mean, std, err := Standardize(x)
	if err != nil {
		t.Fatalf("Standardize : %v", err)
	}
	if mean.Data()[0] != 2 || mean.Data()[1] != 20 {
		t.Errorf("moyennes = %v", mean.Data())
	}
	if std.Data()[0] != 1 || std.Data()[1] != 10 {
		t.Errorf("écarts-types = %v", std.Data())
	}
	got := out.Data()
	attendu := []float64{-1, -1, 1, 1}
	for i := range attendu {
		if math.Abs(got[i]-attendu[i]) > 1e-9 {
			t.Errorf("standardisé = %v, attendu %v", got, attendu)
			break
		}
	}
}

func TestLinearRegressionApprendUneDroite(t *testing.T) {
	// Données synthétiques : y = 2*x1 + 3*x2 + 1 (bruit nul).
	m := 200
	xd := make([]float64, m*2)
	yd := make([]float64, m)
	for i := 0; i < m; i++ {
		x1 := float64(i%10) - 5
		x2 := float64((i*7)%13) - 6
		xd[i*2] = x1
		xd[i*2+1] = x2
		yd[i] = 2*x1 + 3*x2 + 1
	}
	X, _ := nd.New([]int{m, 2}, xd)
	y, _ := nd.New([]int{m}, yd)

	model := NewLinearRegression(0.01, 5000)
	if err := model.Fit(X, y); err != nil {
		t.Fatalf("Fit : %v", err)
	}

	// Les poids doivent converger vers [2, 3] et le biais vers 1.
	w := model.W.Data()
	if math.Abs(w[0]-2) > 0.05 || math.Abs(w[1]-3) > 0.05 {
		t.Errorf("poids = %v, attendu ~[2 3]", w)
	}
	if math.Abs(model.B-1) > 0.05 {
		t.Errorf("biais = %v, attendu ~1", model.B)
	}

	// L'erreur quadratique moyenne doit être quasi nulle.
	pred, _ := model.Predict(X)
	mse, _ := MSE(pred, y)
	if mse > 1e-2 {
		t.Errorf("MSE = %v, attendu ~0", mse)
	}
}

func TestLinearRegressionNonEntraine(t *testing.T) {
	model := NewLinearRegression(0.01, 100)
	X, _ := nd.New([]int{2, 2}, []float64{1, 2, 3, 4})
	if _, err := model.Predict(X); err == nil {
		t.Error("erreur attendue : modèle non entraîné")
	}
}
