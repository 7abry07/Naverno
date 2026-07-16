package util

func Remove[t any](slice []t, elem t, compare func(e1, e2 t) bool) []t {
	result := []t{}
	for _, e := range slice {
		if !compare(e, elem) {
			result = append(result, e)
		}
	}
	return result
}
