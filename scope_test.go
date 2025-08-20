package do

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"errors"

	"github.com/stretchr/testify/assert"
)

func TestNewScope(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	scope := newScope("foobar", nil, nil)

	is.NotEmpty(scope.ID())
	is.Len(scope.ID(), 36)
}

func TestScope_Name(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	scope := newScope("foobar", nil, nil)

	is.NotEmpty(scope.Name())
	is.Equal(scope.Name(), "foobar")
}

func TestScope_Scope(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
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

func TestScope_Scope_race(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	injector := New()

	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(j int) {
			injector.Scope(fmt.Sprintf("test-%d", j))
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestScope_RootScope(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	_, _ = invokeByName[int](child1, "child1-a")
	_, _ = invokeByName[int](child2a, "child2a-a")
	_, _ = invokeByName[int](child2a, "child2a-b")
	_, _ = invokeByName[int](child2b, "child2b-a")
	_, _ = invokeByName[int](child3, "child3-a")

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

	_, _ = invokeByName[*lazyTestHeathcheckerKO](rootScope, "root-a")
	_, _ = invokeByName[*lazyTestHeathcheckerOK](child1, "child1-a")
	_, _ = invokeByName[*lazyTestHeathcheckerOK](child2a, "child2a-a")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child2a, "child2a-b")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child2b, "child2b-a")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child3, "child3-a")

	is.EqualValues(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": assert.AnError, "root-a": assert.AnError}, child2a.HealthCheck())
	is.EqualValues(map[string]error{"root-a": assert.AnError}, rootScope.HealthCheckWithContext(context.Background()))
}

func TestScope_Shutdown(t *testing.T) {
	is := assert.New(t)

	i := New()

	ProvideNamedValue(i, "lazy-ok", &lazyTestShutdownerOK{})
	ProvideNamedValue(i, "lazy-ko", &lazyTestShutdownerKO{})
	_, _ = InvokeNamed[*lazyTestShutdownerOK](i, "lazy-ok")
	_, _ = InvokeNamed[*lazyTestShutdownerKO](i, "lazy-ko")

	is.EqualValues(&ShutdownErrors{EdgeService{ScopeID: i.self.id, ScopeName: i.self.name, Service: "lazy-ko"}: assert.AnError}, i.Shutdown())
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

	provider1 := func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	}
	provider2 := func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	}

	rootScope.serviceSet("root-a", newServiceLazy("root-a", provider2))
	child1.serviceSet("child1-a", newServiceLazy("child1-a", provider1))
	child2a.serviceSet("child2a-a", newServiceLazy("child2a-a", provider1))
	child2a.serviceSet("child2a-b", newServiceLazy("child2a-b", provider2))
	child2b.serviceSet("child2b-a", newServiceLazy("child2b-a", provider2))

	_, _ = invokeByName[*lazyTestShutdownerKO](rootScope, "root-a")
	_, _ = invokeByName[*lazyTestShutdownerOK](child1, "child1-a")
	_, _ = invokeByName[*lazyTestShutdownerOK](child2a, "child2a-a")
	_, _ = invokeByName[*lazyTestShutdownerKO](child2a, "child2a-b")
	_, _ = invokeByName[*lazyTestShutdownerKO](child2b, "child2b-a")

	// // from rootScope POV
	is.Equal(assert.AnError, rootScope.serviceShutdown(ctx, "root-a"))
	is.ErrorContains(rootScope.serviceShutdown(ctx, "child1-a"), "could not find service")
	is.ErrorContains(rootScope.serviceShutdown(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(rootScope.serviceShutdown(ctx, "child2a-b"), "could not find service")
	is.ErrorContains(rootScope.serviceShutdown(ctx, "child2b-a"), "could not find service")

	// from child1 POV
	is.ErrorContains(child1.serviceShutdown(ctx, "root-a"), "could not find service")
	is.Equal(nil, child1.serviceShutdown(ctx, "child1-a"))
	is.ErrorContains(child1.serviceShutdown(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(child1.serviceShutdown(ctx, "child2a-b"), "could not find service")
	is.ErrorContains(child1.serviceShutdown(ctx, "child2b-a"), "could not find service")

	// from child2a POV
	is.ErrorContains(child2a.serviceShutdown(ctx, "root-a"), "could not find service")
	is.ErrorContains(child2a.serviceShutdown(ctx, "child1-a"), "could not find service")
	is.Equal(nil, child2a.serviceShutdown(ctx, "child2a-a"))
	is.Equal(assert.AnError, child2a.serviceShutdown(ctx, "child2a-b"))
	is.ErrorContains(child2a.serviceShutdown(ctx, "child2b-a"), "could not find service")

	// from child2b POV
	is.ErrorContains(child2b.serviceShutdown(ctx, "root-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child1-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child2a-b"), "could not find service")
	is.Equal(assert.AnError, child2b.serviceShutdown(ctx, "child2b-a"))

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
	_, _ = invokeByName[string](child, "child-a")
	_, _ = invokeByName[string](child, "child-b")

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
	_, _ = invokeByName[int](child1, "child1-a")
	_, _ = invokeByName[int](child2a, "child2a-a")
	_, _ = invokeByName[int](child2a, "child2a-b")
	_, _ = invokeByName[int](child2b, "child2b-a")
	_, _ = invokeByName[int](child3, "child3-a")

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
	counter := 0
	rootScope.serviceForEach(func(name string, scope *Scope, service any) bool {
		counter++
		switch name {
		case "root-a":
			is.Equal(svc1, service)
		default:
			is.Fail("should not be called")
		}
		is.Equal(rootScope.ID(), scope.ID())
		return true
	})
	is.Equal(1, counter)

	// from child POV
	counter = 0
	child.serviceForEach(func(name string, scope *Scope, service any) bool {
		counter++
		switch name {
		case "child-a":
			is.Equal(svc2, service)
		case "child-b":
			is.Equal(svc3, service)
		default:
			is.Fail("should not be called")
		}
		is.Equal(child.ID(), scope.ID())
		return true
	})
	is.Equal(2, counter)
}

func TestScope_serviceForEachRec(t *testing.T) {
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
	counter := 0
	rootScope.serviceForEachRec(func(name string, scope *Scope, service any) bool {
		counter++
		switch name {
		case "root-a":
			is.Equal(svc1, service)
		default:
			is.Fail("should not be called")
		}
		is.Equal(rootScope.ID(), scope.ID())
		return true
	})
	is.Equal(1, counter)

	// from child POV
	counter = 0
	child.serviceForEachRec(func(name string, scope *Scope, service any) bool {
		counter++
		switch name {
		case "root-a":
			is.Equal(svc1, service)
			is.Equal(rootScope.ID(), scope.ID())
		case "child-a":
			is.Equal(svc2, service)
			is.Equal(child.ID(), scope.ID())
		case "child-b":
			is.Equal(svc3, service)
			is.Equal(child.ID(), scope.ID())
		default:
			is.Fail("should not be called")
		}
		return true
	})
	is.Equal(3, counter)
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

type shutdownRecorder struct{ called *int32 }

func (s *shutdownRecorder) Shutdown() error {
	atomic.AddInt32(s.called, 1)
	return nil
}

// Test the shutdown of services that have no dependencies and no dependents.
func TestScope_Shutdown_NoDependencies(t *testing.T) {
	is := assert.New(t)

	i := New()

	var a, b, c int32
	ProvideNamedValue(i, "svc-a", &shutdownRecorder{called: &a})
	ProvideNamedValue(i, "svc-b", &shutdownRecorder{called: &b})
	ProvideNamedValue(i, "svc-c", &shutdownRecorder{called: &c})

	// No dependencies between services; shutdown should call all three
	is.Nil(i.Shutdown())

	is.EqualValues(int32(1), atomic.LoadInt32(&a))
	is.EqualValues(int32(1), atomic.LoadInt32(&b))
	is.EqualValues(int32(1), atomic.LoadInt32(&c))

	// Services should be removed after shutdown
	_, errA := InvokeNamed[*shutdownRecorder](i, "svc-a")
	_, errB := InvokeNamed[*shutdownRecorder](i, "svc-b")
	_, errC := InvokeNamed[*shutdownRecorder](i, "svc-c")
	is.NotNil(errA)
	is.NotNil(errB)
	is.NotNil(errC)
}

// Test services for context expiration testing
var _ ShutdownerWithContextAndError = (*scopeTestSlowShutdowner)(nil)

type scopeTestSlowShutdowner struct {
	shutdownDelay time.Duration
	shutdownCount int32
}

func newScopeTestSlowShutdowner(delay time.Duration) *scopeTestSlowShutdowner {
	return &scopeTestSlowShutdowner{shutdownDelay: delay}
}

func (s *scopeTestSlowShutdowner) Shutdown(ctx context.Context) error {
	atomic.AddInt32(&s.shutdownCount, 1)

	// Use a timer with context to avoid goroutine leaks
	timer := time.NewTimer(s.shutdownDelay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *scopeTestSlowShutdowner) getShutdownCount() int {
	return int(atomic.LoadInt32(&s.shutdownCount))
}

var _ HealthcheckerWithContext = (*scopeTestSlowHealthchecker)(nil)

type scopeTestSlowHealthchecker struct {
	healthcheckDelay time.Duration
	healthcheckCount int32
}

func newScopeTestSlowHealthchecker(delay time.Duration) *scopeTestSlowHealthchecker {
	return &scopeTestSlowHealthchecker{healthcheckDelay: delay}
}

func (s *scopeTestSlowHealthchecker) HealthCheck(ctx context.Context) error {
	atomic.AddInt32(&s.healthcheckCount, 1)

	// Use a timer with context to avoid goroutine leaks
	timer := time.NewTimer(s.healthcheckDelay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *scopeTestSlowHealthchecker) getHealthcheckCount() int {
	return int(atomic.LoadInt32(&s.healthcheckCount))
}

var _ ShutdownerWithContextAndError = (*scopeTestBlockingShutdowner)(nil)

type scopeTestBlockingShutdowner struct {
	blocked       chan struct{}
	shutdownCount int32
}

func newScopeTestBlockingShutdowner() *scopeTestBlockingShutdowner {
	return &scopeTestBlockingShutdowner{
		blocked: make(chan struct{}),
	}
}

func (s *scopeTestBlockingShutdowner) Shutdown(ctx context.Context) error {
	atomic.AddInt32(&s.shutdownCount, 1)

	select {
	case <-s.blocked:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *scopeTestBlockingShutdowner) getShutdownCount() int {
	return int(atomic.LoadInt32(&s.shutdownCount))
}

var _ HealthcheckerWithContext = (*scopeTestBlockingHealthchecker)(nil)

type scopeTestBlockingHealthchecker struct {
	blocked          chan struct{}
	healthcheckCount int32
}

func newScopeTestBlockingHealthchecker() *scopeTestBlockingHealthchecker {
	return &scopeTestBlockingHealthchecker{
		blocked: make(chan struct{}),
	}
}

func (s *scopeTestBlockingHealthchecker) HealthCheck(ctx context.Context) error {
	atomic.AddInt32(&s.healthcheckCount, 1)

	select {
	case <-s.blocked:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *scopeTestBlockingHealthchecker) getHealthcheckCount() int {
	return int(atomic.LoadInt32(&s.healthcheckCount))
}

// Test shutdown context expiration
func TestScope_ShutdownWithContextExpiration_Timeout(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)
	injector := New()

	// Create a service that takes longer than the context timeout
	slowService := newScopeTestSlowShutdowner(200 * time.Millisecond)
	ProvideNamedValue(injector, "slow-service", slowService)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Shutdown should timeout
	start := time.Now()
	errors := injector.ShutdownWithContext(ctx)
	duration := time.Since(start)

	// Should complete quickly due to timeout
	is.Less(duration, 70*time.Millisecond)

	// Should have shutdown errors due to timeout
	is.NotNil(errors)
	is.Equal(1, errors.Len())

	// Check that the service was attempted to be shut down
	is.Equal(1, slowService.getShutdownCount())
}

func TestScope_ShutdownWithContextExpiration_Cancellation(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)
	injector := New()

	// Create a blocking service
	blockingService := newScopeTestBlockingShutdowner()
	ProvideNamedValue(injector, "blocking-service", blockingService)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start shutdown in goroutine
	var shutdownErrors *ShutdownErrors
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErrors = injector.ShutdownWithContext(ctx)
	}()

	// Cancel context after short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for shutdown to complete
	wg.Wait()

	// Should have shutdown errors due to cancellation
	is.NotNil(shutdownErrors)
	is.Equal(1, shutdownErrors.Len())

	// Check that the service was attempted to be shut down
	is.Equal(1, blockingService.getShutdownCount())
}

func TestScope_ShutdownWithContextExpiration_MultipleServices(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)
	injector := New()

	// Create multiple services with different delays
	fastService := newScopeTestSlowShutdowner(10 * time.Millisecond)
	slowService := newScopeTestSlowShutdowner(200 * time.Millisecond)

	ProvideNamedValue(injector, "fast-service", fastService)
	ProvideNamedValue(injector, "slow-service", slowService)

	// Create context with timeout between fast and slow service delays
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Shutdown should timeout
	errors := injector.ShutdownWithContext(ctx)

	// Should have shutdown errors due to timeout
	is.NotNil(errors)
	is.Equal(1, errors.Len())

	// Both services should have been attempted
	is.Equal(1, fastService.getShutdownCount())
	is.Equal(1, slowService.getShutdownCount())
}

func TestScope_ShutdownWithContextExpiration_ChildScopes(t *testing.T) {
	is := assert.New(t)
	injector := New()
	childScope := injector.Scope("child")

	// Create services in both scopes with delays longer than timeout
	parentService := newScopeTestSlowShutdowner(200 * time.Millisecond)
	childService := newScopeTestSlowShutdowner(200 * time.Millisecond)

	ProvideNamedValue(injector, "parent-service", parentService)
	ProvideNamedValue(childScope, "child-service", childService)

	// Invoke the services to ensure they are instantiated
	_, err1 := InvokeNamed[*scopeTestSlowShutdowner](injector, "parent-service")
	_, err2 := InvokeNamed[*scopeTestSlowShutdowner](childScope, "child-service")
	is.Nil(err1)
	is.Nil(err2)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Shutdown should timeout - call shutdown on root scope should shut down all child scopes
	errors := injector.ShutdownWithContext(ctx)
	is.NotNil(errors)
	is.Equal(2, errors.Len())

	// Child service should have been attempted (child scopes are shut down first)
	is.Equal(1, childService.getShutdownCount())

	// Parent service might not be attempted if context times out during child shutdown
	// This is expected behavior for this test - we only verify child scope shutdown works
	is.Equal(0, parentService.getShutdownCount())
}

// Test healthcheck context expiration
func TestScope_HealthCheckWithContextExpiration_Timeout(t *testing.T) {
	is := assert.New(t)
	injector := New()

	// Create a service that takes longer than the context timeout
	slowService := newScopeTestSlowHealthchecker(200 * time.Millisecond)
	ProvideNamedValue(injector, "slow-healthcheck", slowService)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Healthcheck should timeout
	start := time.Now()
	results := injector.HealthCheckWithContext(ctx)
	duration := time.Since(start)

	// Should complete quickly due to timeout
	is.Less(duration, 70*time.Millisecond)

	// Should have healthcheck errors due to timeout
	is.NotEmpty(results)
	is.Len(results, 1)
	for _, err := range results {
		is.NotNil(err)
		// Check if it's a timeout error (could be wrapped)
		is.True(errors.Is(err, context.DeadlineExceeded) || err.Error() == "DI: health check timeout: context deadline exceeded")
	}

	// Check that the service was attempted to be health checked
	is.Equal(1, slowService.getHealthcheckCount())
}

func TestScope_HealthCheckWithContextExpiration_Cancellation(t *testing.T) {
	is := assert.New(t)
	injector := New()

	// Create a blocking service
	blockingService := newScopeTestBlockingHealthchecker()
	ProvideNamedValue(injector, "blocking-healthcheck", blockingService)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start healthcheck in goroutine
	var results map[string]error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		results = injector.HealthCheckWithContext(ctx)
	}()

	// Cancel context after short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for healthcheck to complete
	wg.Wait()

	// Should have healthcheck errors due to cancellation
	is.NotEmpty(results)
	is.Len(results, 1)
	for _, err := range results {
		is.NotNil(err)
		// Check if it's a cancellation error (could be wrapped)
		is.True(errors.Is(err, context.Canceled) || err.Error() == "DI: health check timeout: context canceled")
	}

	// Check that the service was attempted to be health checked
	is.Equal(1, blockingService.getHealthcheckCount())
}

func TestScope_HealthCheckWithContextExpiration_MultipleServices(t *testing.T) {
	is := assert.New(t)
	injector := New()

	// Create multiple services with different delays
	fastService := newScopeTestSlowHealthchecker(10 * time.Millisecond)
	slowService := newScopeTestSlowHealthchecker(200 * time.Millisecond)

	ProvideNamedValue(injector, "fast-healthcheck", fastService)
	ProvideNamedValue(injector, "slow-healthcheck", slowService)

	// Invoke the services to ensure they are instantiated
	_, err1 := InvokeNamed[*scopeTestSlowHealthchecker](injector, "fast-healthcheck")
	_, err2 := InvokeNamed[*scopeTestSlowHealthchecker](injector, "slow-healthcheck")
	is.Nil(err1)
	is.Nil(err2)

	// Create context with timeout between fast and slow service delays
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Healthcheck should timeout
	results := injector.HealthCheckWithContext(ctx)

	// Should have healthcheck results
	is.NotEmpty(results)
	is.Len(results, 2)

	// Check that both services were attempted
	is.Equal(1, fastService.getHealthcheckCount())
	is.Equal(1, slowService.getHealthcheckCount())

	// At least one service should have a timeout error
	hasTimeoutError := 0
	for _, err := range results {
		if err != nil && (errors.Is(err, context.DeadlineExceeded) || err.Error() == "DI: health check timeout: context deadline exceeded") {
			hasTimeoutError++
		}
	}
	is.Equal(1, hasTimeoutError, "Expected at least one timeout error")
}

func TestScope_HealthCheckWithContextExpiration_ChildScopes(t *testing.T) {
	is := assert.New(t)
	injector := New()
	childScope := injector.Scope("child")

	// Create services in both scopes
	parentService := newScopeTestSlowHealthchecker(200 * time.Millisecond)
	childService := newScopeTestSlowHealthchecker(200 * time.Millisecond)

	ProvideNamedValue(injector, "parent-healthcheck", parentService)
	ProvideNamedValue(childScope, "child-healthcheck", childService)

	// Invoke the services to ensure they are instantiated
	_, err1 := InvokeNamed[*scopeTestSlowHealthchecker](injector, "parent-healthcheck")
	_, err2 := InvokeNamed[*scopeTestSlowHealthchecker](childScope, "child-healthcheck")
	is.Nil(err1)
	is.Nil(err2)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Healthcheck should timeout - call healthcheck on child scope first, then parent scope
	results1 := childScope.HealthCheckWithContext(ctx)
	_ = injector.HealthCheckWithContext(ctx)

	// Should have healthcheck results
	is.NotEmpty(results1)

	// Check that both services were attempted
	is.Equal(1, parentService.getHealthcheckCount())
	is.Equal(1, childService.getHealthcheckCount())

	// At least one service should have a timeout error
	hasTimeoutError := 0
	for _, err := range results1 {
		if err != nil && (errors.Is(err, context.DeadlineExceeded) || err.Error() == "DI: health check timeout: context deadline exceeded") {
			hasTimeoutError++
		}
	}
	is.Equal(2, hasTimeoutError, "Expected at least one timeout error")
}

func TestScope_HealthCheckWithContextExpiration_GlobalTimeoutOption(t *testing.T) {
	is := assert.New(t)
	// Create injector with global healthcheck timeout
	injector := NewWithOpts(&InjectorOpts{
		HealthCheckGlobalTimeout: 50 * time.Millisecond,
	})

	// Create a service that takes longer than the global timeout
	slowService := newScopeTestSlowHealthchecker(200 * time.Millisecond)
	ProvideNamedValue(injector, "global-timeout-service", slowService)

	// Invoke the service to ensure it is instantiated
	_, err := InvokeNamed[*scopeTestSlowHealthchecker](injector, "global-timeout-service")
	is.Nil(err)

	// Healthcheck should timeout due to global timeout
	start := time.Now()
	results := injector.HealthCheckWithContext(context.Background())
	duration := time.Since(start)

	// Should complete quickly due to global timeout
	is.Less(duration, 100*time.Millisecond)

	// Should have healthcheck errors due to timeout
	is.NotEmpty(results)
	for _, err := range results {
		is.NotNil(err)
		// Check if it's a timeout error (could be wrapped)
		is.True(errors.Is(err, context.DeadlineExceeded) || err.Error() == "DI: health check timeout: context deadline exceeded")
	}

	// Check that the service was attempted to be health checked
	is.Equal(1, slowService.getHealthcheckCount())
}

// Test mixed scenarios
func TestScope_ContextExpiration_ShutdownAndHealthcheckSameTimeout(t *testing.T) {
	is := assert.New(t)
	injector := New()

	// Create services that implement both interfaces
	shutdownService := newScopeTestSlowShutdowner(200 * time.Millisecond)
	healthcheckService := newScopeTestSlowHealthchecker(200 * time.Millisecond)

	ProvideNamedValue(injector, "shutdown-service", shutdownService)
	ProvideNamedValue(injector, "healthcheck-service", healthcheckService)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Test shutdown timeout
	shutdownErrors := injector.ShutdownWithContext(ctx)
	is.NotNil(shutdownErrors)
	is.Equal(1, shutdownErrors.Len())

	// Recreate injector for healthcheck test
	injector2 := New()
	ProvideNamedValue(injector2, "healthcheck-service", healthcheckService)

	// Test healthcheck timeout
	healthcheckResults := injector2.HealthCheckWithContext(ctx)
	is.NotEmpty(healthcheckResults)
	is.Len(healthcheckResults, 1)
	for _, err := range healthcheckResults {
		is.NotNil(err)
		// Check if it's a timeout error (could be wrapped)
		is.True(errors.Is(err, context.DeadlineExceeded) || err.Error() == "DI: health check timeout: context deadline exceeded")
	}
}

func TestScope_ContextExpiration_ParallelOperations(t *testing.T) {
	is := assert.New(t)
	injector := New()

	// Create multiple blocking services with unique names
	services := make([]*scopeTestBlockingShutdowner, 5)
	for i := range services {
		services[i] = newScopeTestBlockingShutdowner()
		ProvideNamedValue(injector, fmt.Sprintf("blocking-service-%d", i), services[i])
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start shutdown in goroutine
	var shutdownErrors *ShutdownErrors
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdownErrors = injector.ShutdownWithContext(ctx)
	}()

	// Cancel context after short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for shutdown to complete
	wg.Wait()

	// Should have shutdown errors due to cancellation
	is.NotNil(shutdownErrors)
	is.Equal(5, shutdownErrors.Len())

	// All services should have been attempted
	for _, service := range services {
		is.Equal(1, service.getShutdownCount())
	}
}
