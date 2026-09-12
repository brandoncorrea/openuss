package util

func Map[T, R any](coll []T, f func(T) R) []R {
	out := make([]R, len(coll))
	for i, item := range coll {
		out[i] = f(item)
	}
	return out
}
