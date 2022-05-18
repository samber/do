package do

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceLazyName(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	provider1 := func(i *Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i *Injector) (test, error) {
		return _test, nil
	}

	service1 := newServiceLazy("foobar", provider1)
	is.Equal("foobar", service1.getName())

	service2 := newServiceLazy("foobar", provider2)
	is.Equal("foobar", service2.getName())
}

func TestServiceLazyInstance(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	provider1 := func(i *Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i *Injector) (test, error) {
		return _test, nil
	}
	provider3 := func(i *Injector) (int, error) {
		panic("error")
	}
	provider4 := func(i *Injector) (int, error) {
		panic(fmt.Errorf("error"))
	}
	provider5 := func(i *Injector) (int, error) {
		return 42, fmt.Errorf("error")
	}

	i := New()

	service1 := newServiceLazy("foobar", provider1)
	instance1, err1 := service1.getInstance(i)
	is.Nil(err1)
	is.Equal(42, instance1)

	service2 := newServiceLazy("hello", provider2)
	instance2, err2 := service2.getInstance(i)
	is.Nil(err2)
	is.Equal(_test, instance2)

	is.Panics(func() {
		service3 := newServiceLazy("baz", provider3)
		_, _ = service3.getInstance(i)
	})

	is.NotPanics(func() {
		service4 := newServiceLazy("plop", provider4)
		instance4, err4 := service4.getInstance(i)
		is.NotNil(err4)
		is.Empty(instance4)
		expected := fmt.Errorf("error")
		is.Equal(expected, err4)
	})

	is.NotPanics(func() {
		service5 := newServiceLazy("plop", provider5)
		instance5, err5 := service5.getInstance(i)
		is.NotNil(err5)
		is.Empty(instance5)
		expected := fmt.Errorf("error")
		is.Equal(expected, err5)
	})
}
