package do

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvokeAnyByName(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// test default injector vs scope
	i := New()
	ProvideNamedValue(i, "foo", "baz")
	svc, err := invokeAnyByName(i, "foo")
	is.Nil(err)
	is.Equal("baz", svc)

	// service not found
	svc, err = invokeAnyByName(i, "not_found")
	is.NotNil(err)
	is.Empty(svc)
	is.Contains(err.Error(), "DI: could not find service `not_found`, available services: ")

	// test virtual scope wrapper
	called := false
	ProvideNamed(i, "hello", func(ivs Injector) (string, error) {
		// check we received a virtualScope
		vs, ok := ivs.(*virtualScope)
		is.True(ok)
		is.Equal([]string{"hello"}, vs.invokerChain)
		is.NotEqual(i, ivs)

		// create a dependency/dependent relationship
		_, _ = invokeAnyByName(ivs, "foo")

		called = true
		return "foobar", nil
	})
	_, _ = invokeAnyByName(i, "hello")
	is.True(called)
	// check dependency/dependent relationship
	dependencies, dependents := i.dag.explainService(i.self.id, i.self.name, "hello")
	is.ElementsMatch([]EdgeService{{ScopeID: i.self.id, ScopeName: i.self.name, Service: "foo"}}, dependencies)
	is.ElementsMatch([]EdgeService{}, dependents)

	// test circular dependency
	vs := newVirtualScope(i, []string{"foo", "bar"})

	svc, err = invokeAnyByName(vs, "foo")
	is.Error(err)

	is.Empty(svc)
	is.ErrorIs(err, ErrCircularDependency)
	is.EqualError(err, "DI: circular dependency detected: `foo` -> `bar` -> `foo`")

	// @TODO
}

func TestInvokeByName(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// test default injector vs scope
	ProvideNamedValue(DefaultRootScope, "foo", "bar")
	svc, err := invokeByName[string](nil, "foo")
	is.Nil(err)
	is.Equal("bar", svc)

	// test default injector vs scope
	i := New()
	ProvideNamedValue(i, "foo", "baz")
	svc, err = invokeByName[string](i, "foo")
	is.Nil(err)
	is.Equal("baz", svc)

	// service not found
	svc, err = invokeByName[string](nil, "not_found")
	is.NotNil(err)
	is.Empty(svc)
	is.Contains(err.Error(), "DI: could not find service `not_found`, available services: ")

	// test virtual scope wrapper
	called := false
	ProvideNamed(i, "hello", func(ivs Injector) (string, error) {
		// check we received a virtualScope
		vs, ok := ivs.(*virtualScope)
		is.True(ok)
		is.Equal([]string{"hello"}, vs.invokerChain)
		is.NotEqual(i, ivs)

		// create a dependency/dependent relationship
		_, _ = invokeByName[string](ivs, "foo")

		called = true
		return "foobar", nil
	})
	_, _ = invokeByName[string](i, "hello")
	is.True(called)
	// check dependency/dependent relationship
	dependencies, dependents := i.dag.explainService(i.self.id, i.self.name, "hello")
	is.ElementsMatch([]EdgeService{{ScopeID: i.self.id, ScopeName: i.self.name, Service: "foo"}}, dependencies)
	is.ElementsMatch([]EdgeService{}, dependents)

	// test circular dependency
	vs := newVirtualScope(i, []string{"foo", "bar"})

	svc, err = invokeByName[string](vs, "foo")
	is.Error(err)

	is.Empty(svc)
	is.ErrorIs(err, ErrCircularDependency)
	is.EqualError(err, "DI: circular dependency detected: `foo` -> `bar` -> `foo`")

	// @TODO
}

func TestInvokeByGenericType(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// test default injector vs scope
	ProvideValue(DefaultRootScope, &eagerTest{foobar: "foobar"})
	svc1, err := invokeByGenericType[*eagerTest](nil)
	is.EqualValues(&eagerTest{foobar: "foobar"}, svc1)
	is.Nil(err)

	// test default injector vs scope
	i := New()
	ProvideValue(i, &lazyTest{foobar: "baz"})
	svc2, err := invokeByGenericType[*lazyTest](i)
	is.EqualValues(&lazyTest{foobar: "baz"}, svc2)
	is.Nil(err)

	// service not found
	svcX, err := invokeByGenericType[string](i)
	is.Empty(svcX)
	is.NotNil(err)
	is.Contains(err.Error(), "DI: could not find service satisfying interface `string`, available services: `*github.com/samber/do/v2.lazyTest`")

	// test virtual scope wrapper
	called := false
	Provide(i, func(ivs Injector) (*eagerTest, error) {
		// check we received a virtualScope
		vs, ok := ivs.(*virtualScope)
		is.True(ok)
		is.Equal([]string{"*github.com/samber/do/v2.eagerTest"}, vs.invokerChain)
		is.NotEqual(i, ivs)

		// create a dependency/dependent relationship
		_, _ = invokeByGenericType[*lazyTest](ivs)

		called = true
		return &eagerTest{}, nil
	})
	_, _ = invokeByGenericType[*eagerTest](i)
	is.True(called)
	// check dependency/dependent relationship
	dependencies, dependents := i.dag.explainService(i.self.id, i.self.name, "*github.com/samber/do/v2.eagerTest")
	is.ElementsMatch([]EdgeService{{ScopeID: i.self.id, ScopeName: i.self.name, Service: "*github.com/samber/do/v2.lazyTest"}}, dependencies)
	is.ElementsMatch([]EdgeService{}, dependents)

	// test circular dependency
	vs := newVirtualScope(i, []string{"*github.com/samber/do/v2.eagerTest", "bar"})
	svc1, err = invokeByGenericType[*eagerTest](vs)
	is.Error(err)

	is.Empty(svc1)
	is.ErrorIs(err, ErrCircularDependency)

	is.EqualError(err, "DI: circular dependency detected: `*github.com/samber/do/v2.eagerTest` -> `bar` -> `*github.com/samber/do/v2.eagerTest`")

	// @TODO
}

func TestInvokeByName_race(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	injector := New()
	child := injector.Scope("child")

	Provide(injector, func(i Injector) (int, error) {
		time.Sleep(3 * time.Millisecond)
		return 42, nil
	})
	Provide(injector, func(i Injector) (*lazyTest, error) {
		time.Sleep(3 * time.Millisecond)
		return &lazyTest{}, nil
	})

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(j int) {
			_, err1 := invokeByName[int](injector, NameOf[int]())
			_, err2 := invokeByName[*lazyTest](child, NameOf[*lazyTest]())

			is.Nil(err1)
			is.Nil(err2)

			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestInvokeByGenericType_race(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	injector := New()
	child := injector.Scope("child")

	Provide(injector, func(i Injector) (int, error) {
		time.Sleep(3 * time.Millisecond)
		return 42, nil
	})
	Provide(injector, func(i Injector) (*lazyTest, error) {
		time.Sleep(3 * time.Millisecond)
		return &lazyTest{}, nil
	})

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(j int) {
			_, err1 := invokeByGenericType[int](injector)
			_, err2 := invokeByGenericType[*lazyTest](child)

			is.Nil(err1)
			is.Nil(err2)

			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestInvokeByTags(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	ProvideValue(i, &eagerTest{foobar: "foobar"})

	// no dependencies
	err := invokeByTags(i, "*myStruct", reflect.ValueOf(&eagerTest{}), false)
	is.Nil(err)

	// not pointer
	err = invokeByTags(i, "*myStruct", reflect.ValueOf(eagerTest{}), false)
	is.Equal("DI: must be a pointer to a struct", err.Error())

	// exported field - generic type
	type hasExportedEagerTestDependency struct {
		EagerTest *eagerTest `do:""`
	}
	test1 := hasExportedEagerTestDependency{}
	err = invokeByTags(i, "*myStruct", reflect.ValueOf(&test1), false)
	is.Nil(err)
	is.Equal("foobar", test1.EagerTest.foobar)

	// unexported field
	type hasNonExportedEagerTestDependency struct {
		eagerTest *eagerTest `do:""`
	}
	test2 := hasNonExportedEagerTestDependency{}
	err = invokeByTags(i, "*myStruct", reflect.ValueOf(&test2), false)
	is.Nil(err)
	is.Equal("foobar", test2.eagerTest.foobar)

	// not found
	type dependencyNotFound struct {
		eagerTest *hasNonExportedEagerTestDependency `do:""` //nolint:unused
	}
	test3 := dependencyNotFound{}
	err = invokeByTags(i, "*myStruct", reflect.ValueOf(&test3), false)
	is.Equal(serviceNotFound(i, ErrServiceNotFound, []string{inferServiceName[*hasNonExportedEagerTestDependency]()}).Error(), err.Error())

	// use tag
	type namedDependency struct {
		eagerTest *eagerTest `do:"int"` //nolint:unused
	}
	test4 := namedDependency{}
	err = invokeByTags(i, "*myStruct", reflect.ValueOf(&test4), false)
	is.Equal(serviceNotFound(i, ErrServiceNotFound, []string{inferServiceName[int]()}).Error(), err.Error())

	// named service
	ProvideNamedValue(i, "foobar", 42)
	type namedService struct {
		EagerTest int `do:"foobar"`
	}
	test5 := namedService{}
	err = invokeByTags(i, "*myStruct", reflect.ValueOf(&test5), false)
	is.Nil(err)
	is.Equal(42, test5.EagerTest)

	// use tag but wrong type
	type namedDependencyButTypeMismatch struct {
		EagerTest *int `do:"*github.com/samber/do/v2.eagerTest"`
	}
	test6 := namedDependencyButTypeMismatch{}
	err = invokeByTags(i, "*myStruct", reflect.ValueOf(&test6), false)
	is.Equal("DI: `*github.com/samber/do/v2.eagerTest` is not assignable to field `*myStruct.EagerTest`", err.Error())

	// use a custom tag
	i = NewWithOpts(&InjectorOpts{StructTagKey: "hello"})
	ProvideNamedValue(i, "foobar", 42)
	type namedServiceWithCustomTag struct {
		EagerTest int `hello:"foobar"`
	}
	test7 := namedServiceWithCustomTag{}
	err = invokeByTags(i, "*myStruct", reflect.ValueOf(&test7), false)
	is.Nil(err)
	is.Equal(42, test7.EagerTest)
}

func TestInvokeByTags_ImplicitAliasing_FallbackOnNotFound(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	type foobar struct {
		Dep Healthchecker `do:""`
	}

	// First, declare a service with no dependencies

	s := foobar{}
	err := invokeByTags(i, "*foobar", reflect.ValueOf(&s), true)
	is.NotNil(err)
	is.Equal("DI: could not find service `github.com/samber/do/v2.Healthchecker`, no service available", err.Error())
	is.Nil(s.Dep)

	// Now, declare a service with the right type

	Provide(i, func(injector Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})

	s = foobar{}
	err = invokeByTags(i, "*foobar", reflect.ValueOf(&s), true)
	is.Nil(err)
	is.NotNil(s.Dep)
	is.Equal("foobar", s.Dep.(*lazyTestHeathcheckerOK).foobar)

	// Now, declare a service with an interface assignable to interface

	i = New()
	Override(i, func(injector Injector) (iTestHeathchecker, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})

	s = foobar{}
	err = invokeByTags(i, "*foobar", reflect.ValueOf(&s), true)
	is.Nil(err)
	is.NotNil(s.Dep)
	is.Equal("foobar", s.Dep.(*lazyTestHeathcheckerOK).foobar)
}

func TestInvokeByTags_ImplicitAliasing_NoFallbackOnOtherErrors(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()
	Provide(i, func(injector Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})

	// First, lets check the situation the service is not found

	type foobar struct {
		Dep Healthchecker `do:"bad"`
	}

	s := foobar{}
	err := invokeByTags(i, "*foobar", reflect.ValueOf(&s), true)
	is.NotNil(err)
	is.Equal("DI: could not find service `bad`, available services: `*github.com/samber/do/v2.lazyTestHeathcheckerOK`", err.Error())
	is.Nil(s.Dep)

	// Now, declare a service with a wrong type

	ProvideNamed(i, "bad", func(ivs Injector) (*eagerTest, error) {
		return &eagerTest{}, nil
	})

	s = foobar{}
	err = invokeByTags(i, "*foobar", reflect.ValueOf(&s), true)
	is.NotNil(err)
	is.Equal("DI: `bad` is not assignable to field `*foobar.Dep`", err.Error())
	is.Nil(s.Dep)

	// Now, declare a service with the right type, but with provider error

	OverrideNamed(i, "bad", func(ivs Injector) (*eagerTest, error) {
		return nil, fmt.Errorf("boom")
	})

	s = foobar{}
	err = invokeByTags(i, "*foobar", reflect.ValueOf(&s), true)
	is.NotNil(err)
	is.Equal("boom", err.Error())
	is.Nil(s.Dep)
}

func TestServiceNotFound(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")
	child3 := child2a.Scope("child3")

	err := serviceNotFound(child1, ErrServiceNotFound, []string{"not-found"})
	is.Error(err)
	is.ErrorIs(err, ErrServiceNotFound)
	is.EqualError(err, "DI: could not find service `not-found`, no service available")

	err = serviceNotFound(child1, ErrServiceNotFound, []string{"not-found1", "not-found2"})
	is.Error(err)
	is.ErrorIs(err, ErrServiceNotFound)
	is.EqualError(err, "DI: could not find service `not-found2`, no service available, path: `not-found1` -> `not-found2`")

	rootScope.serviceSet("root-a", newServiceLazy("root-a", func(i Injector) (int, error) { return 0, nil }))
	child1.serviceSet("child1-a", newServiceLazy("child1-a", func(i Injector) (int, error) { return 1, nil }))
	child2a.serviceSet("child2a-a", newServiceLazy("child2a-a", func(i Injector) (int, error) { return 2, nil }))
	child2a.serviceSet("child2a-b", newServiceLazy("child2a-b", func(i Injector) (int, error) { return 3, nil }))
	child2b.serviceSet("child2b-a", newServiceLazy("child2b-a", func(i Injector) (int, error) { return 4, nil }))
	child3.serviceSet("child3-a", newServiceLazy("child3-a", func(i Injector) (int, error) { return 5, nil }))

	err = serviceNotFound(child1, ErrServiceNotFound, []string{"not-found"})
	is.Error(err)
	is.ErrorIs(err, ErrServiceNotFound)
	is.EqualError(err, "DI: could not find service `not-found`, available services: `child1-a`, `root-a`")

	err = serviceNotFound(child1, ErrServiceNotFound, []string{"not-found1", "not-found2"})
	is.Error(err)
	is.ErrorIs(err, ErrServiceNotFound)
	is.EqualError(err, "DI: could not find service `not-found2`, available services: `child1-a`, `root-a`, path: `not-found1` -> `not-found2`")

	// @TODO: test service ordering
}

func TestHandleProviderPanic(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	is.NotPanics(func() {
		// ok
		svc, err := handleProviderPanic(func(i Injector) (int, error) {
			return 42, nil
		}, nil)
		is.Nil(err)
		is.Equal(42, svc)

		// should return an error
		svc, err = handleProviderPanic(func(i Injector) (int, error) {
			return 0, assert.AnError
		}, nil)
		is.Equal(assert.AnError, err)
		is.Equal(0, svc)

		// shoud not return 42
		svc, err = handleProviderPanic(func(i Injector) (int, error) {
			return 42, assert.AnError
		}, nil)
		is.Equal(assert.AnError, err)
		is.Equal(0, svc)

		// panics with string
		svc, err = handleProviderPanic(func(i Injector) (int, error) {
			panic("aïe")
		}, nil)
		is.EqualError(err, "DI: aïe")
		is.Equal(0, svc)

		// panics with error
		svc, err = handleProviderPanic(func(i Injector) (int, error) {
			panic(assert.AnError)
		}, nil)
		is.Equal(assert.AnError, err)
		is.Equal(0, svc)
	})
}
