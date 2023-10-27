package do

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScope(t *testing.T) {
	is := assert.New(t)

	parentScope := &Scope{name: "[root]"}
	rootScope := &RootScope{self: parentScope}

	scope := newScope("foobar", rootScope, parentScope)
	is.NotEmpty(scope.id)
	is.Equal("foobar", scope.name)
	is.Equal(rootScope, scope.rootScope)
	is.Equal(parentScope, scope.parentScope)
	is.Equal(map[string]*Scope{}, scope.childScopes)
	is.Equal(make(map[string]any), scope.services)
	is.Equal(map[string]int{}, scope.orderedInvocation)
	is.Equal(0, scope.orderedInvocationIndex)
}

func TestScope_ID(t *testing.T) {
	is := assert.New(t)

	scope := newScope("foobar", nil, nil)

	is.NotEmpty(scope.ID())
	is.Len(scope.ID(), 36)
}

func TestScope_Name(t *testing.T) {
	is := assert.New(t)

	scope := newScope("foobar", nil, nil)

	is.NotEmpty(scope.Name())
	is.Equal(scope.Name(), "foobar")
}

func TestScope_Scope(t *testing.T) {
	is := assert.New(t)

	parentScope := &Scope{name: "[root]"}
	rootScope := &RootScope{self: parentScope}

	scope := newScope("foobar", rootScope, parentScope)

	// create children
	child1 := scope.Scope("child1")
	child2 := scope.Scope("child2")
	is.Panics(func() {
		scope.Scope("child2")
	})

	// parent
	is.Len(scope.childScopes, 2)
	is.Equal(child1, scope.childScopes["child1"])
	is.Equal(child2, scope.childScopes["child2"])

	// child 1
	is.Equal("child1", child1.name)
	is.Equal(rootScope, child1.rootScope)
	is.Equal(scope, child1.parentScope)
	is.Equal(map[string]*Scope{}, child1.childScopes)
	is.Equal(make(map[string]any), child1.services)
	is.Equal(map[string]int{}, child1.orderedInvocation)
	is.Equal(0, child1.orderedInvocationIndex)

	// child 2
	is.Equal("child2", child2.name)
	is.Equal(rootScope, child2.rootScope)
	is.Equal(scope, child2.parentScope)
	is.Equal(map[string]*Scope{}, child2.childScopes)
	is.Equal(make(map[string]any), child2.services)
	is.Equal(map[string]int{}, child2.orderedInvocation)
	is.Equal(0, child2.orderedInvocationIndex)
}

func TestScope_RootScope(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")
	child3 := child2a.Scope("child3")

	is.Equal(rootScope, child1.RootScope())
	is.Equal(rootScope, child2a.RootScope())
	is.Equal(rootScope, child2b.RootScope())
	is.Equal(rootScope, child3.RootScope())
}

func TestScope_Ancestors(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")
	child3 := child2a.Scope("child3")

	is.Equal([]*Scope{rootScope.self}, child1.Ancestors())
	is.Equal([]*Scope{child1, rootScope.self}, child2a.Ancestors())
	is.Equal([]*Scope{child1, rootScope.self}, child2b.Ancestors())
	is.Equal([]*Scope{child2a, child1, rootScope.self}, child3.Ancestors())
}

func TestScope_Children(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")
	child3 := child2a.Scope("child3")

	is.ElementsMatch([]*Scope{child1}, rootScope.Children())
	is.ElementsMatch([]*Scope{child2a, child2b}, child1.Children())
	is.ElementsMatch([]*Scope{child3}, child2a.Children())
	is.ElementsMatch([]*Scope{}, child2b.Children())
	is.ElementsMatch([]*Scope{}, child3.Children())
}

func TestScope_ChildByID(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2 := rootScope.Scope("child2")
	child2a := child2.Scope("child2a")

	rootScope.self.id = "[root]"
	child1.id = "child1"
	child2.id = "child2"
	child2a.id = "child2a"

	// from root POV
	is.NotPanics(func() {
		s, ok := rootScope.ChildByID("[root]")
		is.False(ok)
		is.Nil(s)
		s, ok = rootScope.ChildByID("child1")
		is.True(ok)
		is.Equal(child1.ID(), s.ID())
		s, ok = rootScope.ChildByID("child2")
		is.True(ok)
		is.Equal(child2.ID(), s.ID())
		s, ok = rootScope.ChildByID("child2a")
		is.True(ok)
		is.Equal(child2a.ID(), s.ID())
	})

	// from child1 POV
	is.NotPanics(func() {
		s, ok := child1.ChildByID("[root]")
		is.False(ok)
		is.Nil(s)
		s, ok = child1.ChildByID("child1")
		is.False(ok)
		is.Nil(s)
		s, ok = child1.ChildByID("child2")
		is.False(ok)
		is.Nil(s)
		s, ok = child1.ChildByID("child2a")
		is.False(ok)
		is.Nil(s)
	})

	// from child2 POV
	is.NotPanics(func() {
		s, ok := child2.ChildByID("[root]")
		is.False(ok)
		is.Nil(s)
		s, ok = child2.ChildByID("child1")
		is.False(ok)
		is.Nil(s)
		s, ok = child2.ChildByID("child2")
		is.False(ok)
		is.Nil(s)
		s, ok = child2.ChildByID("child2a")
		is.True(ok)
		is.Equal(child2a.ID(), s.ID())
	})

	// from child2a POV
	is.NotPanics(func() {
		s, ok := child2a.ChildByID("[root]")
		is.False(ok)
		is.Nil(s)
		s, ok = child2a.ChildByID("child1")
		is.False(ok)
		is.Nil(s)
		s, ok = child2a.ChildByID("child2")
		is.False(ok)
		is.Nil(s)
		s, ok = child2a.ChildByID("child2a")
		is.False(ok)
		is.Nil(s)
	})
}

func TestScope_ChildByName(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2 := rootScope.Scope("child2")
	child2a := child2.Scope("child2a")

	// from root POV
	is.NotPanics(func() {
		s, ok := rootScope.ChildByName("[root]")
		is.False(ok)
		is.Nil(s)
		s, ok = rootScope.ChildByName("child1")
		is.True(ok)
		is.Equal(child1.ID(), s.ID())
		s, ok = rootScope.ChildByName("child2")
		is.True(ok)
		is.Equal(child2.ID(), s.ID())
		s, ok = rootScope.ChildByName("child2a")
		is.True(ok)
		is.Equal(child2a.ID(), s.ID())
	})

	// from child1 POV
	is.NotPanics(func() {
		s, ok := child1.ChildByName("[root]")
		is.False(ok)
		is.Nil(s)
		s, ok = child1.ChildByName("child1")
		is.False(ok)
		is.Nil(s)
		s, ok = child1.ChildByName("child2")
		is.False(ok)
		is.Nil(s)
		s, ok = child1.ChildByName("child2a")
		is.False(ok)
		is.Nil(s)
	})

	// from child2 POV
	is.NotPanics(func() {
		s, ok := child2.ChildByName("[root]")
		is.False(ok)
		is.Nil(s)
		s, ok = child2.ChildByName("child1")
		is.False(ok)
		is.Nil(s)
		s, ok = child2.ChildByName("child2")
		is.False(ok)
		is.Nil(s)
		s, ok = child2.ChildByName("child2a")
		is.True(ok)
		is.Equal(child2a.ID(), s.ID())
	})

	// from child2a POV
	is.NotPanics(func() {
		s, ok := child2a.ChildByName("[root]")
		is.False(ok)
		is.Nil(s)
		s, ok = child2a.ChildByName("child1")
		is.False(ok)
		is.Nil(s)
		s, ok = child2a.ChildByName("child2")
		is.False(ok)
		is.Nil(s)
		s, ok = child2a.ChildByName("child2a")
		is.False(ok)
		is.Nil(s)
	})
}

func TestScope_ListProvidedServices(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")
	child3 := child2a.Scope("child3")

	rootScope.serviceSet("root-a", newServiceEager("root-a", 0))
	child1.serviceSet("child1-a", newServiceEager("child1-a", 1))
	child2a.serviceSet("child2a-a", newServiceEager("child2a-a", 2))
	child2a.serviceSet("child2a-b", newServiceEager("child2a-b", 3))
	child2b.serviceSet("child2b-a", newServiceEager("child2b-a", 4))
	child3.serviceSet("child3-a", newServiceEager("child3-a", 5))

	is.ElementsMatch([]EdgeService{newEdgeService(rootScope.ID(), rootScope.Name(), "root-a")}, rootScope.ListProvidedServices())
	is.ElementsMatch([]EdgeService{newEdgeService(child1.ID(), child1.Name(), "child1-a"), newEdgeService(rootScope.ID(), rootScope.Name(), "root-a")}, child1.ListProvidedServices())
	is.ElementsMatch([]EdgeService{newEdgeService(child2a.ID(), child2a.Name(), "child2a-a"), newEdgeService(child2a.ID(), child2a.Name(), "child2a-b"), newEdgeService(child1.ID(), child1.Name(), "child1-a"), newEdgeService(rootScope.ID(), rootScope.Name(), "root-a")}, child2a.ListProvidedServices())
	is.ElementsMatch([]EdgeService{newEdgeService(child2b.ID(), child2b.Name(), "child2b-a"), newEdgeService(child1.ID(), child1.Name(), "child1-a"), newEdgeService(rootScope.ID(), rootScope.Name(), "root-a")}, child2b.ListProvidedServices())
	is.ElementsMatch([]EdgeService{newEdgeService(child3.ID(), child3.Name(), "child3-a"), newEdgeService(child2a.ID(), child2a.Name(), "child2a-a"), newEdgeService(child2a.ID(), child2a.Name(), "child2a-b"), newEdgeService(child1.ID(), child1.Name(), "child1-a"), newEdgeService(rootScope.ID(), rootScope.Name(), "root-a")}, child3.ListProvidedServices())
}

func TestScope_ListInvokedServices(t *testing.T) {
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

	// invokeImplem[int](rootScope, "root-a")	// root-a is not invoked
	_, _ = invoke[int](child1, "child1-a")
	_, _ = invoke[int](child2a, "child2a-a")
	_, _ = invoke[int](child2a, "child2a-b")
	_, _ = invoke[int](child2b, "child2b-a")
	_, _ = invoke[int](child3, "child3-a")

	is.ElementsMatch([]EdgeService{}, rootScope.ListInvokedServices())
	is.ElementsMatch([]EdgeService{newEdgeService(child1.ID(), child1.Name(), "child1-a")}, child1.ListInvokedServices())
	is.ElementsMatch([]EdgeService{newEdgeService(child2a.ID(), child2a.Name(), "child2a-a"), newEdgeService(child2a.ID(), child2a.Name(), "child2a-b"), newEdgeService(child1.ID(), child1.Name(), "child1-a")}, child2a.ListInvokedServices())
	is.ElementsMatch([]EdgeService{newEdgeService(child2b.ID(), child2b.Name(), "child2b-a"), newEdgeService(child1.ID(), child1.Name(), "child1-a")}, child2b.ListInvokedServices())
	is.ElementsMatch([]EdgeService{newEdgeService(child3.ID(), child3.Name(), "child3-a"), newEdgeService(child2a.ID(), child2a.Name(), "child2a-a"), newEdgeService(child2a.ID(), child2a.Name(), "child2a-b"), newEdgeService(child1.ID(), child1.Name(), "child1-a")}, child3.ListInvokedServices())

	is.Equal(map[string]int{}, rootScope.self.orderedInvocation)
	is.Equal(map[string]int{"child1-a": 0}, child1.orderedInvocation)
	is.Equal(map[string]int{"child2a-a": 0, "child2a-b": 1}, child2a.orderedInvocation)
	is.Equal(map[string]int{"child2b-a": 0}, child2b.orderedInvocation)
	is.Equal(map[string]int{"child3-a": 0}, child3.orderedInvocation)
}

func TestScope_HealthCheck(t *testing.T) {
	// @TODO
}

// @TODO: missing tests for context
func TestScope_HealthCheckWithContext(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")
	child3 := child2a.Scope("child3")

	provider1 := func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	}
	provider2 := func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	}

	rootScope.serviceSet("root-a", newServiceLazy("root-a", provider2))
	child1.serviceSet("child1-a", newServiceLazy("child1-a", provider1))
	child2a.serviceSet("child2a-a", newServiceLazy("child2a-a", provider1))
	child2a.serviceSet("child2a-b", newServiceLazy("child2a-b", provider2))
	child2b.serviceSet("child2b-a", newServiceLazy("child2b-a", provider2))
	child3.serviceSet("child3-a", newServiceLazy("child3-a", provider2))

	is.EqualValues(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": nil, "root-a": nil}, child2a.HealthCheck())

	_, _ = invoke[*lazyTestHeathcheckerKO](rootScope, "root-a")
	_, _ = invoke[*lazyTestHeathcheckerOK](child1, "child1-a")
	_, _ = invoke[*lazyTestHeathcheckerOK](child2a, "child2a-a")
	_, _ = invoke[*lazyTestHeathcheckerKO](child2a, "child2a-b")
	_, _ = invoke[*lazyTestHeathcheckerKO](child2b, "child2b-a")
	_, _ = invoke[*lazyTestHeathcheckerKO](child3, "child3-a")

	is.EqualValues(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": assert.AnError, "root-a": assert.AnError}, child2a.HealthCheck())
	is.EqualValues(map[string]error{"root-a": assert.AnError}, rootScope.HealthCheckWithContext(context.Background()))
}

func TestScope_Shutdown(t *testing.T) {
	// @TODO
}

// @TODO: missing tests for context
func TestScope_ShutdownWithContext(t *testing.T) {
	is := assert.New(t)

	ctx := context.Background()
	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")

	provider1 := func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	}
	provider2 := func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	}

	rootScope.serviceSet("root-a", newServiceLazy("root-a", provider2))
	child1.serviceSet("child1-a", newServiceLazy("child1-a", provider1))
	child2a.serviceSet("child2a-a", newServiceLazy("child2a-a", provider1))
	child2a.serviceSet("child2a-b", newServiceLazy("child2a-b", provider2))
	child2b.serviceSet("child2b-a", newServiceLazy("child2b-a", provider2))

	_, _ = invoke[*lazyTestHeathcheckerKO](rootScope, "root-a")
	_, _ = invoke[*lazyTestHeathcheckerOK](child1, "child1-a")
	_, _ = invoke[*lazyTestHeathcheckerOK](child2a, "child2a-a")
	_, _ = invoke[*lazyTestHeathcheckerKO](child2a, "child2a-b")
	_, _ = invoke[*lazyTestHeathcheckerKO](child2b, "child2b-a")

	// from rootScope POV
	is.Equal(assert.AnError, rootScope.serviceHealthCheck(ctx, "root-a"))
	is.ErrorContains(rootScope.serviceHealthCheck(ctx, "child1-a"), "could not find service")
	is.ErrorContains(rootScope.serviceHealthCheck(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(rootScope.serviceHealthCheck(ctx, "child2a-b"), "could not find service")
	is.ErrorContains(rootScope.serviceHealthCheck(ctx, "child2b-a"), "could not find service")

	// from child1 POV
	is.ErrorContains(child1.serviceHealthCheck(ctx, "root-a"), "could not find service")
	is.Equal(nil, child1.serviceHealthCheck(ctx, "child1-a"))
	is.ErrorContains(child1.serviceHealthCheck(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(child1.serviceHealthCheck(ctx, "child2a-b"), "could not find service")
	is.ErrorContains(child1.serviceHealthCheck(ctx, "child2b-a"), "could not find service")

	// from child2a POV
	is.ErrorContains(child2a.serviceHealthCheck(ctx, "root-a"), "could not find service")
	is.ErrorContains(child2a.serviceHealthCheck(ctx, "child1-a"), "could not find service")
	is.Equal(nil, child2a.serviceHealthCheck(ctx, "child2a-a"))
	is.Equal(assert.AnError, child2a.serviceHealthCheck(ctx, "child2a-b"))
	is.ErrorContains(child2a.serviceHealthCheck(ctx, "child2b-a"), "could not find service")

	// from child2b POV
	is.ErrorContains(child2b.serviceHealthCheck(ctx, "root-a"), "could not find service")
	is.ErrorContains(child2b.serviceHealthCheck(ctx, "child1-a"), "could not find service")
	is.ErrorContains(child2b.serviceHealthCheck(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(child2b.serviceHealthCheck(ctx, "child2a-b"), "could not find service")
	is.Equal(assert.AnError, child2b.serviceHealthCheck(ctx, "child2b-a"))
}

func TestScope_clone(t *testing.T) {
	is := assert.New(t)

	rootScope := New()
	child := rootScope.Scope("child1")

	svc1 := newServiceEager("root-a", "root-a")
	svc2 := newServiceEager("child-a", "child-a")
	svc3 := newServiceEager("child-b", "child-b")

	rootScope.serviceSet("root-a", svc1)
	child.serviceSet("child-a", svc2)
	child.serviceSet("child-b", svc3)

	cloneRoot := rootScope.self.clone(rootScope, nil)
	is.Len(cloneRoot.childScopes, 1)
	is.Len(cloneRoot.services, 1)
	is.Len(cloneRoot.childScopes["child1"].services, 2)

	// invoke some services from initial scope -> must not be invoked in clone
	_, _ = invoke[string](child, "child-a")
	_, _ = invoke[string](child, "child-b")

	// invoked list is not cloned
	is.NotEqual(cloneRoot.childScopes["child1"].orderedInvocation, rootScope.self.childScopes["child1"].orderedInvocation)
	is.Len(rootScope.self.childScopes["child1"].orderedInvocation, 2)
	is.Len(cloneRoot.childScopes["child1"].orderedInvocation, 0)

	// @TODO: missing tests
}

func TestScope_serviceHealthCheck(t *testing.T) {
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

	// invokeImplem[int](rootScope, "root-a")	// root-a is not invoked
	_, _ = invoke[int](child1, "child1-a")
	_, _ = invoke[int](child2a, "child2a-a")
	_, _ = invoke[int](child2a, "child2a-b")
	_, _ = invoke[int](child2b, "child2b-a")
	_, _ = invoke[int](child3, "child3-a")

	is.ElementsMatch([]EdgeService{newEdgeService(child3.id, child3.name, "child3-a"), newEdgeService(child2a.id, child2a.name, "child2a-a"), newEdgeService(child2a.id, child2a.name, "child2a-b"), newEdgeService(child1.id, child1.name, "child1-a")}, child3.ListInvokedServices())
	is.Nil(child1.Shutdown())
	is.ElementsMatch([]EdgeService{}, child3.ListInvokedServices())
}

func TestScope_serviceGet(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")

	rootScope.serviceSet("root-a", newServiceLazy("root-a", func(i Injector) (int, error) { return 0, nil }))
	child1.serviceSet("child1-a", newServiceLazy("child1-a", func(i Injector) (int, error) { return 1, nil }))
	child2a.serviceSet("child2a-a", newServiceLazy("child2a-a", func(i Injector) (int, error) { return 2, nil }))
	child2a.serviceSet("child2a-b", newServiceLazy("child2a-b", func(i Injector) (int, error) { return 3, nil }))
	child2b.serviceSet("child2b-a", newServiceLazy("child2b-a", func(i Injector) (int, error) { return 4, nil }))

	cb := func(a any, b bool) bool { return b }

	// from rootScope POV
	is.True(cb(rootScope.serviceGet("root-a")))
	is.False(cb(rootScope.serviceGet("child1-a")))
	is.False(cb(rootScope.serviceGet("child2a-a")))
	is.False(cb(rootScope.serviceGet("child2a-b")))
	is.False(cb(rootScope.serviceGet("child2b-a")))

	// from child1 POV
	is.False(cb(child1.serviceGet("root-a")))
	is.True(cb(child1.serviceGet("child1-a")))
	is.False(cb(child1.serviceGet("child2a-a")))
	is.False(cb(child1.serviceGet("child2a-b")))
	is.False(cb(child1.serviceGet("child2b-a")))

	// from child2a POV
	is.False(cb(child2a.serviceGet("root-a")))
	is.False(cb(child2a.serviceGet("child1-a")))
	is.True(cb(child2a.serviceGet("child2a-a")))
	is.True(cb(child2a.serviceGet("child2a-b")))
	is.False(cb(child2a.serviceGet("child2b-a")))

	// from child2b POV
	is.False(cb(child2b.serviceGet("root-a")))
	is.False(cb(child2b.serviceGet("child1-a")))
	is.False(cb(child2b.serviceGet("child2a-a")))
	is.False(cb(child2b.serviceGet("child2a-b")))
	is.True(cb(child2b.serviceGet("child2b-a")))
}

func TestScope_serviceGetRec(t *testing.T) {
	is := assert.New(t)

	rootScope := New()

	// create children
	child1 := rootScope.Scope("child1")
	child2a := child1.Scope("child2a")
	child2b := child1.Scope("child2b")

	rootScope.serviceSet("root-a", newServiceLazy("root-a", func(i Injector) (int, error) { return 0, nil }))
	child1.serviceSet("child1-a", newServiceLazy("child1-a", func(i Injector) (int, error) { return 1, nil }))
	child2a.serviceSet("child2a-a", newServiceLazy("child2a-a", func(i Injector) (int, error) { return 2, nil }))
	child2a.serviceSet("child2a-b", newServiceLazy("child2a-b", func(i Injector) (int, error) { return 3, nil }))
	child2b.serviceSet("child2b-a", newServiceLazy("child2b-a", func(i Injector) (int, error) { return 4, nil }))

	cb := func(a any, b *Scope, c bool) bool { return c }

	// from rootScope POV
	is.True(cb(rootScope.serviceGetRec("root-a")))
	is.False(cb(rootScope.serviceGetRec("child1-a")))
	is.False(cb(rootScope.serviceGetRec("child2a-a")))
	is.False(cb(rootScope.serviceGetRec("child2a-b")))
	is.False(cb(rootScope.serviceGetRec("child2b-a")))

	// from child1 POV
	is.True(cb(child1.serviceGetRec("root-a")))
	is.True(cb(child1.serviceGetRec("child1-a")))
	is.False(cb(child1.serviceGetRec("child2a-a")))
	is.False(cb(child1.serviceGetRec("child2a-b")))
	is.False(cb(child1.serviceGetRec("child2b-a")))

	// from child2a POV
	is.True(cb(child2a.serviceGetRec("root-a")))
	is.True(cb(child2a.serviceGetRec("child1-a")))
	is.True(cb(child2a.serviceGetRec("child2a-a")))
	is.True(cb(child2a.serviceGetRec("child2a-b")))
	is.False(cb(child2a.serviceGetRec("child2b-a")))

	// from child2b POV
	is.True(cb(child2b.serviceGetRec("root-a")))
	is.True(cb(child2b.serviceGetRec("child1-a")))
	is.False(cb(child2b.serviceGetRec("child2a-a")))
	is.False(cb(child2b.serviceGetRec("child2a-b")))
	is.True(cb(child2b.serviceGetRec("child2b-a")))

	// @TODO: check scope returned by serviceGetRec
}

func TestScope_serviceSet(t *testing.T) {
	is := assert.New(t)

	rootScope := New()
	childa := rootScope.Scope("child1")

	svc1 := newServiceEager("root-a", "root-a")
	svc2 := newServiceEager("child-a", "child-a")

	rootScope.serviceSet("root-a", svc1)
	childa.serviceSet("child-a", svc2)

	// from rootScope POV
	is.Equal(svc1, rootScope.self.services["root-a"])
	is.Nil(rootScope.self.services["child-a"])

	// from childa POV
	is.Nil(childa.services["root-a"])
	is.Equal(svc2, childa.services["child-a"])
}

func TestScope_serviceForEach(t *testing.T) {
	is := assert.New(t)

	rootScope := New()
	child := rootScope.Scope("child1")

	svc1 := newServiceEager("root-a", "root-a")
	svc2 := newServiceEager("child-a", "child-a")
	svc3 := newServiceEager("child-b", "child-b")

	rootScope.serviceSet("root-a", svc1)
	child.serviceSet("child-a", svc2)
	child.serviceSet("child-b", svc3)

	// from rootScope POV
	rootScope.serviceForEach(func(name string, service any) {
		switch name {
		case "root-a":
			is.Equal(svc1, service)
		default:
			is.Fail("should not be called")
		}
	})

	// from child POV
	child.serviceForEach(func(name string, service any) {
		switch name {
		case "child-a":
			is.Equal(svc2, service)
		case "child-b":
			is.Equal(svc3, service)
		default:
			is.Fail("should not be called")
		}
	})
}

func TestScope_serviceShutdown(t *testing.T) {
	// @TODO
}

func TestScope_onServiceRegistration(t *testing.T) {
	// @TODO
}

func TestScope_onServiceInvoke(t *testing.T) {
	// @TODO
}

func TestScope_onServiceShutdown(t *testing.T) {
	// @TODO
}

func TestScope_logs(t *testing.T) {
	// @TODO
}
