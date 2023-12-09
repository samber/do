package di

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultInjector(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}

	DefaultInjector = New()

	Provide(nil, func(i *Injector) (*test, error) {
		return &test{foobar: "42"}, nil
	})

	is.Len(DefaultInjector.services, 1)

	service, err := Invoke[*test](nil)

	is.Equal(test{foobar: "42"}, *service)
	is.Nil(err)
}

func TestProvide(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	i := New()

	Provide(i, func(i *Injector) (*test, error) {
		return &test{}, nil
	})

	Provide(i, func(i *Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	})

	is.Panics(func() {
		// try to erase previous instance
		Provide(i, func(i *Injector) (test, error) {
			return test{}, fmt.Errorf("error")
		})
	})

	is.Len(i.services, 2)

	s1, ok1 := i.services["*di.test"]
	is.True(ok1)
	if ok1 {
		s, ok := s1.(Service[*test])
		is.True(ok)
		if ok {
			is.Equal("*di.test", s.getName())
		}
	}

	s2, ok2 := i.services["di.test"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(Service[test])
		is.True(ok)
		if ok {
			is.Equal("di.test", s.getName())
		}
	}

	_, ok3 := i.services["*di.plop"]
	is.False(ok3)
}

func TestProvideValue(t *testing.T) {
	is := assert.New(t)

	i := New()

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	ProvideNamedValue(i, "foobar", 42)
	ProvideNamedValue(i, "hello", _test)

	is.Len(i.services, 2)

	s1, ok1 := i.services["foobar"]
	is.True(ok1)
	if ok1 {
		s, ok := s1.(Service[int])
		is.True(ok)
		if ok {
			is.Equal("foobar", s.getName())
			instance, err := s.getInstance(i)
			is.EqualValues(42, instance)
			is.Nil(err)
		}
	}

	s2, ok2 := i.services["hello"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(Service[test])
		is.True(ok)
		if ok {
			is.Equal("hello", s.getName())
			instance, err := s.getInstance(i)
			is.EqualValues(_test, instance)
			is.Nil(err)
		}
	}
}

func TestInvoke(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}

	i := New()

	Provide(i, func(i *Injector) (test, error) {
		return test{foobar: "foobar"}, nil
	})

	is.Len(i.services, 1)

	s0a, ok0a := i.services["di.test"]
	is.True(ok0a)

	s0b, ok0b := s0a.(*ServiceLazy[test])
	is.True(ok0b)
	is.False(s0b.built)

	s1, err1 := Invoke[test](i)
	is.Nil(err1)
	if err1 == nil {
		is.Equal("foobar", s1.foobar)
	}

	is.True(s0b.built)

	_, err2 := Invoke[*test](i)
	is.NotNil(err2)
	is.Errorf(err2, "do: service not found")
}

func TestInvokeNamed(t *testing.T) {
	is := assert.New(t)

	i := New()

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	ProvideNamedValue(i, "foobar", 42)
	ProvideNamedValue(i, "hello", _test)

	is.Len(i.services, 2)

	service0, err0 := InvokeNamed[string](i, "plop")
	is.NotNil(err0)
	is.Empty(service0)

	instance1, err1 := InvokeNamed[test](i, "hello")
	is.Nil(err1)
	is.EqualValues(_test, instance1)
	is.EqualValues("foobar", instance1.foobar)

	instance2, err2 := InvokeNamed[int](i, "foobar")
	is.Nil(err2)
	is.EqualValues(42, instance2)
}

func TestMustInvoke(t *testing.T) {
	is := assert.New(t)

	i := New()

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	Provide(i, func(i *Injector) (test, error) {
		return _test, nil
	})

	is.Len(i.services, 1)

	is.Panics(func() {
		_ = MustInvoke[string](i)
	})

	is.NotPanics(func() {
		instance1 := MustInvoke[test](i)
		is.EqualValues(_test, instance1)
	})
}

func TestMustInvokeNamed(t *testing.T) {
	is := assert.New(t)

	i := New()

	ProvideNamedValue(i, "foobar", 42)

	is.Len(i.services, 1)

	is.Panics(func() {
		_ = MustInvokeNamed[string](i, "hello")
	})

	is.Panics(func() {
		_ = MustInvokeNamed[string](i, "foobar")
	})

	is.NotPanics(func() {
		instance1 := MustInvokeNamed[int](i, "foobar")
		is.EqualValues(42, instance1)
	})
}

func TestShutdown(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}

	i := New()

	Provide(i, func(i *Injector) (test, error) {
		return test{foobar: "foobar"}, nil
	})

	instance, err := Invoke[test](i)
	is.Equal(test{foobar: "foobar"}, instance)
	is.Nil(err)

	err = Shutdown[test](i)
	is.Nil(err)

	instance, err = Invoke[test](i)
	is.Empty(instance)
	is.NotNil(err)

	err = Shutdown[test](i)
	is.NotNil(err)
}

func TestMustShutdown(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}

	i := New()

	Provide(i, func(i *Injector) (test, error) {
		return test{foobar: "foobar"}, nil
	})

	instance, err := Invoke[test](i)
	is.Equal(test{foobar: "foobar"}, instance)
	is.Nil(err)

	is.NotPanics(func() {
		MustShutdown[test](i)
	})

	instance, err = Invoke[test](i)
	is.Empty(instance)
	is.NotNil(err)

	is.Panics(func() {
		MustShutdown[test](i)
	})
}

func TestShutdownNamed(t *testing.T) {
	is := assert.New(t)

	i := New()

	ProvideNamedValue(i, "foobar", 42)

	instance, err := InvokeNamed[int](i, "foobar")
	is.Equal(42, instance)
	is.Nil(err)

	err = ShutdownNamed(i, "foobar")
	is.Nil(err)

	instance, err = InvokeNamed[int](i, "foobar")
	is.Empty(instance)
	is.NotNil(err)

	err = ShutdownNamed(i, "foobar")
	is.NotNil(err)
}

func TestMustShutdownNamed(t *testing.T) {
	is := assert.New(t)

	i := New()

	ProvideNamedValue(i, "foobar", 42)

	instance, err := InvokeNamed[int](i, "foobar")
	is.Equal(42, instance)
	is.Nil(err)

	is.NotPanics(func() {
		MustShutdownNamed(i, "foobar")
	})

	instance, err = InvokeNamed[int](i, "foobar")
	is.Empty(instance)
	is.NotNil(err)

	is.Panics(func() {
		MustShutdownNamed(i, "foobar")
	})
}

func TestDoubleInjection(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	i := New()

	is.NotPanics(func() {
		Provide(i, func(i *Injector) (*test, error) {
			return &test{}, nil
		})
	})

	is.PanicsWithError("DI: service `*di.test` has already been declared", func() {
		Provide(i, func(i *Injector) (*test, error) {
			return &test{}, nil
		})
	})

	is.PanicsWithError("DI: service `*di.test` has already been declared", func() {
		ProvideValue(i, &test{})
	})

	is.PanicsWithError("DI: service `*di.test` has already been declared", func() {
		ProvideNamed(i, "*di.test", func(i *Injector) (*test, error) {
			return &test{}, nil
		})
	})

	is.PanicsWithError("DI: service `*di.test` has already been declared", func() {
		ProvideNamedValue(i, "*di.test", &test{})
	})
}

func TestOverride(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar int
	}

	i := New()

	is.NotPanics(func() {
		Provide(i, func(i *Injector) (*test, error) {
			return &test{42}, nil
		})
		is.Equal(42, MustInvoke[*test](i).foobar)

		Override(i, func(i *Injector) (*test, error) {
			return &test{1}, nil
		})
		is.Equal(1, MustInvoke[*test](i).foobar)

		OverrideNamed(i, "*di.test", func(i *Injector) (*test, error) {
			return &test{2}, nil
		})
		is.Equal(2, MustInvoke[*test](i).foobar)

		OverrideValue(i, &test{3})
		is.Equal(3, MustInvoke[*test](i).foobar)

		OverrideNamedValue(i, "*di.test", &test{4})
		is.Equal(4, MustInvoke[*test](i).foobar)
	})
}
