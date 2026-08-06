package ml

import (
	"fmt"

	nd "github.com/bmarty/xarray/ndarray"
)

// LinearRegression est une régression linéaire y = X·w + b entraînée par
// descente de gradient (erreur quadratique moyenne).
type LinearRegression struct {
	W      *nd.NDArray // poids (n)
	B      float64     // biais
	LR     float64     // taux d'apprentissage
	Epochs int         // nombre d'itérations
}

// NewLinearRegression crée un modèle avec des hyperparamètres par défaut
// raisonnables (à ajuster selon les données ; standardiser aide la convergence).
func NewLinearRegression(lr float64, epochs int) *LinearRegression {
	return &LinearRegression{LR: lr, Epochs: epochs}
}

// Predict renvoie X·w + b pour X (m×n).
func (m *LinearRegression) Predict(x *nd.NDArray) (*nd.NDArray, error) {
	if m.W == nil {
		return nil, fmt.Errorf("ml: modèle non entraîné")
	}
	xw, err := nd.MatVec(x, m.W)
	if err != nil {
		return nil, err
	}
	return xw.AddScalar(m.B), nil
}

// Fit entraîne le modèle sur X (m×n) et y (m) par descente de gradient.
func (m *LinearRegression) Fit(x, y *nd.NDArray) error {
	xs := x.Shape()
	if len(xs) != 2 {
		return fmt.Errorf("ml: X doit être 2D (%dD)", len(xs))
	}
	if y.Ndim() != 1 || y.Shape()[0] != xs[0] {
		return fmt.Errorf("ml: y doit être 1D de longueur %d", xs[0])
	}
	nSamples, nFeat := xs[0], xs[1]
	m.W = nd.Zeros(nFeat)
	m.B = 0
	xt, err := x.T() // (n×m), réutilisée à chaque époque
	if err != nil {
		return err
	}
	invM := 1.0 / float64(nSamples)

	for e := 0; e < m.Epochs; e++ {
		pred, err := m.Predict(x) // (m)
		if err != nil {
			return err
		}
		errVec, err := pred.Sub(y) // (m) : prédiction - cible
		if err != nil {
			return err
		}
		// Gradient des poids : (1/m) Xᵀ·err
		gw, err := nd.MatVec(xt, errVec) // (n)
		if err != nil {
			return err
		}
		gw = gw.MulScalar(m.LR * invM)
		newW, err := m.W.Sub(gw)
		if err != nil {
			return err
		}
		m.W = newW
		// Gradient du biais : (1/m) Σ err = moyenne(err)
		m.B -= m.LR * errVec.Mean()
	}
	return nil
}

// MSE renvoie l'erreur quadratique moyenne entre deux vecteurs.
func MSE(pred, y *nd.NDArray) (float64, error) {
	d, err := pred.Sub(y)
	if err != nil {
		return 0, err
	}
	sq, err := d.Mul(d)
	if err != nil {
		return 0, err
	}
	return sq.Mean(), nil
}
