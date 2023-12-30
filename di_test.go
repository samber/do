package do

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameOf(t *testing.T) {
	is := assert.New(t)

	is.Equal("int", NameOf[int]())
	is.Equal("github.com/samber/do/v2.eagerTest", NameOf[eagerTest]())
	is.Equal("*github.com/samber/do/v2.eagerTest", NameOf[*eagerTest]())
	is.Equal("*map[int]bool", NameOf[*map[int]bool]())
	is.Equal("*github.com/samber/do/v2.Service[int]", NameOf[Service[int]]())
}

func TestProvide(t *testing.T) {
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
		s, ok := s1.(Service[*test])
		is.True(ok)
		if ok {
			is.Equal("*github.com/samber/do/v2.test", s.getName())
		}
	}

	s2, ok2 := i.self.services["github.com/samber/do/v2.test"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(Service[test])
		is.True(ok)
		if ok {
			is.Equal("github.com/samber/do/v2.test", s.getName())
		}
	}

	_, ok3 := i.self.services["github.com/samber/do/v2.*plop"]
	is.False(ok3)

	// @TODO: check that all services share the same references
}

func TestProvideNamed(t *testing.T) {
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
		s, ok := s1.(Service[*test])
		is.True(ok)
		if ok {
			is.Equal("*foobar", s.getName())
		}
	}

	s2, ok2 := i.self.services["foobar"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(Service[test])
		is.True(ok)
		if ok {
			is.Equal("foobar", s.getName())
		}
	}

	_, ok3 := i.self.services["*do.plop"]
	is.False(ok3)

	// @TODO: check that all services share the same references
}

func TestProvideValue(t *testing.T) {
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
		s, ok := s1.(Service[int])
		is.True(ok)
		if ok {
			is.Equal("int", s.getName())
			instance, err := s.getInstance(i)
			is.EqualValues(42, instance)
			is.Nil(err)
		}
	}

	s2, ok2 := i.self.services["github.com/samber/do/v2.test"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(Service[test])
		is.True(ok)
		if ok {
			is.Equal("github.com/samber/do/v2.test", s.getName())
			instance, err := s.getInstance(i)
			is.EqualValues(_test, instance)
			is.Nil(err)
		}
	}

	// @TODO: check that all services share the same references
}

func TestProvideNamedValue(t *testing.T) {
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
		s, ok := s1.(Service[int])
		is.True(ok)
		if ok {
			is.Equal("foobar", s.getName())
			instance, err := s.getInstance(i)
			is.EqualValues(42, instance)
			is.Nil(err)
		}
	}

	s2, ok2 := i.self.services["hello"]
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

	// @TODO: check that all services share the same references
}

func TestProvideTransient(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	i := New()

	ProvideTransient(i, func(i Injector) (*test, error) {
		return &test{}, nil
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
		s, ok := s1.(Service[*test])
		is.True(ok)
		if ok {
			is.Equal("*github.com/samber/do/v2.test", s.getName())
		}
	}

	s2, ok2 := i.self.services["github.com/samber/do/v2.test"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(Service[test])
		is.True(ok)
		if ok {
			is.Equal("github.com/samber/do/v2.test", s.getName())
		}
	}

	_, ok3 := i.self.services["github.com/samber/do/v2.*plop"]
	is.False(ok3)

	// @TODO: check that all services share the same references
}
func TestProvideNamedTransient(t *testing.T) {
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
		s, ok := s1.(Service[*test])
		is.True(ok)
		if ok {
			is.Equal("*foobar", s.getName())
		}
	}

	s2, ok2 := i.self.services["foobar"]
	is.True(ok2)
	if ok2 {
		s, ok := s2.(Service[test])
		is.True(ok)
		if ok {
			is.Equal("foobar", s.getName())
		}
	}

	_, ok3 := i.self.services["*do.plop"]
	is.False(ok3)

	// @TODO: check that all services share the same references
}

func TestProvide_race(t *testing.T) {
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

		// OverrideNamed(i, "*github.com/samber/do/v2.test", func(i Injector) (*test, error) {
		// 	return &test{2}, nil
		// })
		// is.Equal(2, MustInvoke[*test](i).foobar)

		// OverrideValue(i, &test{3})
		// is.Equal(3, MustInvoke[*test](i).foobar)

		// OverrideNamedValue(i, "*github.com/samber/do/v2.test", &test{4})
		// is.Equal(4, MustInvoke[*test](i).foobar)
	})
}

func TestOverrideNamed(t *testing.T) {
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
	// @TODO
}

func TestOverrideNamedTransient(t *testing.T) {
	// @TODO
}

func TestInvoke(t *testing.T) {
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
	is.Nil(err1)
	if err1 == nil {
		is.Equal("foobar", s1.foobar)
	}

	is.True(s0b.built)

	_, err2 := Invoke[*test](i)
	is.NotNil(err2)
	is.Errorf(err2, "do: service not found")
}

func TestMustInvoke(t *testing.T) {
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
		is.EqualValues(_test, instance1)
	})
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

	is.Len(i.self.services, 2)

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

func TestMustInvokeNamed(t *testing.T) {
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
		is.EqualValues(42, instance1)
	})
}

func TestInvokeStruct(t *testing.T) {
	is := assert.New(t)

	i := New()
	ProvideValue(i, &eagerTest{foobar: "foobar"})

	// no dependencies
	test0, err := InvokeStruct[eagerTest](i)
	is.Nil(err)
	is.Empty(test0)

	// not a struct
	test1, err := InvokeStruct[int](i)
	is.Nil(test1)
	is.Equal("DI: not a struct", err.Error())

	// exported field - generic type
	type hasExportedEagerTestDependency struct {
		EagerTest *eagerTest `do:""`
	}
	test2, err := InvokeStruct[hasExportedEagerTestDependency](i)
	is.Nil(err)
	is.Equal("foobar", test2.EagerTest.foobar)

	// unexported field
	type hasNonExportedEagerTestDependency struct {
		eagerTest *eagerTest `do:""`
	}
	test3, err := InvokeStruct[hasNonExportedEagerTestDependency](i)
	is.Nil(err)
	is.Equal("foobar", test3.eagerTest.foobar)

	// not found
	type dependencyNotFound struct {
		eagerTest *hasNonExportedEagerTestDependency `do:""`
	}
	test4, err := InvokeStruct[dependencyNotFound](i)
	is.Equal(serviceNotFound(i, inferServiceName[*hasNonExportedEagerTestDependency]()).Error(), err.Error())
	is.Nil(test4)

	// use tag
	type namedDependency struct {
		eagerTest *eagerTest `do:"int"`
	}
	test5, err := InvokeStruct[namedDependency](i)
	is.Equal(serviceNotFound(i, inferServiceName[int]()).Error(), err.Error())
	is.Nil(test5)

	// named service
	ProvideNamedValue(i, "foobar", 42)
	type namedService struct {
		EagerTest int `do:"foobar"`
	}
	test6, err := InvokeStruct[namedService](i)
	is.Nil(err)
	is.Equal(42, test6.EagerTest)

	// use tag but wrong type
	type namedDependencyButTypeMismatch struct {
		EagerTest *int `do:"*github.com/samber/do/v2.eagerTest"`
	}
	test7, err := InvokeStruct[namedDependencyButTypeMismatch](i)
	is.Equal("DI: field 'EagerTest' is not assignable to service *github.com/samber/do/v2.eagerTest", err.Error())
	is.Nil(test7)

	// use a custom tag
	i = NewWithOpts(&InjectorOpts{StructTagKey: "hello"})
	ProvideNamedValue(i, "foobar", 42)
	type namedServiceWithCustomTag struct {
		EagerTest int `hello:"foobar"`
	}
	test8, err := InvokeStruct[namedServiceWithCustomTag](i)
	is.Nil(err)
	is.Equal(42, test8.EagerTest)
}

func TestMustInvokeStruct(t *testing.T) {
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
