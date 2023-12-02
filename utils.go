package do

import (
	"crypto/rand"
	"fmt"
	"io"
)

//
// This file could be replaced with a dependency on a library like samber/lo, but I wanted to keep the dependencies to a minimum.
//

func empty[T any]() (t T) {
	return
}

func must0(err error) {
	if err != nil {
		panic(err)
	}
}

func must1[A any](a A, err error) A {
	if err != nil {
		panic(err)
	}

	return a
}

func keys[K comparable, V any](in map[K]V) []K {
	result := make([]K, 0, len(in))

	for k := range in {
		result = append(result, k)
	}

	return result
}

func mAp[T any, R any](collection []T, iteratee func(T, int) R) []R {
	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item, i)
	}

	return result
}

func filter[V any](collection []V, predicate func(item V, index int) bool) []V {
	result := make([]V, 0, len(collection))

	for i, item := range collection {
		if predicate(item, i) {
			result = append(result, item)
		}
	}

	return result
}

func invertMap[K comparable, V comparable](in map[K]V) map[V]K {
	out := map[V]K{}

	for k, v := range in {
		out[v] = k
	}

	return out
}

func reverseSlice[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func orderedUniq[V comparable](in []V) []V {
	out := []V{}
	present := map[V]struct{}{}

	for _, v := range in {
		if _, ok := present[v]; !ok {
			out = append(out, v)
			present[v] = struct{}{}
		}
	}

	return out
}

func contains[T comparable](list []T, elem T) bool {
	for _, v := range list {
		if v == elem {
			return true
		}
	}

	return false
}

// https://gist.github.com/rkravchik/d9733e1d2d626188eb91df751471d739
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
