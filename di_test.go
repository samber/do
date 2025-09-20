package do

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNameOf(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	is.Equal("int", NameOf[int]())
	is.Equal("github.com/samber/do/v2.eagerTest", NameOf[eagerTest]())
	is.Equal("*github.com/samber/do/v2.eagerTest", NameOf[*eagerTest]())
	is.Equal("*map[int]bool", NameOf[*map[int]bool]())
	is.Equal("github.com/samber/do/v2.serviceWrapper[int]", NameOf[serviceWrapper[int]]())
}

func TestProvide(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}

	i := New()

	Provide(i, func(i Injector) (*test, error) {
		return &test{}, nil
	})

	Provide(i, func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	})

	is.Panics(func() {
		// try to erase previous instance
		Provide(i, func(i Injector) (test, error) {
			return test{}, fmt.Errorf("error")
		})
	})

	is.Len(i.self.services, 2)

	s1, ok1 := i.self.services["*github.com/samber/do/v2.test"]
	is.True(ok1)
	if ok1 {
		s, ok := s1.(serviceWrapper[*test])
		is.True(ok)
		if ok {
			is.Equal("*github.com/samber/do/v2.test", s.getName())
		}
	}

	s2, ok2 := i.self.services["github.com/samber/do/v2.test"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(serviceWrapper[test])
		is.True(ok)
		if ok {
			is.Equal("github.com/samber/do/v2.test", s.getName())
		}
	}

	_, ok3 := i.self.services["github.com/samber/do/v2.*plop"]
	is.False(ok3)

	// Test that all services share the same references when invoked multiple times
	instance1, err1 := Invoke[*test](i)
	is.NoError(err1)
	is.NotNil(instance1)

	instance2, err2 := Invoke[*test](i)
	is.NoError(err2)
	is.NotNil(instance2)

	// Lazy services should return the same instance (singleton behavior)
	is.Same(instance1, instance2, "Lazy services should return the same instance")

	// Test that error services are handled correctly
	instance3, err3 := Invoke[test](i)
	is.Error(err3)
	is.Empty(instance3)
	is.EqualError(err3, "error")
}

func TestProvideNamed(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}

	i := New()

	ProvideNamed(i, "*foobar", func(i Injector) (*test, error) {
		return &test{}, nil
	})

	ProvideNamed(i, "foobar", func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	})

	is.Panics(func() {
		// try to erase previous instance
		ProvideNamed(i, "foobar", func(i Injector) (test, error) {
			return test{}, fmt.Errorf("error")
		})
	})

	is.Len(i.self.services, 2)

	s1, ok1 := i.self.services["*foobar"]
	is.True(ok1)
	if ok1 {
		s, ok := s1.(serviceWrapper[*test])
		is.True(ok)
		if ok {
			is.Equal("*foobar", s.getName())
		}
	}

	s2, ok2 := i.self.services["foobar"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(serviceWrapper[test])
		is.True(ok)
		if ok {
			is.Equal("foobar", s.getName())
		}
	}

	_, ok3 := i.self.services["*do.plop"]
	is.False(ok3)

	// Test that all services share the same references when invoked multiple times
	instance1, err1 := InvokeNamed[*test](i, "*foobar")
	is.NoError(err1)
	is.NotNil(instance1)

	instance2, err2 := InvokeNamed[*test](i, "*foobar")
	is.NoError(err2)
	is.NotNil(instance2)

	// Lazy services should return the same instance (singleton behavior)
	is.Same(instance1, instance2, "Named lazy services should return the same instance")

	// Test that error services are handled correctly
	instance3, err3 := InvokeNamed[test](i, "foobar")
	is.Error(err3)
	is.Empty(instance3)
	is.EqualError(err3, "error")
}

func TestProvideValue(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	ProvideValue(i, 42)
	ProvideValue(i, _test)

	is.Len(i.self.services, 2)

	s1, ok1 := i.self.services["int"]
	is.True(ok1)
	if ok1 {
		s, ok := s1.(serviceWrapper[int])
		is.True(ok)
		if ok {
			is.Equal("int", s.getName())
			instance, err := s.getInstance(i)
			is.Equal(42, instance)
			is.NoError(err)
		}
	}

	s2, ok2 := i.self.services["github.com/samber/do/v2.test"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(serviceWrapper[test])
		is.True(ok)
		if ok {
			is.Equal("github.com/samber/do/v2.test", s.getName())
			instance, err := s.getInstance(i)
			is.Equal(_test, instance)
			is.NoError(err)
		}
	}

	// Test that all services share the same references when invoked multiple times
	instance1, err1 := Invoke[int](i)
	is.NoError(err1)
	is.Equal(42, instance1)

	instance2, err2 := Invoke[int](i)
	is.NoError(err2)
	is.Equal(42, instance2)

	// Eager value services should return the same value (not necessarily same reference for primitives)
	is.Equal(instance1, instance2, "Value services should return the same value")

	// Test struct values
	structInstance1, err3 := Invoke[test](i)
	is.NoError(err3)
	is.Equal(_test, structInstance1)

	structInstance2, err4 := Invoke[test](i)
	is.NoError(err4)
	is.Equal(_test, structInstance2)

	// Value services should return the same value
	is.Equal(structInstance1, structInstance2, "Struct value services should return the same value")
}

func TestProvideNamedValue(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	ProvideNamedValue(i, "foobar", 42)
	ProvideNamedValue(i, "hello", _test)

	is.Len(i.self.services, 2)

	s1, ok1 := i.self.services["foobar"]
	is.True(ok1)
	if ok1 {
		s, ok := s1.(serviceWrapper[int])
		is.True(ok)
		if ok {
			is.Equal("foobar", s.getName())
			instance, err := s.getInstance(i)
			is.Equal(42, instance)
			is.NoError(err)
		}
	}

	s2, ok2 := i.self.services["hello"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(serviceWrapper[test])
		is.True(ok)
		if ok {
			is.Equal("hello", s.getName())
			instance, err := s.getInstance(i)
			is.Equal(_test, instance)
			is.NoError(err)
		}
	}

	// Test that all services share the same references when invoked multiple times
	instance1, err1 := InvokeNamed[int](i, "foobar")
	is.NoError(err1)
	is.Equal(42, instance1)

	instance2, err2 := InvokeNamed[int](i, "foobar")
	is.NoError(err2)
	is.Equal(42, instance2)

	// Named value services should return the same value
	is.Equal(instance1, instance2, "Named value services should return the same value")

	// Test struct values
	structInstance1, err3 := InvokeNamed[test](i, "hello")
	is.NoError(err3)
	is.Equal(_test, structInstance1)

	structInstance2, err4 := InvokeNamed[test](i, "hello")
	is.NoError(err4)
	is.Equal(_test, structInstance2)

	// Named value services should return the same value
	is.Equal(structInstance1, structInstance2, "Named struct value services should return the same value")
}

func TestProvideTransient(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		ID int
	}

	i := New()

	ProvideTransient(i, func(i Injector) (*test, error) {
		return &test{ID: 1}, nil
	})

	ProvideTransient(i, func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	})

	is.Panics(func() {
		// try to erase previous instance
		Provide(i, func(i Injector) (test, error) {
			return test{}, fmt.Errorf("error")
		})
	})

	is.Len(i.self.services, 2)

	s1, ok1 := i.self.services["*github.com/samber/do/v2.test"]
	is.True(ok1)
	if ok1 {
		s, ok := s1.(serviceWrapper[*test])
		is.True(ok)
		if ok {
			is.Equal("*github.com/samber/do/v2.test", s.getName())
		}
	}

	s2, ok2 := i.self.services["github.com/samber/do/v2.test"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(serviceWrapper[test])
		is.True(ok)
		if ok {
			is.Equal("github.com/samber/do/v2.test", s.getName())
		}
	}

	_, ok3 := i.self.services["github.com/samber/do/v2.*plop"]
	is.False(ok3)

	// Test that transient services create new instances each time
	instance1, err1 := Invoke[*test](i)
	is.NoError(err1)
	is.NotNil(instance1)

	instance2, err2 := Invoke[*test](i)
	is.NoError(err2)
	is.NotNil(instance2)

	// Transient services should return different instances (not the same reference)
	is.NotSame(instance1, instance2, "Transient services should return different instances")

	// Test that error services are handled correctly
	instance3, err3 := Invoke[test](i)
	is.Error(err3)
	is.Empty(instance3)
	is.EqualError(err3, "error")
}

func TestProvideNamedTransient(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		ID int
	}

	i := New()

	ProvideNamedTransient(i, "*foobar", func(i Injector) (*test, error) {
		return &test{ID: 1}, nil
	})

	ProvideNamedTransient(i, "foobar", func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	})

	is.Panics(func() {
		// try to erase previous instance
		ProvideNamedTransient(i, "foobar", func(i Injector) (test, error) {
			return test{}, fmt.Errorf("error")
		})
	})

	is.Len(i.self.services, 2)

	s1, ok1 := i.self.services["*foobar"]
	is.True(ok1)
	if ok1 {
		s, ok := s1.(serviceWrapper[*test])
		is.True(ok)
		if ok {
			is.Equal("*foobar", s.getName())
		}
	}

	s2, ok2 := i.self.services["foobar"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(serviceWrapper[test])
		is.True(ok)
		if ok {
			is.Equal("foobar", s.getName())
		}
	}

	_, ok3 := i.self.services["*do.plop"]
	is.False(ok3)

	// Test that named transient services create new instances each time
	instance1, err1 := InvokeNamed[*test](i, "*foobar")
	is.NoError(err1)
	is.NotNil(instance1)

	instance2, err2 := InvokeNamed[*test](i, "*foobar")
	is.NoError(err2)
	is.NotNil(instance2)

	// Named transient services should return different instances (not the same reference)
	is.NotSame(instance1, instance2, "Named transient services should return different instances")

	// Test that error services are handled correctly
	instance3, err3 := InvokeNamed[test](i, "foobar")
	is.Error(err3)
	is.Empty(instance3)
	is.EqualError(err3, "error")
}

func TestProvide_race(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	injector := New()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		Provide(injector, func(i Injector) (int, error) {
			return 42, nil
		})
		wg.Done()
	}()

	go func() {
		Provide(injector, func(i Injector) (*lazyTest, error) {
			return &lazyTest{}, nil
		})
		wg.Done()
	}()

	wg.Wait()
}

func TestOverride(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		foobar int
	}

	i := New()

	is.NotPanics(func() {
		Provide(i, func(i Injector) (*test, error) {
			return &test{42}, nil
		})
		is.Equal(42, MustInvoke[*test](i).foobar)

		Override(i, func(i Injector) (*test, error) {
			return &test{1}, nil
		})
		is.Equal(1, MustInvoke[*test](i).foobar)

		OverrideNamed(i, "*github.com/samber/do/v2.test", func(i Injector) (*test, error) {
			return &test{2}, nil
		})
		is.Equal(2, MustInvoke[*test](i).foobar)

		OverrideValue(i, &test{3})
		is.Equal(3, MustInvoke[*test](i).foobar)

		OverrideNamedValue(i, "*github.com/samber/do/v2.test", &test{4})
		is.Equal(4, MustInvoke[*test](i).foobar)
	})
}

func TestOverrideNamed(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		foobar int
	}

	i := New()

	Provide(i, func(i Injector) (*test, error) {
		return &test{42}, nil
	})
	is.Equal(42, MustInvoke[*test](i).foobar)

	OverrideNamed(i, "*github.com/samber/do/v2.test", func(i Injector) (*test, error) {
		return &test{2}, nil
	})
	is.Equal(2, MustInvoke[*test](i).foobar)
}

func TestOverrideValue(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		foobar int
	}

	i := New()

	Provide(i, func(i Injector) (*test, error) {
		return &test{42}, nil
	})
	is.Equal(42, MustInvoke[*test](i).foobar)

	OverrideNamed(i, "*github.com/samber/do/v2.test", func(i Injector) (*test, error) {
		return &test{2}, nil
	})
	is.Equal(2, MustInvoke[*test](i).foobar)
}

func TestOverrideNamedValue(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		foobar int
	}

	i := New()

	Provide(i, func(i Injector) (*test, error) {
		return &test{42}, nil
	})
	is.Equal(42, MustInvoke[*test](i).foobar)

	OverrideNamedValue(i, "*github.com/samber/do/v2.test", &test{4})
	is.Equal(4, MustInvoke[*test](i).foobar)
}

func TestOverrideTransient(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		foobar int
	}

	i := New()

	// Provide initial transient service
	ProvideTransient(i, func(i Injector) (*test, error) {
		return &test{42}, nil
	})

	// Test initial service
	instance1, err1 := Invoke[*test](i)
	is.NoError(err1)
	is.Equal(42, instance1.foobar)

	instance2, err2 := Invoke[*test](i)
	is.NoError(err2)
	is.Equal(42, instance2.foobar)

	// Transient services should return different instances
	is.NotSame(instance1, instance2, "Transient services should return different instances")

	// Override with new transient provider
	OverrideTransient(i, func(i Injector) (*test, error) {
		return &test{100}, nil
	})

	// Test overridden service
	instance3, err3 := Invoke[*test](i)
	is.NoError(err3)
	is.Equal(100, instance3.foobar)

	instance4, err4 := Invoke[*test](i)
	is.NoError(err4)
	is.Equal(100, instance4.foobar)

	// Overridden transient services should still return different instances
	is.NotSame(instance3, instance4, "Overridden transient services should return different instances")

	// Test OverrideNamedTransient
	OverrideNamedTransient(i, NameOf[*test](), func(i Injector) (*test, error) {
		return &test{200}, nil
	})

	// Test final overridden service
	instance5, err5 := Invoke[*test](i)
	is.NoError(err5)
	is.Equal(200, instance5.foobar)
}

func TestOverrideNamedTransient(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		foobar int
	}

	i := New()

	// Provide initial named transient service
	ProvideNamedTransient(i, "test-service", func(i Injector) (*test, error) {
		return &test{42}, nil
	})

	// Test initial service
	instance1, err1 := InvokeNamed[*test](i, "test-service")
	is.NoError(err1)
	is.Equal(42, instance1.foobar)

	instance2, err2 := InvokeNamed[*test](i, "test-service")
	is.NoError(err2)
	is.Equal(42, instance2.foobar)

	// Named transient services should return different instances
	is.NotSame(instance1, instance2, "Named transient services should return different instances")

	// Override with new named transient provider
	OverrideNamedTransient(i, "test-service", func(i Injector) (*test, error) {
		return &test{100}, nil
	})

	// Test overridden service
	instance3, err3 := InvokeNamed[*test](i, "test-service")
	is.NoError(err3)
	is.Equal(100, instance3.foobar)

	instance4, err4 := InvokeNamed[*test](i, "test-service")
	is.NoError(err4)
	is.Equal(100, instance4.foobar)

	// Overridden named transient services should still return different instances
	is.NotSame(instance3, instance4, "Overridden named transient services should return different instances")

	// Test override with different service name
	OverrideNamedTransient(i, "another-service", func(i Injector) (*test, error) {
		return &test{200}, nil
	})

	// Test that original service is still available
	instance5, err5 := InvokeNamed[*test](i, "test-service")
	is.NoError(err5)
	is.Equal(100, instance5.foobar)

	// Test new service
	instance6, err6 := InvokeNamed[*test](i, "another-service")
	is.NoError(err6)
	is.Equal(200, instance6.foobar)
}

func TestInvoke(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct {
		foobar string
	}

	i := New()

	Provide(i, func(i Injector) (test, error) {
		return test{foobar: "foobar"}, nil
	})

	is.Len(i.self.services, 1)

	s0a, ok0a := i.self.services["github.com/samber/do/v2.test"]
	is.True(ok0a)

	s0b, ok0b := s0a.(*serviceLazy[test])
	is.True(ok0b)
	is.False(s0b.built)

	s1, err1 := Invoke[test](i)
	is.NoError(err1)
	if err1 == nil {
		is.Equal("foobar", s1.foobar)
	}

	is.True(s0b.built)

	_, err2 := Invoke[*test](i)
	is.Error(err2)
	is.EqualError(err2, "DI: could not find service `*github.com/samber/do/v2.test`, available services: `github.com/samber/do/v2.test`")

	ProvideNamedValue(i, NameOf[any](), 0)

	_, err3 := Invoke[any](i)
	is.ErrorContains(err3, "type mismatch: invoking `interface {}` but registered `int`")
}

func TestMustInvoke(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	Provide(i, func(i Injector) (test, error) {
		return _test, nil
	})

	is.Len(i.self.services, 1)

	is.Panics(func() {
		_ = MustInvoke[string](i)
	})

	is.NotPanics(func() {
		instance1 := MustInvoke[test](i)
		is.Equal(_test, instance1)
	})
}

func TestInvokeNamed(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	type test struct {
		foobar string
	}
	_test := test{foobar: "foobar"}

	ProvideNamedValue(i, "foobar", 42)
	ProvideNamedValue(i, "hello", _test)

	is.Len(i.self.services, 2)

	service0, err0 := InvokeNamed[string](i, "plop")
	is.Error(err0)
	is.Empty(service0)

	instance1, err1 := InvokeNamed[test](i, "hello")
	is.NoError(err1)
	is.Equal(_test, instance1)
	is.Equal("foobar", instance1.foobar)

	instance1any, err2 := InvokeNamed[any](i, "hello")
	is.NoError(err2)
	is.Equal(instance1, instance1any)

	instance2, err2 := InvokeNamed[int](i, "foobar")
	is.NoError(err2)
	is.Equal(42, instance2)

	instance2any, err2 := InvokeNamed[any](i, "foobar")
	is.NoError(err2)
	is.Equal(instance2, instance2any)

	instance3, err3 := InvokeNamed[string](i, "foobar")
	is.EqualError(err3, "DI: service found, but type mismatch: invoking `string` but registered `int`")
	is.Empty(instance3)
}

func TestMustInvokeNamed(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	ProvideNamedValue(i, "foobar", 42)

	is.Len(i.self.services, 1)

	is.Panics(func() {
		_ = MustInvokeNamed[string](i, "hello")
	})

	is.Panics(func() {
		_ = MustInvokeNamed[string](i, "foobar")
	})

	is.NotPanics(func() {
		instance1 := MustInvokeNamed[int](i, "foobar")
		is.Equal(42, instance1)
	})
}

func TestInvokeStruct(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	ProvideValue(i, &eagerTest{foobar: "foobar"})

	// no dependencies
	test0, err := InvokeStruct[eagerTest](i)
	is.NoError(err)
	is.Empty(test0)

	// not a struct
	test1, err := InvokeStruct[int](i)
	is.Equal("DI: must be a struct or a pointer to a struct, but got `int`", err.Error())
	is.Empty(test1)

	// exported field - generic type
	type hasExportedEagerTestDependency struct {
		EagerTest *eagerTest `do:""`
	}
	test2, err := InvokeStruct[hasExportedEagerTestDependency](i)
	is.NoError(err)
	is.Equal("foobar", test2.EagerTest.foobar)

	// unexported field
	type hasNonExportedEagerTestDependency struct {
		eagerTest *eagerTest `do:""`
	}
	test3, err := InvokeStruct[hasNonExportedEagerTestDependency](i)
	is.NoError(err)
	is.Equal("foobar", test3.eagerTest.foobar)

	// not found
	type dependencyNotFound struct {
		eagerTest *hasNonExportedEagerTestDependency `do:""` //nolint:unused
	}
	test4, err := InvokeStruct[dependencyNotFound](i)
	is.Equal(serviceNotFound(i, ErrServiceNotFound, []string{inferServiceName[*hasNonExportedEagerTestDependency]()}).Error(), err.Error())
	is.Empty(test4)

	// use tag
	type namedDependency struct {
		eagerTest *eagerTest `do:"int"` //nolint:unused
	}
	test5, err := InvokeStruct[namedDependency](i)
	is.Equal(serviceNotFound(i, ErrServiceNotFound, []string{inferServiceName[int]()}).Error(), err.Error())
	is.Empty(test5)

	// named service
	ProvideNamedValue(i, "foobar", 42)
	type namedService struct {
		EagerTest int `do:"foobar"`
	}
	test6, err := InvokeStruct[namedService](i)
	is.NoError(err)
	is.Equal(42, test6.EagerTest)

	// use tag but wrong type
	type namedDependencyButTypeMismatch struct {
		EagerTest *int `do:"*github.com/samber/do/v2.eagerTest"`
	}
	test7, err := InvokeStruct[namedDependencyButTypeMismatch](i)
	is.Equal("DI: `*github.com/samber/do/v2.eagerTest` is not assignable to field `github.com/samber/do/v2.namedDependencyButTypeMismatch.EagerTest`", err.Error())
	is.Empty(test7)

	// use a custom tag
	i = NewWithOpts(&InjectorOpts{StructTagKey: "hello"})
	ProvideNamedValue(i, "foobar", 42)
	type namedServiceWithCustomTag struct {
		EagerTest int `hello:"foobar"`
	}
	test8, err := InvokeStruct[namedServiceWithCustomTag](i)
	is.NoError(err)
	is.Equal(42, test8.EagerTest)

	// assign to interface
	i = New()
	Provide(i, func(i Injector) (Healthchecker, error) {
		return &eagerTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	type serviceWithInterface struct {
		eagerTest Healthchecker `do:""`
	}
	test9, err := InvokeStruct[serviceWithInterface](i)
	is.NoError(err)
	is.NotNil(test9.eagerTest)

	test10, err := InvokeStruct[*serviceWithInterface](i)
	is.NoError(err)
	is.NotNil((*test10).eagerTest) //nolint:gocritic

	test11, err := InvokeStruct[****serviceWithInterface](i)
	is.NoError(err)
	is.NotNil((****test11).eagerTest) //nolint:gocritic
}

func TestMustInvokeStruct(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)
	i := New()

	// use a custom tag
	type namedServiceWithCustomTag struct {
		EagerTest int `do:"foobar"`
	}

	is.Panics(func() {
		_ = MustInvokeStruct[namedServiceWithCustomTag](i)
	})

	ProvideNamedValue(i, "foobar", 42)
	is.NotPanics(func() {
		test := MustInvokeStruct[namedServiceWithCustomTag](i)
		is.Equal(42, test.EagerTest)
	})
}

/////////////////////////////////////////////////////////////////////////////
// 							Explicit aliases
/////////////////////////////////////////////////////////////////////////////

func TestAs(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })

	is.NoError(As[*lazyTestHeathcheckerOK, Healthchecker](i))
	is.EqualError(As[*lazyTestShutdownerOK, Healthchecker](i), "DI: `*github.com/samber/do/v2.lazyTestShutdownerOK` does not implement `github.com/samber/do/v2.Healthchecker`")
	is.EqualError(As[*lazyTestHeathcheckerKO, Healthchecker](i), "DI: service `*github.com/samber/do/v2.lazyTestHeathcheckerKO` has not been declared")
	is.EqualError(As[*lazyTestShutdownerOK, *lazyTestShutdownerOK](i), "DI: service `*github.com/samber/do/v2.lazyTestShutdownerOK` has not been declared")

	i = New()
	type iTestHeathchecker interface {
		HealthCheck() error
	}
	Provide(i, func(i Injector) (iTestHeathchecker, error) { return &lazyTestHeathcheckerOK{}, nil })
	is.NoError(As[iTestHeathchecker, Healthchecker](i))
}

func TestMustAs(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })

	// Test successful alias creation
	is.NotPanics(func() {
		MustAs[*lazyTestHeathcheckerOK, Healthchecker](i)
	})

	// Test that the alias was created and can be invoked
	hc, err := Invoke[Healthchecker](i)
	is.NoError(err)
	is.NotNil(hc)

	// Test panic on invalid interface implementation
	is.Panics(func() {
		MustAs[*lazyTestShutdownerOK, Healthchecker](i)
	})

	// Test panic on service not declared
	is.Panics(func() {
		MustAs[*lazyTestHeathcheckerKO, Healthchecker](i)
	})

	// Test panic on self-reference
	is.Panics(func() {
		MustAs[*lazyTestShutdownerOK, *lazyTestShutdownerOK](i)
	})

	// Test with interface-to-interface
	i = New()
	type iTestHeathchecker interface {
		HealthCheck() error
	}
	Provide(i, func(i Injector) (iTestHeathchecker, error) { return &lazyTestHeathcheckerOK{}, nil })

	is.NotPanics(func() {
		MustAs[iTestHeathchecker, Healthchecker](i)
	})

	// Test that the interface alias works
	hc2, err2 := Invoke[Healthchecker](i)
	is.NoError(err2)
	is.NotNil(hc2)
}

func TestAsNamed(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })

	is.NoError(AsNamed[*lazyTestHeathcheckerOK, Healthchecker](i, "*github.com/samber/do/v2.lazyTestHeathcheckerOK", "github.com/samber/do/v2.Healthchecker"))
	is.EqualError(AsNamed[*lazyTestShutdownerOK, Healthchecker](i, "*github.com/samber/do/v2.lazyTestShutdownerOK", "github.com/samber/do/v2.Healthchecker"), "DI: `*github.com/samber/do/v2.lazyTestShutdownerOK` does not implement `github.com/samber/do/v2.Healthchecker`")
	is.EqualError(AsNamed[*lazyTestHeathcheckerKO, Healthchecker](i, "*github.com/samber/do/v2.lazyTestHeathcheckerKO", "github.com/samber/do/v2.Healthchecker"), "DI: service `*github.com/samber/do/v2.lazyTestHeathcheckerKO` has not been declared")
	is.EqualError(AsNamed[*lazyTestShutdownerOK, *lazyTestShutdownerOK](i, "*github.com/samber/do/v2.lazyTestShutdownerOK", "*github.com/samber/do/v2.lazyTestShutdownerOK"), "DI: service `*github.com/samber/do/v2.lazyTestShutdownerOK` has not been declared")

	i = New()
	Provide(i, func(i Injector) (iTestHeathchecker, error) { return &lazyTestHeathcheckerOK{}, nil })
	is.NoError(AsNamed[iTestHeathchecker, Healthchecker](i, "github.com/samber/do/v2.iTestHeathchecker", "github.com/samber/do/v2.Healthchecker"))
}

func TestMustAsNamed(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })

	// Test successful named alias creation
	is.NotPanics(func() {
		MustAsNamed[*lazyTestHeathcheckerOK, Healthchecker](i, "*github.com/samber/do/v2.lazyTestHeathcheckerOK", "github.com/samber/do/v2.Healthchecker")
	})

	// Test that the named alias was created and can be invoked
	hc, err := InvokeNamed[Healthchecker](i, "github.com/samber/do/v2.Healthchecker")
	is.NoError(err)
	is.NotNil(hc)

	// Test panic on invalid interface implementation
	is.Panics(func() {
		MustAsNamed[*lazyTestShutdownerOK, Healthchecker](i, "*github.com/samber/do/v2.lazyTestShutdownerOK", "github.com/samber/do/v2.Healthchecker")
	})

	// Test panic on service not declared
	is.Panics(func() {
		MustAsNamed[*lazyTestHeathcheckerKO, Healthchecker](i, "*github.com/samber/do/v2.lazyTestHeathcheckerKO", "github.com/samber/do/v2.Healthchecker")
	})

	// Test panic on self-reference
	is.Panics(func() {
		MustAsNamed[*lazyTestShutdownerOK, *lazyTestShutdownerOK](i, "*github.com/samber/do/v2.lazyTestShutdownerOK", "*github.com/samber/do/v2.lazyTestShutdownerOK")
	})

	// Test with interface-to-interface named alias
	i = New()
	type iTestHeathchecker interface {
		HealthCheck() error
	}
	Provide(i, func(i Injector) (iTestHeathchecker, error) { return &lazyTestHeathcheckerOK{}, nil })

	is.NotPanics(func() {
		MustAsNamed[iTestHeathchecker, Healthchecker](i, "github.com/samber/do/v2.iTestHeathchecker", "github.com/samber/do/v2.Healthchecker")
	})

	// Test that the interface named alias works
	hc2, err2 := InvokeNamed[Healthchecker](i, "github.com/samber/do/v2.Healthchecker")
	is.NoError(err2)
	is.NotNil(hc2)

	// Test custom named services
	ProvideNamed(i, "custom-source", func(i Injector) (iTestHeathchecker, error) { return &lazyTestHeathcheckerOK{}, nil })

	is.NotPanics(func() {
		MustAsNamed[iTestHeathchecker, Healthchecker](i, "custom-source", "custom-target")
	})

	hc3, err3 := InvokeNamed[Healthchecker](i, "custom-target")
	is.NoError(err3)
	is.NotNil(hc3)
}

/////////////////////////////////////////////////////////////////////////////
// 							Implicit aliases
/////////////////////////////////////////////////////////////////////////////

func TestInvokeAs(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "hello world"}, nil
	})

	// found
	svc0, err := InvokeAs[*lazyTestHeathcheckerOK](i)
	is.Equal(&lazyTestHeathcheckerOK{foobar: "hello world"}, svc0)
	is.NoError(err)

	// found via interface
	svc1, err := InvokeAs[Healthchecker](i)
	is.EqualValues(&lazyTestHeathcheckerOK{foobar: "hello world"}, svc1)
	is.NoError(err)

	// not found
	svc2, err := InvokeAs[Shutdowner](i)
	is.Empty(svc2)
	is.EqualError(err, "DI: could not find service satisfying interface `github.com/samber/do/v2.Shutdowner`, available services: `*github.com/samber/do/v2.lazyTestHeathcheckerOK`")
}

func TestMustInvokeAs(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "hello world"}, nil
	})

	// Test successful invocation by concrete type
	is.NotPanics(func() {
		svc := MustInvokeAs[*lazyTestHeathcheckerOK](i)
		is.Equal(&lazyTestHeathcheckerOK{foobar: "hello world"}, svc)
	})

	// Test successful invocation by interface
	is.NotPanics(func() {
		svc := MustInvokeAs[Healthchecker](i)
		is.EqualValues(&lazyTestHeathcheckerOK{foobar: "hello world"}, svc)
	})

	// Test panic on interface not found
	is.Panics(func() {
		_ = MustInvokeAs[Shutdowner](i)
	})

	// Test panic on concrete type not found
	is.Panics(func() {
		_ = MustInvokeAs[*lazyTestShutdownerOK](i)
	})

	// Test with multiple services that implement the same interface
	Provide(i, func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "ko service"}, nil
	})

	// Should still work and return the first compatible service
	is.NotPanics(func() {
		svc := MustInvokeAs[Healthchecker](i)
		is.NotNil(svc)
		// Note: The implementation determines which service is returned when multiple match
	})

	// Test with no services at all
	emptyInjector := New()
	is.Panics(func() {
		_ = MustInvokeAs[Healthchecker](emptyInjector)
	})

	is.Panics(func() {
		_ = MustInvokeAs[*lazyTestHeathcheckerOK](emptyInjector)
	})
}

/////////////////////////////////////////////////////////////////////////////
// 							Package-level declaration
/////////////////////////////////////////////////////////////////////////////

func TestPackage(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}
	type iTest interface{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	pkg := Package(
		Lazy(provider1),
		Eager(test{}),
		Bind[*test, iTest](),
	)

	root := New()
	pkg(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())
	svc3 := newEdgeService(root.ID(), root.Name(), NameOf[iTest]())

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_ = MustInvoke[test](root)
		_ = MustInvoke[iTest](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListInvokedServices())
}

func TestNewWithPackage(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}
	type iTest interface{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	pkg := Package(
		Lazy(provider1),
		Eager(test{}),
	)

	root := New(
		pkg,
		Bind[*test, iTest](),
	)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())
	svc3 := newEdgeService(root.ID(), root.Name(), NameOf[iTest]())

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_ = MustInvoke[test](root)
		_ = MustInvoke[iTest](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListInvokedServices())
}

func TestLazy(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	Lazy(provider1)(root)
	Lazy(provider2)(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_, _ = Invoke[test](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1}, root.ListInvokedServices())
}

func TestLazyNamed(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	LazyNamed("p1", provider1)(root)
	LazyNamed("p2", provider2)(root)

	svc1 := newEdgeService(root.ID(), root.Name(), "p1")
	svc2 := newEdgeService(root.ID(), root.Name(), "p2")

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvokeNamed[*test](root, "p1")
		_, _ = InvokeNamed[test](root, "p2")
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1}, root.ListInvokedServices())
}

func TestEager(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}

	root := New()
	Eager(&test{})(root)
	Eager(test{})(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_ = MustInvoke[test](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListInvokedServices())
}

func TestEagerNamed(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}

	root := New()
	EagerNamed("p1", &test{})(root)
	EagerNamed("p2", test{})(root)

	svc1 := newEdgeService(root.ID(), root.Name(), "p1")
	svc2 := newEdgeService(root.ID(), root.Name(), "p2")

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvokeNamed[*test](root, "p1")
		_ = MustInvokeNamed[test](root, "p2")
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListInvokedServices())
}

func TestTransient(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	Transient(provider1)(root)
	Transient(provider2)(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_, _ = Invoke[test](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1}, root.ListInvokedServices())
}

func TestTransientNamed(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	TransientNamed("p1", provider1)(root)
	TransientNamed("p2", provider2)(root)

	svc1 := newEdgeService(root.ID(), root.Name(), "p1")
	svc2 := newEdgeService(root.ID(), root.Name(), "p2")

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvokeNamed[*test](root, "p1")
		_, _ = InvokeNamed[test](root, "p2")
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1}, root.ListInvokedServices())
}

func TestBind(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}
	type iTest interface{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	Lazy(provider1)(root)
	Lazy(provider2)(root)
	Bind[*test, iTest]()(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())
	svc3 := newEdgeService(root.ID(), root.Name(), NameOf[iTest]())

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_, _ = Invoke[test](root)
		_ = MustInvoke[iTest](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc3}, root.ListInvokedServices())
}

func TestBindNamed(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type test struct{}
	type iTest interface{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	Lazy(provider1)(root)
	Lazy(provider2)(root)
	BindNamed[*test, iTest](NameOf[*test](), NameOf[iTest]())(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())
	svc3 := newEdgeService(root.ID(), root.Name(), NameOf[iTest]())

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_, _ = Invoke[test](root)
		_ = MustInvoke[iTest](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc3}, root.ListInvokedServices())
}
