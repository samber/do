package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestServiceNotFound(t *testing.T) {
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
