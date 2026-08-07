package xarray

import "fmt"

// Rééchantillonnage et regroupement temporel au niveau Dataset. Les bins sont
// calculés depuis la coordonnée partagée de dim ; l'agrégation (Sum/Mean/Min/Max)
// est propagée via DatasetGroupBy (les variables sans la dimension sont
// conservées/converties).

func (ds *Dataset[T]) resampleCoord(dim string) ([]T, error) {
	cv, ok := ds.coords[dim]
	if !ok {
		return nil, fmt.Errorf("xarray: aucune coordonnée %q dans le dataset", dim)
	}
	return cv.data, nil
}

// Resample regroupe par intervalles réguliers (largeur freq) d'une coordonnée
// numérique. Renvoie un DatasetGroupBy (agrégations Sum/Mean/Min/Max).
func (ds *Dataset[T]) Resample(dim string, freq T) (*DatasetGroupBy[T], error) {
	data, err := ds.resampleCoord(dim)
	if err != nil {
		return nil, err
	}
	labels, groups := binGroups(data, freq)
	return &DatasetGroupBy[T]{ds: ds, dim: dim, labels: labels, groups: groups}, nil
}

// ResampleCalendar regroupe par période calendaire (coordonnée en secondes epoch).
func (ds *Dataset[T]) ResampleCalendar(dim string, p Period) (*DatasetGroupBy[T], error) {
	data, err := ds.resampleCoord(dim)
	if err != nil {
		return nil, err
	}
	labels, groups := calendarGroups(data, p)
	return &DatasetGroupBy[T]{ds: ds, dim: dim, labels: labels, groups: groups}, nil
}

// GroupByTime regroupe par composante temporelle (mois, saison, …) — climatologie.
func (ds *Dataset[T]) GroupByTime(dim string, c TimeComponent) (*DatasetGroupBy[T], error) {
	data, err := ds.resampleCoord(dim)
	if err != nil {
		return nil, err
	}
	labels, groups := componentGroups(data, c)
	return &DatasetGroupBy[T]{ds: ds, dim: dim, labels: labels, groups: groups}, nil
}
