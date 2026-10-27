package bslices

// All reports whether pred holds for every element of xs.
// It returns true for an empty slice.
func All[T any](xs []T, pred func(T) bool) bool {
	for _, x := range xs {
		if !pred(x) {
			return false
		}
	}
	return true
}

// Map returns a new slice with f applied to every element of xs.
func Map[T, U any](xs []T, f func(T) U) []U {
	result := make([]U, len(xs))
	for i, x := range xs {
		result[i] = f(x)
	}
	return result
}

// MapWithError returns a new slice with f applied to every element of xs.
// It stops at the first error and returns it with a nil slice.
func MapWithError[T, U any](xs []T, f func(T) (U, error)) ([]U, error) {
	result := make([]U, len(xs))
	for i, x := range xs {
		y, err := f(x)
		if err != nil {
			return nil, err
		}
		result[i] = y
	}
	return result, nil
}
