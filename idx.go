package xarray

import "fmt"

// IdxMinAxis renvoie, le long de dim, l'étiquette de coordonnée correspondant au
// minimum (équivalent de `idxmin` de xarray). La dimension doit avoir une
// coordonnée.
func (da *DataArray[T]) IdxMinAxis(dim string) (*DataArray[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: idxmin nécessite une coordonnée pour %q", dim)
	}
	labels := cv.Data()
	return reduceAxisDA[T, T](da, dim, func(s []T) T {
		best := 0
		for i, x := range s[1:] {
			if x < s[best] {
				best = i + 1
			}
		}
		return labels[best]
	})
}

// IdxMaxAxis renvoie, le long de dim, l'étiquette de coordonnée correspondant au
// maximum (équivalent de `idxmax` de xarray).
func (da *DataArray[T]) IdxMaxAxis(dim string) (*DataArray[T], error) {
	cv, ok := da.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: idxmax nécessite une coordonnée pour %q", dim)
	}
	labels := cv.Data()
	return reduceAxisDA[T, T](da, dim, func(s []T) T {
		best := 0
		for i, x := range s[1:] {
			if x > s[best] {
				best = i + 1
			}
		}
		return labels[best]
	})
}
