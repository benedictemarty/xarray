package xarray

// Number est la contrainte des types numériques supportés par les tableaux.
// Elle couvre les entiers signés/non signés et les flottants.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}
