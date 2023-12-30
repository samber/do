package do

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvokeAnyByName(t *testing.T) {
	is := assert.New(t)

	// test default injector vs scope
	i := New()
	ProvideNamedValue(i, "foo", "baz")
	svc, err := invokeAnyByName(i, "foo")
	is.Nil(err)
	is.Equal("baz", svc)

	// service not found
	svc, err = invokeAnyByName(nil, "not_found")
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
	vs := virtualScope{invokerChain: []string{"foo", "bar"}, self: i}

	svc, err = invokeAnyByName(&vs, "foo")
	is.Error(err)

	is.Empty(svc)
	is.ErrorIs(err, ErrCircularDependency)
	is.EqualError(err, "DI: circular dependency detected: foo -> bar -> foo")

	// @TODO
}

func TestInvokeByName(t *testing.T) {
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
	vs := virtualScope{invokerChain: []string{"foo", "bar"}, self: i}

	svc, err = invokeByName[string](&vs, "foo")
	is.Error(err)

	is.Empty(svc)
	is.ErrorIs(err, ErrCircularDependency)
	is.EqualError(err, "DI: circular dependency detected: foo -> bar -> foo")

	// @TODO
}

func TestInvokeByGenericType(t *testing.T) {
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
	is.Contains(err.Error(), "DI: could not find service `string`, available services: `*github.com/samber/do/v2.lazyTest`")

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
	vs := virtualScope{invokerChain: []string{"*github.com/samber/do/v2.eagerTest", "bar"}, self: i}
	svc1, err = invokeByGenericType[*eagerTest](&vs)
	is.Error(err)

	is.Empty(svc1)
	is.ErrorIs(err, ErrCircularDependency)

	is.EqualError(err, "DI: circular dependency detected: *github.com/samber/do/v2.eagerTest -> bar -> *github.com/samber/do/v2.eagerTest")

	// @TODO
}

func TestInvokeByName_race(t *testing.T) {
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
	is := assert.New(t)

	i := New()
	ProvideValue(i, &eagerTest{foobar: "foobar"})

	// no dependencies
	err := invokeByTags(i, reflect.ValueOf(&eagerTest{}))
	is.Nil(err)

	// not pointer
	err = invokeByTags(i, reflect.ValueOf(eagerTest{}))
	is.Equal("DI: not a pointer", err.Error())

	// exported field - generic type
	type hasExportedEagerTestDependency struct {
		EagerTest *eagerTest `do:""`
	}
	test1 := hasExportedEagerTestDependency{}
	err = invokeByTags(i, reflect.ValueOf(&test1))
	is.Nil(err)
	is.Equal("foobar", test1.EagerTest.foobar)

	// unexported field
	type hasNonExportedEagerTestDependency struct {
		eagerTest *eagerTest `do:""`
	}
	test2 := hasNonExportedEagerTestDependency{}
	err = invokeByTags(i, reflect.ValueOf(&test2))
	is.Nil(err)
	is.Equal("foobar", test2.eagerTest.foobar)

	// not found
	type dependencyNotFound struct {
		eagerTest *hasNonExportedEagerTestDependency `do:""`
	}
	test3 := dependencyNotFound{}
	err = invokeByTags(i, reflect.ValueOf(&test3))
	is.Equal(serviceNotFound(i, inferServiceName[*hasNonExportedEagerTestDependency]()).Error(), err.Error())

	// use tag
	type namedDependency struct {
		eagerTest *eagerTest `do:"int"`
	}
	test4 := namedDependency{}
	err = invokeByTags(i, reflect.ValueOf(&test4))
	is.Equal(serviceNotFound(i, inferServiceName[int]()).Error(), err.Error())

	// named service
	ProvideNamedValue(i, "foobar", 42)
	type namedService struct {
		EagerTest int `do:"foobar"`
	}
	test5 := namedService{}
	err = invokeByTags(i, reflect.ValueOf(&test5))
	is.Nil(err)
	is.Equal(42, test5.EagerTest)

	// use tag but wrong type
	type namedDependencyButTypeMismatch struct {
		EagerTest *int `do:"*github.com/samber/do/v2.eagerTest"`
	}
	test6 := namedDependencyButTypeMismatch{}
	err = invokeByTags(i, reflect.ValueOf(&test6))
	is.Equal("DI: field 'EagerTest' is not assignable to service *github.com/samber/do/v2.eagerTest", err.Error())

	// use a custom tag
	i = NewWithOpts(&InjectorOpts{StructTagKey: "hello"})
	ProvideNamedValue(i, "foobar", 42)
	type namedServiceWithCustomTag struct {
		EagerTest int `hello:"foobar"`
	}
	test7 := namedServiceWithCustomTag{}
	err = invokeByTags(i, reflect.ValueOf(&test7))
	is.Nil(err)
	is.Equal(42, test7.EagerTest)
}

func TestServiceNotFound(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")
	child3 := child2a.Scope("child3")

	err := serviceNotFound(child1, "not-found")
	is.Error(err)

	is.ErrorIs(err, ErrServiceNotFound)
	is.EqualError(err, "DI: could not find service `not-found`, no service available")

	rootScope.serviceSet("root-a", newServiceLazy("root-a", func(i Injector) (int, error) { return 0, nil }))
	child1.serviceSet("child1-a", newServiceLazy("child1-a", func(i Injector) (int, error) { return 1, nil }))
	child2a.serviceSet("child2a-a", newServiceLazy("child2a-a", func(i Injector) (int, error) { return 2, nil }))
	child2a.serviceSet("child2a-b", newServiceLazy("child2a-b", func(i Injector) (int, error) { return 3, nil }))
	child2b.serviceSet("child2b-a", newServiceLazy("child2b-a", func(i Injector) (int, error) { return 4, nil }))
	child3.serviceSet("child3-a", newServiceLazy("child3-a", func(i Injector) (int, error) { return 5, nil }))

	err = serviceNotFound(child1, "not-found")
	is.Error(err)

	is.ErrorIs(err, ErrServiceNotFound)
	is.EqualError(err, "DI: could not find service `not-found`, available services: `child1-a`, `root-a`")

	// @TODO: test service ordering
}
