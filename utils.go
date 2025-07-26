package do

func empty[T any]() (t T) {
	return
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func keys[K comparable, V any](in map[K]V) []K {
	out := make([]K, 0, len(in))

	for k := range in {
		out = append(out, k)
	}

	return out
}

func mAp[T any, R any](in []T, f func(T) R) []R {
	out := make([]R, len(in))

	for i, item := range in {
		out[i] = f(item)
	}

	return out
}

func invertMap[K comparable, V comparable](in map[K]V) map[V]K {
	out := make(map[V]K, len(in))

	for k, v := range in {
		out[v] = k
	}

	return out
}
