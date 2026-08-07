//go:build !eccodes

// Ce fichier fournit un remplacement pur-Go quand le backend ecCodes n'est pas
// compilé (absence du tag `eccodes`). Il permet à `go build ./...` de réussir
// sans la dépendance C, tout en indiquant clairement comment activer le backend.
package eccodesgrib

import (
	"fmt"

	"github.com/benedictemarty/xarray"
)

// ReadFile renvoie une erreur explicite tant que le backend ecCodes n'est pas
// activé. Recompilez avec `-tags eccodes` (et libeccodes disponible).
func ReadFile(path string) ([]*xarray.DataArray[float64], error) {
	return nil, fmt.Errorf("eccodesgrib: backend non compilé ; recompilez avec -tags eccodes (nécessite libeccodes)")
}
