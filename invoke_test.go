package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvokeImplem(t *testing.T) {
	is := assert.New(t)

	// test default injector vs scope
	ProvideNamedValue(DefaultRootScope, "foo", "bar")
	svc, err := invoke[string](nil, "foo")
	is.Equal("bar", svc)
	is.Nil(err)

	// test default injector vs scope
	i := New()
	ProvideNamedValue(i, "foo", "baz")
	svc, err = invoke[string](i, "foo")
	is.Equal("baz", svc)
	is.Nil(err)

	// service not found
	svc, err = invoke[string](nil, "not_found")
	is.Empty(svc)
	is.NotNil(err)
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
		_, _ = invoke[string](ivs, "foo")

		called = true
		return "foobar", nil
	})
	_, _ = invoke[string](i, "hello")
	is.True(called)
	// check dependency/dependent relationship
	dependencies, dependents := i.dag.explainService(i.self.id, i.self.name, "hello")
	is.ElementsMatch([]EdgeService{{ScopeID: i.self.id, ScopeName: i.self.name, Service: "foo"}}, dependencies)
	is.ElementsMatch([]EdgeService{}, dependents)

	// test circular dependency
	vs := virtualScope{invokerChain: []string{"foo", "bar"}, self: i}
	svc, err = invoke[string](&vs, "foo")
	is.Empty(svc)
	is.Error(err)
	is.EqualError(err, "DI: circular dependency detected: foo -> bar -> foo")

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

	rootScope.serviceSet("root-a", newServiceLazy("root-a", func(i Injector) (int, error) { return 0, nil }))
	child1.serviceSet("child1-a", newServiceLazy("child1-a", func(i Injector) (int, error) { return 1, nil }))
	child2a.serviceSet("child2a-a", newServiceLazy("child2a-a", func(i Injector) (int, error) { return 2, nil }))
	child2a.serviceSet("child2a-b", newServiceLazy("child2a-b", func(i Injector) (int, error) { return 3, nil }))
	child2b.serviceSet("child2b-a", newServiceLazy("child2b-a", func(i Injector) (int, error) { return 4, nil }))
	child3.serviceSet("child3-a", newServiceLazy("child3-a", func(i Injector) (int, error) { return 5, nil }))

	is.EqualError(serviceNotFound(child1, "not-found"), "DI: could not find service `not-found`, available services: `child1-a`, `root-a`")
}
