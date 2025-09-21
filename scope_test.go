package do

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewScope(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	scope := newScope("foobar", nil, nil)

	is.NotEmpty(scope.ID())
	is.Len(scope.ID(), 36)
}

func TestScope_Name(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	scope := newScope("foobar", nil, nil)

	is.NotEmpty(scope.Name())
	is.Equal("foobar", scope.Name())
}

func TestScope_Scope(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	// override in child3
	child3.serviceSet("root-a", newServiceEager("root-a", 0))

	is.ElementsMatch([]ServiceDescription{newServiceDescription(rootScope.ID(), rootScope.Name(), "root-a")}, rootScope.ListProvidedServices())
	is.ElementsMatch([]ServiceDescription{newServiceDescription(child1.ID(), child1.Name(), "child1-a"), newServiceDescription(rootScope.ID(), rootScope.Name(), "root-a")}, child1.ListProvidedServices())
	is.ElementsMatch([]ServiceDescription{newServiceDescription(child2a.ID(), child2a.Name(), "child2a-a"), newServiceDescription(child2a.ID(), child2a.Name(), "child2a-b"), newServiceDescription(child1.ID(), child1.Name(), "child1-a"), newServiceDescription(rootScope.ID(), rootScope.Name(), "root-a")}, child2a.ListProvidedServices())
	is.ElementsMatch([]ServiceDescription{newServiceDescription(child2b.ID(), child2b.Name(), "child2b-a"), newServiceDescription(child1.ID(), child1.Name(), "child1-a"), newServiceDescription(rootScope.ID(), rootScope.Name(), "root-a")}, child2b.ListProvidedServices())
	is.ElementsMatch(
		[]ServiceDescription{
			newServiceDescription(child3.ID(), child3.Name(), "child3-a"),
			newServiceDescription(child3.ID(), child3.Name(), "root-a"),
			newServiceDescription(child2a.ID(), child2a.Name(), "child2a-a"),
			newServiceDescription(child2a.ID(), child2a.Name(), "child2a-b"),
			newServiceDescription(child1.ID(), child1.Name(), "child1-a"),
			newServiceDescription(rootScope.ID(), rootScope.Name(), "root-a"),
		},
		child3.ListProvidedServices(),
	)
}

func TestScope_ListInvokedServices(t *testing.T) {
	t.Parallel()
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

	is.ElementsMatch([]ServiceDescription{}, rootScope.ListInvokedServices())
	is.ElementsMatch([]ServiceDescription{newServiceDescription(child1.ID(), child1.Name(), "child1-a")}, child1.ListInvokedServices())
	is.ElementsMatch([]ServiceDescription{newServiceDescription(child2a.ID(), child2a.Name(), "child2a-a"), newServiceDescription(child2a.ID(), child2a.Name(), "child2a-b"), newServiceDescription(child1.ID(), child1.Name(), "child1-a")}, child2a.ListInvokedServices())
	is.ElementsMatch([]ServiceDescription{newServiceDescription(child2b.ID(), child2b.Name(), "child2b-a"), newServiceDescription(child1.ID(), child1.Name(), "child1-a")}, child2b.ListInvokedServices())
	is.ElementsMatch([]ServiceDescription{newServiceDescription(child3.ID(), child3.Name(), "child3-a"), newServiceDescription(child2a.ID(), child2a.Name(), "child2a-a"), newServiceDescription(child2a.ID(), child2a.Name(), "child2a-b"), newServiceDescription(child1.ID(), child1.Name(), "child1-a")}, child3.ListInvokedServices())

	is.Equal(map[string]int{}, rootScope.self.orderedInvocation)
	is.Equal(map[string]int{"child1-a": 0}, child1.orderedInvocation)
	is.Equal(map[string]int{"child2a-a": 0, "child2a-b": 1}, child2a.orderedInvocation)
	is.Equal(map[string]int{"child2b-a": 0}, child2b.orderedInvocation)
	is.Equal(map[string]int{"child3-a": 0}, child3.orderedInvocation)
}

func TestScope_HealthCheck(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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

	// Test healthcheck before services are invoked
	is.Equal(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": nil, "root-a": nil}, child2a.HealthCheck())

	// Invoke services to make them healthcheckable
	_, _ = invokeByName[*lazyTestHeathcheckerKO](rootScope, "root-a")
	_, _ = invokeByName[*lazyTestHeathcheckerOK](child1, "child1-a")
	_, _ = invokeByName[*lazyTestHeathcheckerOK](child2a, "child2a-a")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child2a, "child2a-b")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child2b, "child2b-a")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child3, "child3-a")

	// Test healthcheck after services are invoked
	is.Equal(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": assert.AnError, "root-a": assert.AnError}, child2a.HealthCheck())
	is.Equal(map[string]error{"root-a": assert.AnError}, rootScope.HealthCheck())

	// Test healthcheck from different scopes
	is.Equal(map[string]error{"child2b-a": assert.AnError, "child1-a": nil, "root-a": assert.AnError}, child2b.HealthCheck())
	is.Equal(map[string]error{"child3-a": assert.AnError, "child2a-a": nil, "child2a-b": assert.AnError, "child1-a": nil, "root-a": assert.AnError}, child3.HealthCheck())

	// Test healthcheck with no healthcheckable services
	emptyScope := rootScope.Scope("empty")
	is.Equal(map[string]error{"root-a": assert.AnError}, emptyScope.HealthCheck()) // Includes ancestor services

	// Test healthcheck with services that don't implement Healthchecker
	nonHealthcheckableScope := rootScope.Scope("non-healthcheckable")
	nonHealthcheckableScope.serviceSet("non-healthcheckable-a", newServiceLazy("non-healthcheckable-a", func(i Injector) (int, error) { return 42, nil }))
	_, _ = invokeByName[int](nonHealthcheckableScope, "non-healthcheckable-a")
	is.Equal(map[string]error{"non-healthcheckable-a": nil, "root-a": assert.AnError}, nonHealthcheckableScope.HealthCheck()) // Includes ancestor services
}

func TestScope_HealthCheckWithContext(t *testing.T) {
	t.Parallel()
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

	is.Equal(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": nil, "root-a": nil}, child2a.HealthCheck())

	_, _ = invokeByName[*lazyTestHeathcheckerKO](rootScope, "root-a")
	_, _ = invokeByName[*lazyTestHeathcheckerOK](child1, "child1-a")
	_, _ = invokeByName[*lazyTestHeathcheckerOK](child2a, "child2a-a")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child2a, "child2a-b")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child2b, "child2b-a")
	_, _ = invokeByName[*lazyTestHeathcheckerKO](child3, "child3-a")

	is.Equal(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": assert.AnError, "root-a": assert.AnError}, child2a.HealthCheck())
	is.Equal(map[string]error{"root-a": assert.AnError}, rootScope.HealthCheckWithContext(context.Background()))

	// Test with different context scenarios
	ctx := context.Background()

	// Test with background context
	results1 := child2a.HealthCheckWithContext(ctx)
	is.Equal(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": assert.AnError, "root-a": assert.AnError}, results1)

	// Test with canceled context - should return timeout errors
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	results2 := child2a.HealthCheckWithContext(canceledCtx)
	// When context is canceled, some services might return timeout errors, others might return original errors
	is.Len(results2, 4) // child1-a, child2a-a, child2a-b, root-a

	// Test with timeout context - should return timeout errors
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	results3 := child2a.HealthCheckWithContext(timeoutCtx)
	// When context times out, some services might return timeout errors, others might return original errors
	is.Len(results3, 4) // child1-a, child2a-a, child2a-b, root-a

	// Test with deadline context - should return timeout errors
	deadlineCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(100*time.Millisecond))
	defer cancel()
	results4 := child2a.HealthCheckWithContext(deadlineCtx)
	// When context deadline is exceeded, some services might return timeout errors, others might return original errors
	is.Len(results4, 4) // child1-a, child2a-a, child2a-b, root-a

	// Test with value context - should return original errors
	valueCtx := context.WithValue(context.Background(), ctxTestKey, "test-value")
	results5 := child2a.HealthCheckWithContext(valueCtx)
	is.Equal(map[string]error{"child1-a": nil, "child2a-a": nil, "child2a-b": assert.AnError, "root-a": assert.AnError}, results5)
}

func TestScope_Shutdown(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()

	ProvideNamedValue(i, "lazy-ok", &lazyTestShutdownerOK{})
	ProvideNamedValue(i, "lazy-ko", &lazyTestShutdownerKO{})
	_, _ = InvokeNamed[*lazyTestShutdownerOK](i, "lazy-ok")
	_, _ = InvokeNamed[*lazyTestShutdownerKO](i, "lazy-ko")

	shutdownReport := i.Shutdown()
	is.Len(shutdownReport.Errors, 1)
	is.Contains(shutdownReport.Errors, ServiceDescription{ScopeID: i.self.id, ScopeName: i.self.name, Service: "lazy-ko"})
}

func TestScope_ShutdownWithContext(t *testing.T) {
	t.Parallel()
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
	is.NoError(child1.serviceShutdown(ctx, "child1-a"))
	is.ErrorContains(child1.serviceShutdown(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(child1.serviceShutdown(ctx, "child2a-b"), "could not find service")
	is.ErrorContains(child1.serviceShutdown(ctx, "child2b-a"), "could not find service")

	// from child2a POV
	is.ErrorContains(child2a.serviceShutdown(ctx, "root-a"), "could not find service")
	is.ErrorContains(child2a.serviceShutdown(ctx, "child1-a"), "could not find service")
	is.NoError(child2a.serviceShutdown(ctx, "child2a-a"))
	is.Equal(assert.AnError, child2a.serviceShutdown(ctx, "child2a-b"))
	is.ErrorContains(child2a.serviceShutdown(ctx, "child2b-a"), "could not find service")

	// from child2b POV
	is.ErrorContains(child2b.serviceShutdown(ctx, "root-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child1-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child2a-b"), "could not find service")
	is.Equal(assert.AnError, child2b.serviceShutdown(ctx, "child2b-a"))

	// Test with different context scenarios
	// Test with background context
	results1 := child2a.ShutdownWithContext(ctx)
	is.NotNil(results1)
	is.Empty(results1.Errors) // child2a-b already shut down

	// Test with canceled context - create new scope
	rootScope2 := New()
	child1_2 := rootScope2.Scope("child1")
	child2a_2 := child1_2.Scope("child2a")
	child2a_2.serviceSet("child2a-a", newServiceLazy("child2a-a", provider1))
	child2a_2.serviceSet("child2a-b", newServiceLazy("child2a-b", provider2))
	_, _ = invokeByName[*lazyTestShutdownerOK](child2a_2, "child2a-a")
	_, _ = invokeByName[*lazyTestShutdownerKO](child2a_2, "child2a-b")

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	results2 := child2a_2.ShutdownWithContext(canceledCtx)
	is.NotNil(results2)
	is.Len(results2.Errors, 2) // both services should fail due to context cancellation

	// Test with timeout context - create new scope
	rootScope3 := New()
	child1_3 := rootScope3.Scope("child1")
	child2a_3 := child1_3.Scope("child2a")
	child2a_3.serviceSet("child2a-a", newServiceLazy("child2a-a", provider1))
	child2a_3.serviceSet("child2a-b", newServiceLazy("child2a-b", provider2))
	_, _ = invokeByName[*lazyTestShutdownerOK](child2a_3, "child2a-a")
	_, _ = invokeByName[*lazyTestShutdownerKO](child2a_3, "child2a-b")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	results3 := child2a_3.ShutdownWithContext(timeoutCtx)
	is.NotNil(results3)
	is.Len(results3.Errors, 1) // only child2a-b should fail

	// Test with deadline context - create new scope
	rootScope4 := New()
	child1_4 := rootScope4.Scope("child1")
	child2a_4 := child1_4.Scope("child2a")
	child2a_4.serviceSet("child2a-a", newServiceLazy("child2a-a", provider1))
	child2a_4.serviceSet("child2a-b", newServiceLazy("child2a-b", provider2))
	_, _ = invokeByName[*lazyTestShutdownerOK](child2a_4, "child2a-a")
	_, _ = invokeByName[*lazyTestShutdownerKO](child2a_4, "child2a-b")

	deadlineCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(100*time.Millisecond))
	defer cancel()
	results4 := child2a_4.ShutdownWithContext(deadlineCtx)
	is.NotNil(results4)
	is.Len(results4.Errors, 1) // only child2a-b should fail

	// Test with value context - create new scope
	rootScope5 := New()
	child1_5 := rootScope5.Scope("child1")
	child2a_5 := child1_5.Scope("child2a")
	child2a_5.serviceSet("child2a-a", newServiceLazy("child2a-a", provider1))
	child2a_5.serviceSet("child2a-b", newServiceLazy("child2a-b", provider2))
	_, _ = invokeByName[*lazyTestShutdownerOK](child2a_5, "child2a-a")
	_, _ = invokeByName[*lazyTestShutdownerKO](child2a_5, "child2a-b")

	valueCtx := context.WithValue(context.Background(), ctxTestKey, "test-value")
	results5 := child2a_5.ShutdownWithContext(valueCtx)
	is.NotNil(results5)
	is.Len(results5.Errors, 1) // child2a-b should fail

	// Test shutdown from root scope - create new scope
	rootScope6 := New()
	rootScope6.serviceSet("root-a", newServiceLazy("root-a", provider2))
	_, _ = invokeByName[*lazyTestShutdownerKO](rootScope6, "root-a")

	rootResults := rootScope6.ShutdownWithContext(ctx)
	is.NotNil(rootResults)
	is.Len(rootResults.Errors, 1) // root-a should fail
}

func TestScope_clone(t *testing.T) {
	t.Parallel()
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
	is.Empty(cloneRoot.childScopes["child1"].orderedInvocation)

	// Test that cloned services are independent
	cloneChild := cloneRoot.childScopes["child1"]

	// Test that services can be invoked in clone
	instance1, err1 := invokeByName[string](cloneChild, "child-a")
	is.NoError(err1)
	is.Equal("child-a", instance1)

	instance2, err2 := invokeByName[string](cloneChild, "child-b")
	is.NoError(err2)
	is.Equal("child-b", instance2)

	// Test that clone has its own invocation tracking
	is.Len(cloneChild.orderedInvocation, 2)
	is.Equal(0, cloneChild.orderedInvocation["child-a"])
	is.Equal(1, cloneChild.orderedInvocation["child-b"])

	// Test that original scope is unaffected by clone operations
	is.Len(rootScope.self.childScopes["child1"].orderedInvocation, 2)
	is.Equal(0, rootScope.self.childScopes["child1"].orderedInvocation["child-a"])
	is.Equal(1, rootScope.self.childScopes["child1"].orderedInvocation["child-b"])

	// Test that services are properly cloned (same references for eager services)
	originalService, _ := rootScope.serviceGet("root-a")
	clonedService, _ := cloneRoot.serviceGet("root-a")
	is.Equal(originalService, clonedService)

	// Test that child services are properly cloned
	originalChildService, _ := child.serviceGet("child-a")
	clonedChildService, _ := cloneChild.serviceGet("child-a")
	is.Equal(originalChildService, clonedChildService)

	// Test that clone has correct scope hierarchy
	is.Equal(rootScope, cloneRoot.rootScope)
	is.Nil(cloneRoot.parentScope) // Clone root has no parent since it's the root
	is.Equal("child1", cloneChild.name)
	is.Equal(rootScope, cloneChild.rootScope)
	is.Equal(cloneRoot, cloneChild.parentScope)

	// Test that clone can have its own children
	cloneGrandchild := cloneChild.Scope("grandchild")
	is.NotNil(cloneGrandchild)
	is.Equal(cloneChild, cloneGrandchild.parentScope)
	is.Equal(rootScope, cloneGrandchild.rootScope)

	// Test that original scope is unaffected by clone's children
	is.Empty(rootScope.self.childScopes["child1"].childScopes)
	is.Len(cloneChild.childScopes, 1)
}

func TestScope_serviceHealthCheck(t *testing.T) {
	t.Parallel()
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

	is.ElementsMatch([]ServiceDescription{newServiceDescription(child3.id, child3.name, "child3-a"), newServiceDescription(child2a.id, child2a.name, "child2a-a"), newServiceDescription(child2a.id, child2a.name, "child2a-b"), newServiceDescription(child1.id, child1.name, "child1-a")}, child3.ListInvokedServices())
	shutdownReport := child1.Shutdown()
	is.Empty(shutdownReport.Errors)
	is.ElementsMatch([]ServiceDescription{}, child3.ListInvokedServices())
}

func TestScope_serviceGet(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

	// Test that serviceGetRec returns the correct scope
	service, scope, found := child2a.serviceGetRec("root-a")
	is.True(found)
	is.Equal(rootScope.self, scope)
	is.NotNil(service)

	service, scope, found = child2a.serviceGetRec("child1-a")
	is.True(found)
	is.Equal(child1, scope)
	is.NotNil(service)

	service, scope, found = child2a.serviceGetRec("child2a-a")
	is.True(found)
	is.Equal(child2a, scope)
	is.NotNil(service)

	service, scope, found = child2a.serviceGetRec("child2a-b")
	is.True(found)
	is.Equal(child2a, scope)
	is.NotNil(service)

	service, scope, found = child2a.serviceGetRec("child2b-a")
	is.False(found)
	is.Nil(scope)
	is.Nil(service)

	// Test from root scope perspective
	service, scope, found = rootScope.serviceGetRec("root-a")
	is.True(found)
	is.Equal(rootScope.self, scope)
	is.NotNil(service)

	service, scope, found = rootScope.serviceGetRec("child1-a")
	is.False(found)
	is.Nil(scope)
	is.Nil(service)
}

func TestScope_serviceSet(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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

	// Invoke services to make them shutdownable
	_, _ = invokeByName[*lazyTestShutdownerKO](rootScope, "root-a")
	_, _ = invokeByName[*lazyTestShutdownerOK](child1, "child1-a")
	_, _ = invokeByName[*lazyTestShutdownerOK](child2a, "child2a-a")
	_, _ = invokeByName[*lazyTestShutdownerKO](child2a, "child2a-b")
	_, _ = invokeByName[*lazyTestShutdownerKO](child2b, "child2b-a")

	// Test serviceShutdown from different scopes
	// from rootScope POV
	is.Equal(assert.AnError, rootScope.serviceShutdown(ctx, "root-a"))
	is.ErrorContains(rootScope.serviceShutdown(ctx, "child1-a"), "could not find service")
	is.ErrorContains(rootScope.serviceShutdown(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(rootScope.serviceShutdown(ctx, "child2a-b"), "could not find service")
	is.ErrorContains(rootScope.serviceShutdown(ctx, "child2b-a"), "could not find service")

	// from child1 POV
	is.ErrorContains(child1.serviceShutdown(ctx, "root-a"), "could not find service")
	is.NoError(child1.serviceShutdown(ctx, "child1-a"))
	is.ErrorContains(child1.serviceShutdown(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(child1.serviceShutdown(ctx, "child2a-b"), "could not find service")
	is.ErrorContains(child1.serviceShutdown(ctx, "child2b-a"), "could not find service")

	// from child2a POV
	is.ErrorContains(child2a.serviceShutdown(ctx, "root-a"), "could not find service")
	is.ErrorContains(child2a.serviceShutdown(ctx, "child1-a"), "could not find service")
	is.NoError(child2a.serviceShutdown(ctx, "child2a-a"))
	is.Equal(assert.AnError, child2a.serviceShutdown(ctx, "child2a-b"))
	is.ErrorContains(child2a.serviceShutdown(ctx, "child2b-a"), "could not find service")

	// from child2b POV
	is.ErrorContains(child2b.serviceShutdown(ctx, "root-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child1-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child2a-a"), "could not find service")
	is.ErrorContains(child2b.serviceShutdown(ctx, "child2a-b"), "could not find service")
	is.Equal(assert.AnError, child2b.serviceShutdown(ctx, "child2b-a"))

	// Test with different context scenarios
	// Test with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	is.ErrorContains(child2a.serviceShutdown(canceledCtx, "child2a-a"), "could not find service")

	// Test with timeout context
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	is.ErrorContains(child2a.serviceShutdown(timeoutCtx, "child2a-a"), "could not find service")

	// Test with non-existent service
	is.ErrorContains(child2a.serviceShutdown(ctx, "non-existent"), "could not find service")

	// Test with service that doesn't implement Shutdowner
	nonShutdownableScope := rootScope.Scope("non-shutdownable")
	nonShutdownableScope.serviceSet("non-shutdownable-a", newServiceLazy("non-shutdownable-a", func(i Injector) (int, error) { return 42, nil }))
	_, _ = invokeByName[int](nonShutdownableScope, "non-shutdownable-a")
	is.NoError(nonShutdownableScope.serviceShutdown(ctx, "non-shutdownable-a"))
}

func TestScope_onServiceRegistration(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test service registration functionality
	rootScope := New()

	// Register services
	rootScope.serviceSet("service1", newServiceEager("service1", "value1"))
	rootScope.serviceSet("service2", newServiceEager("service2", "value2"))

	// Verify services are registered
	service1, found1 := rootScope.serviceGet("service1")
	is.True(found1)
	is.NotNil(service1)

	service2, found2 := rootScope.serviceGet("service2")
	is.True(found2)
	is.NotNil(service2)

	// Test with child scope
	childScope := rootScope.Scope("child")

	// Register services in child scope
	childScope.serviceSet("child-service1", newServiceEager("child-service1", "child-value1"))
	childScope.serviceSet("child-service2", newServiceEager("child-service2", "child-value2"))

	// Verify child services are registered
	childService1, childFound1 := childScope.serviceGet("child-service1")
	is.True(childFound1)
	is.NotNil(childService1)

	childService2, childFound2 := childScope.serviceGet("child-service2")
	is.True(childFound2)
	is.NotNil(childService2)

	// Verify parent scope doesn't have access to child services
	_, parentFound1 := rootScope.serviceGet("child-service1")
	is.False(parentFound1)

	_, parentFound2 := rootScope.serviceGet("child-service2")
	is.False(parentFound2)

	// Test registration with nil service
	rootScope.serviceSet("nil-service", nil)
	nilService, nilFound := rootScope.serviceGet("nil-service")
	is.True(nilFound)
	is.Nil(nilService)

	// Test service replacement
	rootScope.serviceSet("service1", newServiceEager("service1", "new-value1"))
	newService1, newFound1 := rootScope.serviceGet("service1")
	is.True(newFound1)
	is.NotNil(newService1)
}

func TestScope_onServiceInvoke(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	rootScope := New()

	// Register and invoke services
	rootScope.serviceSet("service1", newServiceLazy("service1", func(i Injector) (string, error) { return "value1", nil }))
	rootScope.serviceSet("service2", newServiceLazy("service2", func(i Injector) (string, error) { return "value2", nil }))

	// Invoke services
	_, err1 := invokeByName[string](rootScope, "service1")
	is.NoError(err1)
	_, err2 := invokeByName[string](rootScope, "service2")
	is.NoError(err2)

	// Verify invocation tracking through orderedInvocation
	is.Equal(0, rootScope.self.orderedInvocation["service1"])
	is.Equal(1, rootScope.self.orderedInvocation["service2"])
	is.Equal(2, rootScope.self.orderedInvocationIndex)

	// Test with child scope
	childScope := rootScope.Scope("child")

	// Register and invoke services in child scope
	childScope.serviceSet("child-service1", newServiceLazy("child-service1", func(i Injector) (string, error) { return "child-value1", nil }))
	childScope.serviceSet("child-service2", newServiceLazy("child-service2", func(i Injector) (string, error) { return "child-value2", nil }))

	// Invoke child services
	_, childErr1 := invokeByName[string](childScope, "child-service1")
	is.NoError(childErr1)
	_, childErr2 := invokeByName[string](childScope, "child-service2")
	is.NoError(childErr2)

	// Verify child invocation tracking
	is.Equal(0, childScope.orderedInvocation["child-service1"])
	is.Equal(1, childScope.orderedInvocation["child-service2"])
	is.Equal(2, childScope.orderedInvocationIndex)

	// Verify parent invocation tracking is unchanged
	is.Equal(0, rootScope.self.orderedInvocation["service1"])
	is.Equal(1, rootScope.self.orderedInvocation["service2"])
	is.Equal(2, rootScope.self.orderedInvocationIndex)

	// Test multiple invocations of the same service
	_, err3 := invokeByName[string](rootScope, "service1")
	is.NoError(err3)

	// Verify that orderedInvocation is not changed for repeated invocations
	is.Equal(0, rootScope.self.orderedInvocation["service1"])
	is.Equal(1, rootScope.self.orderedInvocation["service2"])
	is.Equal(2, rootScope.self.orderedInvocationIndex)

	// Test invocation of parent service from child scope
	_, parentErr := invokeByName[string](childScope, "service1")
	is.NoError(parentErr)

	// Verify that parent's orderedInvocation is updated when child invokes parent service
	is.Equal(0, rootScope.self.orderedInvocation["service1"])
	is.Equal(1, rootScope.self.orderedInvocation["service2"])
	is.Equal(2, rootScope.self.orderedInvocationIndex) // Should still be 2, not 3, since service1 was already invoked

	// Test that onServiceInvoke is called automatically during service invocation
	// This is verified by checking that orderedInvocation is properly maintained
	is.Equal(0, childScope.orderedInvocation["child-service1"])
	is.Equal(1, childScope.orderedInvocation["child-service2"])
	is.Equal(2, childScope.orderedInvocationIndex)
}

func TestScope_onServiceShutdown(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	ctx := context.Background()
	rootScope := New()

	// Register shutdownable services
	rootScope.serviceSet("shutdown-ok", newServiceLazy("shutdown-ok", func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "ok"}, nil
	}))
	rootScope.serviceSet("shutdown-ko", newServiceLazy("shutdown-ko", func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "ko"}, nil
	}))

	// Invoke services to make them shutdownable
	_, err1 := invokeByName[*lazyTestShutdownerOK](rootScope, "shutdown-ok")
	is.NoError(err1)
	_, err2 := invokeByName[*lazyTestShutdownerKO](rootScope, "shutdown-ko")
	is.NoError(err2)

	// Test successful shutdown
	err := rootScope.serviceShutdown(ctx, "shutdown-ok")
	is.NoError(err)

	// Verify service is removed from scope
	_, found := rootScope.serviceGet("shutdown-ok")
	is.False(found)

	// Test failed shutdown
	err = rootScope.serviceShutdown(ctx, "shutdown-ko")
	is.Equal(assert.AnError, err)

	// Verify service is still removed from scope even if shutdown fails
	_, found = rootScope.serviceGet("shutdown-ko")
	is.False(found)

	// Test shutdown of non-existent service
	err = rootScope.serviceShutdown(ctx, "non-existent")
	is.ErrorContains(err, "could not find service")

	// Test shutdown of service that doesn't implement Shutdowner
	rootScope.serviceSet("non-shutdownable", newServiceLazy("non-shutdownable", func(i Injector) (int, error) { return 42, nil }))
	_, err3 := invokeByName[int](rootScope, "non-shutdownable")
	is.NoError(err3)

	// This should not panic because the service wrapper handles non-shutdowner services gracefully
	err = rootScope.serviceShutdown(ctx, "non-shutdownable")
	is.NoError(err)

	// Test shutdown with different context scenarios
	rootScope.serviceSet("shutdown-test", newServiceLazy("shutdown-test", func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "test"}, nil
	}))
	_, err4 := invokeByName[*lazyTestShutdownerOK](rootScope, "shutdown-test")
	is.NoError(err4)

	// Test with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = rootScope.serviceShutdown(canceledCtx, "shutdown-test")
	is.Equal(context.Canceled, err)

	// Test with timeout context
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)
	err = rootScope.serviceShutdown(timeoutCtx, "shutdown-test")
	is.ErrorContains(err, "could not find service")
}

func TestScope_logs(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test logging functionality with custom log function
	logMessages := []string{}
	logf := func(format string, args ...any) {
		message := fmt.Sprintf(format, args...)
		logMessages = append(logMessages, message)
	}

	// Create injector with custom logging
	injector := NewWithOpts(&InjectorOpts{
		Logf: logf,
	})

	// Test logging during service registration
	injector.serviceSet("test-service", newServiceEager("test-service", "test-value"))

	// Test logging during service invocation
	_, err := invokeByName[string](injector, "test-service")
	is.NoError(err)

	// Test logging during health check
	healthcheckService := newServiceLazy("healthcheck-service", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "health"}, nil
	})
	injector.serviceSet("healthcheck-service", healthcheckService)
	_, err = invokeByName[*lazyTestHeathcheckerOK](injector, "healthcheck-service")
	is.NoError(err)

	ctx := context.Background()
	_ = injector.serviceHealthCheck(ctx, "healthcheck-service")

	// Test logging during shutdown
	shutdownService := newServiceLazy("shutdown-service", func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "shutdown"}, nil
	})
	injector.serviceSet("shutdown-service", shutdownService)
	_, err = invokeByName[*lazyTestShutdownerOK](injector, "shutdown-service")
	is.NoError(err)

	_ = injector.serviceShutdown(ctx, "shutdown-service")

	// Verify that log messages were generated
	is.NotEmpty(logMessages)

	// Check for specific log message patterns
	hasHealthCheckLog := false
	hasShutdownLog := false
	for _, message := range logMessages {
		if strings.Contains(message, "health check") {
			hasHealthCheckLog = true
		}
		if strings.Contains(message, "shutdown") {
			hasShutdownLog = true
		}
	}

	is.True(hasHealthCheckLog, "Should have health check log message")
	is.True(hasShutdownLog, "Should have shutdown log message")

	// Test logging without custom log function (should not panic)
	injectorNoLog := New()
	injectorNoLog.serviceSet("no-log-service", newServiceEager("no-log-service", "no-log-value"))
	_, err = invokeByName[string](injectorNoLog, "no-log-service")
	is.NoError(err)

	// Test logging with nil log function (should not panic)
	injectorNilLog := NewWithOpts(&InjectorOpts{
		Logf: nil,
	})
	injectorNilLog.serviceSet("nil-log-service", newServiceEager("nil-log-service", "nil-log-value"))
	_, err = invokeByName[string](injectorNilLog, "nil-log-service")
	is.NoError(err)

	// Test logging in child scopes
	childScope := injector.Scope("child")
	childScope.serviceSet("child-service", newServiceEager("child-service", "child-value"))
	_, err = invokeByName[string](childScope, "child-service")
	is.NoError(err)

	// Verify that child scope logging also works
	is.NotEmpty(logMessages)
}

type shutdownRecorder struct{ called *int32 }

func (s *shutdownRecorder) Shutdown() error {
	atomic.AddInt32(s.called, 1)
	return nil
}

// Test the shutdown of services that have no dependencies and no dependents.
func TestScope_Shutdown_NoDependencies(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()

	var a, b, c int32
	ProvideNamedValue(i, "svc-a", &shutdownRecorder{called: &a})
	ProvideNamedValue(i, "svc-b", &shutdownRecorder{called: &b})
	ProvideNamedValue(i, "svc-c", &shutdownRecorder{called: &c})

	// No dependencies between services; shutdown should call all three
	report := i.Shutdown()
	is.True(report.Succeed)
	is.Empty(report.Errors)

	is.Equal(int32(1), atomic.LoadInt32(&a))
	is.Equal(int32(1), atomic.LoadInt32(&b))
	is.Equal(int32(1), atomic.LoadInt32(&c))

	// Services should be removed after shutdown
	_, errA := InvokeNamed[*shutdownRecorder](i, "svc-a")
	_, errB := InvokeNamed[*shutdownRecorder](i, "svc-b")
	_, errC := InvokeNamed[*shutdownRecorder](i, "svc-c")
	is.Error(errA)
	is.Error(errB)
	is.Error(errC)
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
	t.Parallel()
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
	is.Len(errors.Errors, 1)

	// Check that the service was attempted to be shut down
	is.Equal(1, slowService.getShutdownCount())
}

func TestScope_ShutdownWithContextExpiration_Cancellation(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)
	injector := New()

	// Create a blocking service
	blockingService := newScopeTestBlockingShutdowner()
	ProvideNamedValue(injector, "blocking-service", blockingService)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start shutdown in goroutine
	var report *ShutdownReport
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		report = injector.ShutdownWithContext(ctx)
	}()

	// Cancel context after short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for shutdown to complete
	wg.Wait()

	// Should have shutdown errors due to cancellation
	is.NotNil(report)
	is.False(report.Succeed)
	is.Len(report.Errors, 1)

	// Check that the service was attempted to be shut down
	is.Equal(1, blockingService.getShutdownCount())
}

func TestScope_ShutdownWithContextExpiration_MultipleServices(t *testing.T) {
	t.Parallel()
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
	is.Len(errors.Errors, 1)

	// Both services should have been attempted
	is.Equal(1, fastService.getShutdownCount())
	is.Equal(1, slowService.getShutdownCount())
}

func TestScope_ShutdownWithContextExpiration_ChildScopes(t *testing.T) {
	t.Parallel()
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
	is.NoError(err1)
	is.NoError(err2)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Shutdown should timeout - call shutdown on root scope should shut down all child scopes
	errors := injector.ShutdownWithContext(ctx)
	is.Len(errors.Errors, 2)

	// Child service should have been attempted (child scopes are shut down first)
	is.Equal(1, childService.getShutdownCount())

	// Parent service might not be attempted if context times out during child shutdown
	// This is expected behavior for this test - we only verify child scope shutdown works
	is.Equal(0, parentService.getShutdownCount())
}

// Test healthcheck context expiration
func TestScope_HealthCheckWithContextExpiration_Timeout(t *testing.T) {
	t.Parallel()
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
		is.Error(err)
		// Check if it's a timeout error (could be wrapped)
		is.True(errors.Is(err, context.DeadlineExceeded) || err.Error() == "DI: health check timeout: context deadline exceeded")
	}

	// Check that the service was attempted to be health checked
	is.Equal(1, slowService.getHealthcheckCount())
}

func TestScope_HealthCheckWithContextExpiration_Cancellation(t *testing.T) {
	t.Parallel()
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
		is.Error(err)
		// Check if it's a cancellation error (could be wrapped)
		is.True(errors.Is(err, context.Canceled) || err.Error() == "DI: health check timeout: context canceled")
	}

	// Check that the service was attempted to be health checked
	is.Equal(1, blockingService.getHealthcheckCount())
}

func TestScope_HealthCheckWithContextExpiration_MultipleServices(t *testing.T) {
	t.Parallel()
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
	is.NoError(err1)
	is.NoError(err2)

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
	t.Parallel()
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
	is.NoError(err1)
	is.NoError(err2)

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
	t.Parallel()
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
	is.NoError(err)

	// Healthcheck should timeout due to global timeout
	start := time.Now()
	results := injector.HealthCheckWithContext(context.Background())
	duration := time.Since(start)

	// Should complete quickly due to global timeout
	is.Less(duration, 100*time.Millisecond)

	// Should have healthcheck errors due to timeout
	is.NotEmpty(results)
	for _, err := range results {
		is.Error(err)
		// Check if it's a timeout error (could be wrapped)
		is.True(errors.Is(err, context.DeadlineExceeded) || err.Error() == "DI: health check timeout: context deadline exceeded")
	}

	// Check that the service was attempted to be health checked
	is.Equal(1, slowService.getHealthcheckCount())
}

// Test mixed scenarios
func TestScope_ContextExpiration_ShutdownAndHealthcheckSameTimeout(t *testing.T) {
	t.Parallel()
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
	report := injector.ShutdownWithContext(ctx)
	is.NotNil(report)
	is.False(report.Succeed)
	is.Len(report.Errors, 1)

	// Recreate injector for healthcheck test
	injector2 := New()
	ProvideNamedValue(injector2, "healthcheck-service", healthcheckService)

	// Test healthcheck timeout
	healthcheckResults := injector2.HealthCheckWithContext(ctx)
	is.NotEmpty(healthcheckResults)
	is.Len(healthcheckResults, 1)
	for _, err := range healthcheckResults {
		is.Error(err)
		// Check if it's a timeout error (could be wrapped)
		is.True(errors.Is(err, context.DeadlineExceeded) || err.Error() == "DI: health check timeout: context deadline exceeded")
	}
}

func TestScope_ContextExpiration_ParallelOperations(t *testing.T) {
	t.Parallel()
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
	var report *ShutdownReport
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		report = injector.ShutdownWithContext(ctx)
	}()

	// Cancel context after short delay
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for shutdown to complete
	wg.Wait()

	// Should have shutdown errors due to cancellation
	is.NotNil(report)
	is.False(report.Succeed)
	is.Len(report.Errors, 5)

	// All services should have been attempted
	for _, service := range services {
		is.Equal(1, service.getShutdownCount())
	}
}
