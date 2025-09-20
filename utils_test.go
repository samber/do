package do

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUtilsEmpty(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	value1 := empty[int]()
	is.Empty(value1)

	value2 := empty[*int]()
	is.Nil(value2)
	is.Empty(value2)
}

func TestUtilsDeepEmpty(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	value1 := deepEmpty[int]()
	is.Empty(value1)

	value2 := deepEmpty[*int]()
	is.NotNil(value2)
	is.Empty(*value2)

	value3 := deepEmpty[**int]()
	is.NotNil(value3)
	is.NotNil(*value3)
	is.Empty(**value3)

	value4 := deepEmpty[***int]()
	is.NotNil(value4)
	is.NotNil(*value4)
	is.NotNil(**value4)
	is.Empty(***value4)
}

func TestUtilsMust0(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.Panics(func() {
		must0(fmt.Errorf("error"))
	})
	is.NotPanics(func() {
		must0(nil)
	})
}

func TestUtilsMust1(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.Panics(func() {
		is.Equal(0, must1(42, fmt.Errorf("error")))
	})
	is.NotPanics(func() {
		is.Equal(42, must1(42, nil))
	})
}

func TestUtilsKeys(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	value1 := keys(map[int]string{1: "foo", 2: "bar"})
	sort.IntSlice(value1).Sort()
	is.Equal([]int{1, 2}, value1)

	value2 := keys(map[int]string{})
	is.Empty(value2)
}

func TestUtilsFlatten(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result1 := flatten([][]int{{0, 1}, {2, 3, 4, 5}})

	is.Equal(result1, []int{0, 1, 2, 3, 4, 5})
}

func TestUtilsMap(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result1 := mAp([]int{1, 2, 3, 4}, func(x int, index int) string {
		is.Equal(x-1, index)
		return "Hello"
	})
	result2 := mAp([]int64{1, 2, 3, 4}, func(x int64, index int) string {
		is.Equal(int(x-1), index)
		return strconv.FormatInt(x, 10)
	})

	is.Equal(len(result1), 4)
	is.Equal(len(result2), 4)
	is.Equal(result1, []string{"Hello", "Hello", "Hello", "Hello"})
	is.Equal(result2, []string{"1", "2", "3", "4"})
}

func TestTypesEqual(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.True(typeIsAssignable[any, any]())
	is.True(typeIsAssignable[*any, *any]())
	is.True(typeIsAssignable[int, int]())
	is.True(typeIsAssignable[struct{}, struct{}]())
	is.True(typeIsAssignable[struct{ int }, struct{ int }]())
	is.True(typeIsAssignable[interface{}, any]())
	is.True(typeIsAssignable[interface{ fun() }, interface{ fun() }]())

	is.False(typeIsAssignable[int, any]())
	is.False(typeIsAssignable[*any, any]())
	is.False(typeIsAssignable[int, string]())
	is.False(typeIsAssignable[string, any]())
	is.False(typeIsAssignable[struct{ int }, struct{ any }]())
	is.False(typeIsAssignable[any, interface{ fun() }]())
	is.False(typeIsAssignable[interface{ fun1() }, interface{ fun2() }]())
}

func TestFilter(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	r1 := filter([]int{1, 2, 3, 4}, func(x int, _ int) bool {
		return x%2 == 0
	})

	is.Equal(r1, []int{2, 4})

	r2 := filter([]string{"", "foo", "", "bar", ""}, func(x string, _ int) bool {
		return len(x) > 0
	})

	is.Equal(r2, []string{"foo", "bar"})
}

func TestUtilsOrderedUniq(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result1 := orderedUniq([]string{"a", "b", "c"})
	result2 := orderedUniq([]string{"a", "b", "b", "c"})

	is.Equal(len(result1), 3)
	is.Equal(len(result2), 3)
	is.Equal(result1, []string{"a", "b", "c"})
	is.Equal(result2, []string{"a", "b", "c"})
}

func TestUtilsContains(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.True(contains([]string{"a", "b", "c"}, "a"))
	is.False(contains([]string{"a", "b", "c"}, "z"))
}

func TestUtilsCoalesce(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.Empty(coalesce[int]())
	is.Empty(coalesce(0))
	is.Equal(1, coalesce(0, 1, 2))
	is.Equal(1, coalesce(1, 2, 0))
}

func TestUtilsJobPool(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	p := newJobPool[error](42)
	is.Equal(p.parallelism, uint(42))
}

type test1 struct{}

func (t test1) aMethod() string {
	return "test"
}

type test2 struct{}

func (t test2) aMethod() string {
	return "test"
}

type iTest1 interface {
	aMethod() string
}
type iTest2 interface {
	aMethod() string
}

func TestUtilsGenericCanCastToGeneric(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	is.True(genericCanCastToGeneric[test1, iTest1]())
	is.True(genericCanCastToGeneric[test1, iTest2]())
	is.True(genericCanCastToGeneric[iTest1, iTest2]())
	is.False(genericCanCastToGeneric[test1, test2]())
	is.False(genericCanCastToGeneric[*test1, *test2]())
	is.False(genericCanCastToGeneric[iTest1, test1]())
	is.False(genericCanCastToGeneric[iTest1, *test1]())
	is.False(genericCanCastToGeneric[*lazyTestHeathcheckerOK, iTest1]())
}

func TestUtilsTypeCanCastToGeneric(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// // nil without type
	is.False(typeCanCastToGeneric[iTest1](reflect.TypeOf(nil)))
	is.False(typeCanCastToGeneric[test1](reflect.TypeOf(nil)))
	is.False(typeCanCastToGeneric[iTest1](reflect.TypeOf((iTest1)(nil)))) // no concrete type, only interface

	// nil with type
	is.True(typeCanCastToGeneric[*test1](reflect.TypeOf((*test1)(nil))))
	is.True(typeCanCastToGeneric[iTest1](reflect.TypeOf((*test1)(nil))))

	is.True(typeCanCastToGeneric[*test1](reflect.TypeOf(&test1{})))
	is.True(typeCanCastToGeneric[iTest1](reflect.TypeOf(&test1{})))
	is.True(typeCanCastToGeneric[iTest2](reflect.TypeOf(&test1{})))
	is.True(typeCanCastToGeneric[iTest2](reflect.TypeOf((iTest1)(&test1{}))))
	is.False(typeCanCastToGeneric[test1](reflect.TypeOf((iTest1)(&test1{}))))
	is.False(typeCanCastToGeneric[*test2](reflect.TypeOf(&test1{})))
	is.False(typeCanCastToGeneric[iTest1](reflect.TypeOf(&lazyTestHeathcheckerOK{})))
}
