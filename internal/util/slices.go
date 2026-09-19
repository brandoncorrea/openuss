package util

func Map[T, R any](coll []T, f func(T) R) []R {
	out := make([]R, len(coll))
	for i, item := range coll {
		out[i] = f(item)
	}
	return out
}

func Remove[T any](coll []T, pred func(T) bool) []T {
	out := []T{}
	for _, item := range coll {
		if !pred(item) {
			out = append(out, item)
		}
	}
	return out
}
