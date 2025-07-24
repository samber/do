package do

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

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

	value1 := keys[int, string](map[int]string{1: "foo", 2: "bar"})
	sort.IntSlice(value1).Sort()
	is.Equal([]int{1, 2}, value1)

	value2 := keys[int, string](map[int]string{})
	is.Empty(value2)
}

func TestUtilsValues(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	value1 := values[int, string](map[int]string{1: "foo", 2: "bar"})
	sort.StringSlice(value1).Sort()
	is.Equal([]string{"bar", "foo"}, value1)

	value2 := values[int, string](map[int]string{})
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
		is.Equal(int(x-1), index)
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

func TestUtilsMergeMaps(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result1 := mergeMaps(map[string]int{"a": 1, "b": 2, "c": 3}, map[string]int{"c": 4, "d": 5, "e": 6})
	result2 := mergeMaps[string, int]()

	is.Equal(len(result1), 5)
	is.Equal(len(result2), 0)
	is.Equal(result1, map[string]int{"a": 1, "b": 2, "c": 4, "d": 5, "e": 6})
}

func TestUtilsInvertMap(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	result1 := invertMap(map[string]int{"a": 1, "b": 2, "c": 3})
	result2 := invertMap(map[string]int{"a": 1, "b": 1, "c": 3})

	is.Equal(len(result1), 3)
	is.Equal(len(result2), 2)
	is.Equal(result1, map[int]string{1: "a", 2: "b", 3: "c"})
	// is.Equal(result2, map[int]string{1: "b", 3: "c"})
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
	is.Empty(coalesce[int](0))
	is.Equal(1, coalesce[int](0, 1, 2))
	is.Equal(1, coalesce[int](1, 2, 0))
}

func TestUtilsJobPool(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	p := newJobPool[error](42)
	is.Equal(p.parallelism, uint(42))
}
