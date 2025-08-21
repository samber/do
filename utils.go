package do

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"reflect"
	"sync"
)

//
// This file could be replaced with a dependency on a library like samber/lo, but I wanted to keep the dependencies to a minimum.
//

func empty[T any]() (t T) {
	return
}

func deepEmpty[T any]() T {
	var o T
	t := reflect.TypeOf(o)

	v := deepEmptyMakeValue(t) // reflect.Value with the desired shape
	return v.Interface().(T)
}

func deepEmptyMakeValue(t reflect.Type) reflect.Value {
	// Base case: not a pointer -> just zero of this type.
	if t.Kind() != reflect.Ptr {
		return reflect.Zero(t)
	}

	// Recursive case: pointer -> allocate pointer, set its Elem to the
	// recursively-constructed zero value of the element type.
	elem := deepEmptyMakeValue(t.Elem())
	p := reflect.New(t.Elem())
	p.Elem().Set(elem)
	return p
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

func flatten[T any](collection [][]T) []T {
	totalLen := 0
	for i := range collection {
		totalLen += len(collection[i])
	}

	result := make([]T, 0, totalLen)
	for i := range collection {
		result = append(result, collection[i]...)
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

func typeIsAssignable[T, AssignTo any]() bool {
	_, ok := any((*T)(nil)).(*AssignTo)
	return ok
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

func coalesce[T comparable](v ...T) (result T) {
	for _, e := range v {
		if e != result {
			result = e
			break
		}
	}

	return
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

func newJobPool[R any](parallelism uint) *jobPool[R] {
	return &jobPool[R]{
		parallelism: parallelism,
		jobs:        make(chan func(), 1000), // 🤮 @TODO: change that

		startOnce: sync.Once{},
		stopOnce:  sync.Once{},
	}
}

type jobPool[R any] struct {
	parallelism uint
	jobs        chan func()

	startOnce sync.Once
	stopOnce  sync.Once
}

func (p *jobPool[R]) rpc(f func() R) <-chan R {
	c := make(chan R, 1) // a single message will be sent before closing

	p.jobs <- func() {
		defer close(c)
		c <- f()
	}

	return c
}

func (p *jobPool[R]) start() {
	p.startOnce.Do(func() {
		//nolint:gosec
		for i := 0; i < int(p.parallelism); i++ {
			go func() {
				for job := range p.jobs {
					job()
				}
			}()
		}
	})
}

func (p *jobPool[R]) stop() {
	p.stopOnce.Do(func() {
		close(p.jobs)
	})
}

func raceWithTimeout(ctx context.Context, fn func(context.Context) error) error {
	_, ok := ctx.Deadline()
	if !ok {
		return fn(ctx)
	}

	err := make(chan error, 1)
	go func() {
		err <- fn(ctx)
	}()

	select {
	case e := <-err:
		return e
	case <-ctx.Done():
		return fmt.Errorf("%w: %w", ErrHealthCheckTimeout, ctx.Err())
	}
}

// Previously, we used to perform check like this:
// _, ok := any(empty[Initial]()).(Alias)
// But it was not working when Initial was an interface, so we now
// use reflection to check if Initial implements Alias.
func genericCanCastToGeneric[From any, To any]() bool {
	var from From
	anyFrom := any(from)

	if anyFrom == nil { // check for interface by reflect
		typeFrom := reflect.TypeOf(&from).Elem()
		typeTo := reflect.TypeOf((*To)(nil)).Elem()

		toInterface := typeTo.Kind() == reflect.Interface

		return toInterface && typeFrom.Implements(typeTo)
	}

	_, ok := anyFrom.(To)
	return ok
}

func typeCanCastToGeneric[To any](fromType reflect.Type) bool {
	toType := reflect.TypeOf((*To)(nil)).Elem()
	return typeCanCastToType(fromType, toType)
}

func typeCanCastToType(fromType reflect.Type, toType reflect.Type) bool {
	if fromType == nil {
		return false
	}

	if fromType == toType {
		return true
	}

	// Check assignable
	if fromType.AssignableTo(toType) {
		return true
	}

	return false
}
