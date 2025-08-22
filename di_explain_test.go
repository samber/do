package do

import (
	"os"
	"testing"
	"time"

	"github.com/samber/do/v2/stacktrace"
	"github.com/stretchr/testify/assert"
)

/////////////////////////////////////////////////////////////////////////////
// 							Templating helpers
/////////////////////////////////////////////////////////////////////////////

var dirname = must1(os.Getwd())

func fakeProvider1(i Injector) (int, error) {
	return 42, nil
}

func fakeProvider2(i Injector) (int, error) {
	_ = MustInvokeNamed[int](i, "SERVICE-A1")
	_ = MustInvokeNamed[int](i, "SERVICE-A2")
	return 42, nil
}

func fakeProvider3(i Injector) (int, error) {
	_ = MustInvokeNamed[int](i, "SERVICE-B")
	return 42, nil
}

func fakeProvider4(i Injector) (int, error) {
	_ = MustInvokeNamed[int](i, "SERVICE-C1")
	_ = MustInvokeNamed[int](i, "SERVICE-C2")
	return 42, nil
}

func fakeProvider5(i Injector) (int, error) {
	_ = MustInvokeNamed[int](i, "SERVICE-D")
	return 42, nil
}

func fakeProvider6(i Injector) (int, error) {
	_ = MustInvokeNamed[int](i, "SERVICE-E")
	return 42, nil
}

func TestFromTemplate(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	output := fromTemplate("foobar", nil)
	is.Equal("foobar", output)

	output = fromTemplate("foobar", map[string]any{"foo": "bar"})
	is.Equal("foobar", output)

	output = fromTemplate("foo {{.Bar}}", map[string]any{"Bar": "bar"})
	is.Equal("foo bar", output)

	output = fromTemplate("foo {{.Foo}}", map[string]any{"Baz": "bar"})
	is.Equal("foo ", output)
}

/////////////////////////////////////////////////////////////////////////////
// 							Explain services
/////////////////////////////////////////////////////////////////////////////

func TestExplainService_String(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test with all fields populated
	output := ExplainServiceOutput{
		ScopeID:          "scope-123",
		ScopeName:        "test-scope",
		ServiceName:      "test-service",
		ServiceType:      ServiceTypeLazy,
		ServiceBuildTime: 2 * time.Second,
		Invoked:          &stacktrace.Frame{File: "test.go", Line: 42, Function: "provider"},
		Dependencies: []ExplainServiceDependencyOutput{
			{ScopeID: "scope-123", ScopeName: "test-scope", Service: "dep1", Recursive: []ExplainServiceDependencyOutput{}},
			{ScopeID: "scope-123", ScopeName: "test-scope", Service: "dep2", Recursive: []ExplainServiceDependencyOutput{}},
		},
		Dependents: []ExplainServiceDependencyOutput{
			{ScopeID: "scope-123", ScopeName: "test-scope", Service: "dependent1", Recursive: []ExplainServiceDependencyOutput{}},
			{ScopeID: "scope-123", ScopeName: "test-scope", Service: "dependent2", Recursive: []ExplainServiceDependencyOutput{}},
		},
	}

	expected := `
Scope ID: scope-123
Scope name: test-scope

Service name: test-service
Service type: lazy
Service build time: 2s
Invoked: test.go:provider:42

Dependencies:
* dep1 from scope test-scope
* dep2 from scope test-scope

Dependents:
* dependent1 from scope test-scope
* dependent2 from scope test-scope
`
	is.Equal(expected, output.String())

	// Test with minimal fields
	output2 := ExplainServiceOutput{
		ScopeID:     "scope-123",
		ScopeName:   "test-scope",
		ServiceName: "test-service",
		ServiceType: ServiceTypeEager,
	}

	expected2 := `
Scope ID: scope-123
Scope name: test-scope

Service name: test-service
Service type: eager
Invoked: 

Dependencies:


Dependents:

`
	is.Equal(expected2, output2.String())

	// Test with no dependencies or dependents
	output3 := ExplainServiceOutput{
		ScopeID:      "scope-123",
		ScopeName:    "test-scope",
		ServiceName:  "test-service",
		ServiceType:  ServiceTypeTransient,
		Dependencies: []ExplainServiceDependencyOutput{},
		Dependents:   []ExplainServiceDependencyOutput{},
	}

	expected3 := `
Scope ID: scope-123
Scope name: test-scope

Service name: test-service
Service type: transient
Invoked: 

Dependencies:


Dependents:

`
	is.Equal(expected3, output3.String())
}

func TestExplainServiceDependency_String(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	a1 := ExplainServiceDependencyOutput{
		ScopeID:   "1234",
		ScopeName: "scope-a",
		Service:   "service-a1",
		Recursive: []ExplainServiceDependencyOutput{},
	}
	a2 := ExplainServiceDependencyOutput{
		ScopeID:   "1234",
		ScopeName: "scope-a",
		Service:   "service-a2",
		Recursive: []ExplainServiceDependencyOutput{},
	}
	b := ExplainServiceDependencyOutput{
		ScopeID:   "1234",
		ScopeName: "scope-a",
		Service:   "service-b",
		Recursive: []ExplainServiceDependencyOutput{a1, a2},
	}
	c := ExplainServiceDependencyOutput{
		ScopeID:   "1234",
		ScopeName: "scope-a",
		Service:   "service-c",
		Recursive: []ExplainServiceDependencyOutput{b},
	}

	expected := `* service-a1 from scope scope-a`
	is.Equal(expected, a1.String())

	expected = `* service-b from scope scope-a
  * service-a1 from scope scope-a
  * service-a2 from scope scope-a`
	is.Equal(expected, b.String())

	expected = `* service-c from scope scope-a
  * service-b from scope scope-a
    * service-a1 from scope scope-a
    * service-a2 from scope scope-a`
	is.Equal(expected, c.String())
}

func TestExplainService(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// prepare env
	i := New()
	scope := i.Scope("scope-child")
	scope.id = "scope-id-123"
	ProvideNamed(i, "SERVICE-A1", fakeProvider1)
	ProvideNamed(i, "SERVICE-A2", fakeProvider1)
	ProvideNamed(i, "SERVICE-B", fakeProvider2)
	ProvideNamed(scope, "SERVICE-C1", fakeProvider3)
	ProvideNamed(scope, "SERVICE-C2", fakeProvider3)
	ProvideNamed(scope, "SERVICE-D", fakeProvider4)
	ProvideNamed(scope, "SERVICE-E", fakeProvider5)
	_, _ = InvokeNamed[int](scope, "SERVICE-E")

	// Test explaining a service by type (needs to be invoked first)
	_, _ = InvokeNamed[int](scope, "SERVICE-E")
	output, ok := ExplainNamedService(scope, "SERVICE-E")
	is.True(ok)
	is.NotNil(output)
	is.Equal("SERVICE-E", output.ServiceName)
	is.Equal(ServiceTypeLazy, output.ServiceType)
	is.Equal("scope-id-123", output.ScopeID)
	is.Equal("scope-child", output.ScopeName)

	// Test explaining a service that doesn't exist
	output2, ok2 := ExplainService[string](scope)
	is.False(ok2)
	is.Equal(ExplainServiceOutput{}, output2)

	// Test explaining a service that exists but hasn't been invoked
	_, _ = InvokeNamed[int](i, "SERVICE-A1")
	output3, ok3 := ExplainNamedService(i, "SERVICE-A1")
	is.True(ok3)
	is.NotNil(output3)
	is.Equal("SERVICE-A1", output3.ServiceName)
	is.Equal(ServiceTypeLazy, output3.ServiceType)
	is.NotEmpty(output3.ScopeID) // Root scope has a generated ID
	is.Equal("[root]", output3.ScopeName)
}

func TestExplainNamedService(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// prepare env
	i := New()
	scope := i.Scope("scope-child")
	scope.id = "scope-id-123"
	ProvideNamed(i, "SERVICE-A1", fakeProvider1)
	ProvideNamed(i, "SERVICE-A2", fakeProvider1)
	ProvideNamed(i, "SERVICE-B", fakeProvider2)
	ProvideNamed(scope, "SERVICE-C1", fakeProvider3)
	ProvideNamed(scope, "SERVICE-C2", fakeProvider3)
	ProvideNamed(scope, "SERVICE-D", fakeProvider4)
	ProvideNamed(scope, "SERVICE-E", fakeProvider5)
	ProvideNamed(scope, "SERVICE-F", fakeProvider6)
	_, _ = InvokeNamed[int](scope, "SERVICE-F")

	// full explain
	expected := `
Scope ID: scope-id-123
Scope name: scope-child

Service name: SERVICE-E
Service type: lazy
Service build time: 1s
Invoked: ` + dirname + `/di_explain_test.go:fakeProvider5:39

Dependencies:
* SERVICE-D from scope scope-child
  * SERVICE-C1 from scope scope-child
    * SERVICE-B from scope [root]
      * SERVICE-A1 from scope [root]
      * SERVICE-A2 from scope [root]
  * SERVICE-C2 from scope scope-child
    * SERVICE-B from scope [root]
      * SERVICE-A1 from scope [root]
      * SERVICE-A2 from scope [root]

Dependents:
* SERVICE-F from scope scope-child
`
	output, ok := ExplainNamedService(scope, "SERVICE-E")
	is.True(ok)
	output.ServiceBuildTime = 1 * time.Second
	is.Equal(expected, output.String())

	// same test, but without build time
	expected = `
Scope ID: scope-id-123
Scope name: scope-child

Service name: SERVICE-E
Service type: lazy
Invoked: ` + dirname + `/di_explain_test.go:fakeProvider5:39

Dependencies:
* SERVICE-D from scope scope-child
  * SERVICE-C1 from scope scope-child
    * SERVICE-B from scope [root]
      * SERVICE-A1 from scope [root]
      * SERVICE-A2 from scope [root]
  * SERVICE-C2 from scope scope-child
    * SERVICE-B from scope [root]
      * SERVICE-A1 from scope [root]
      * SERVICE-A2 from scope [root]

Dependents:
* SERVICE-F from scope scope-child
`
	output.ServiceBuildTime = 0
	is.Equal(expected, output.String())

	// service not found
	output, ok = ExplainNamedService(scope, "not_found")
	is.False(ok)
	is.Equal(ExplainServiceOutput{}, output)
}

/////////////////////////////////////////////////////////////////////////////
// 							Explain scopes
/////////////////////////////////////////////////////////////////////////////

func TestExplainInjector_String(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test with all fields populated
	output := ExplainInjectorOutput{
		ScopeID:   "scope-123",
		ScopeName: "test-scope",
		DAG: []ExplainInjectorScopeOutput{
			{
				ScopeID:   "scope-123",
				ScopeName: "test-scope",
				Services: []ExplainInjectorServiceOutput{
					{
						ServiceName:      "service1",
						ServiceType:      ServiceTypeLazy,
						ServiceTypeIcon:  "😴",
						ServiceBuildTime: 1 * time.Second,
						IsHealthchecker:  true,
						IsShutdowner:     false,
					},
					{
						ServiceName:      "service2",
						ServiceType:      ServiceTypeEager,
						ServiceTypeIcon:  "🔁",
						ServiceBuildTime: 0,
						IsHealthchecker:  false,
						IsShutdowner:     true,
					},
				},
				Children: []ExplainInjectorScopeOutput{
					{
						ScopeID:   "child-123",
						ScopeName: "child-scope",
						Services:  []ExplainInjectorServiceOutput{},
						Children:  []ExplainInjectorScopeOutput{},
					},
				},
			},
		},
	}

	expected := `Scope ID: scope-123
Scope name: test-scope

DAG:
 |
  \_ test-scope (ID: scope-123)
      * 😴 service1 🫀
      * 🔁 service2 🙅
      |
      |
       \_ child-scope (ID: child-123)
`
	is.Equal(expected, output.String())

	// Test with minimal fields
	output2 := ExplainInjectorOutput{
		ScopeID:   "scope-123",
		ScopeName: "test-scope",
		DAG:       []ExplainInjectorScopeOutput{},
	}

	expected2 := `Scope ID: scope-123
Scope name: test-scope

DAG:

`
	is.Equal(expected2, output2.String())
}

func TestExplainInjectorScope_String(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test with all fields populated
	output := ExplainInjectorScopeOutput{
		ScopeID:   "scope-123",
		ScopeName: "test-scope",
		Services: []ExplainInjectorServiceOutput{
			{
				ServiceName:      "service1",
				ServiceType:      ServiceTypeLazy,
				ServiceTypeIcon:  "😴",
				ServiceBuildTime: 1 * time.Second,
				IsHealthchecker:  true,
				IsShutdowner:     false,
			},
			{
				ServiceName:      "service2",
				ServiceType:      ServiceTypeEager,
				ServiceTypeIcon:  "🔁",
				ServiceBuildTime: 0,
				IsHealthchecker:  false,
				IsShutdowner:     true,
			},
		},
		Children: []ExplainInjectorScopeOutput{
			{
				ScopeID:   "child-123",
				ScopeName: "child-scope",
				Services:  []ExplainInjectorServiceOutput{},
				Children:  []ExplainInjectorScopeOutput{},
			},
		},
		IsAncestor: false,
		IsChildren: true,
	}

	expected := `test-scope (ID: scope-123)
 * 😴 service1 🫀
 * 🔁 service2 🙅
 |
 |
  \_ child-scope (ID: child-123)`
	is.Equal(expected, output.String())

	// Test with minimal fields
	output2 := ExplainInjectorScopeOutput{
		ScopeID:    "scope-123",
		ScopeName:  "test-scope",
		Services:   []ExplainInjectorServiceOutput{},
		Children:   []ExplainInjectorScopeOutput{},
		IsAncestor: true,
		IsChildren: false,
	}

	expected2 := `test-scope (ID: scope-123)`
	is.Equal(expected2, output2.String())
}

func TestExplainInjectorService_String(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	svc := ExplainInjectorServiceOutput{ServiceName: "service-name", ServiceType: ServiceTypeLazy, ServiceTypeIcon: "😴", ServiceBuildTime: 1 * time.Second, IsHealthchecker: true, IsShutdowner: true}
	expected := ` * 😴 service-name 🫀 🙅`
	is.Equal(expected, svc.String())

	svc = ExplainInjectorServiceOutput{ServiceName: "service-name", ServiceType: ServiceTypeEager, ServiceTypeIcon: "🔁", IsHealthchecker: true, IsShutdowner: false}
	expected = ` * 🔁 service-name 🫀`
	is.Equal(expected, svc.String())

	svc = ExplainInjectorServiceOutput{ServiceName: "service-name", IsHealthchecker: true, IsShutdowner: true}
	expected = ` * ❓ service-name`
	is.Equal(expected, svc.String())
}

func TestExplainInjector(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// prepare env
	i := New()
	i.self.id = "scope-id-root"
	scope0 := i.Scope("scope-0")
	scope0.id = "scope-id-0"
	scope1a := scope0.Scope("scope-1a")
	scope1a.id = "scope-id-1a"
	scope1b := scope0.Scope("scope-1b")
	scope1b.id = "scope-id-1b"
	scope2a := scope1a.Scope("scope-2a")
	scope2a.id = "scope-id-2a"
	scope2b := scope1a.Scope("scope-2b")
	scope2b.id = "scope-id-2b"
	ProvideNamed(i, "SERVICE-A1", fakeProvider1)
	ProvideNamed(i, "SERVICE-A2", fakeProvider1)
	ProvideNamed(i, "SERVICE-B", fakeProvider2)
	ProvideNamed(scope1a, "SERVICE-C1", fakeProvider3)
	ProvideNamed(scope1a, "SERVICE-C2", fakeProvider3)
	ProvideNamed(scope1a, "SERVICE-D", fakeProvider4)
	ProvideNamed(scope1a, "SERVICE-E", fakeProvider5)
	ProvideNamed(scope1b, "SERVICE-F", fakeProvider6)
	ProvideNamedTransient[*lazyTest](scope2a, "SERVICE-TRANSIENT-SIMPLE", func(i Injector) (*lazyTest, error) { return &lazyTest{}, nil })
	ProvideNamed[*lazyTestHeathcheckerOK](scope2a, "SERVICE-LAZY-HEALTH", func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })
	ProvideNamed[*lazyTestShutdownerOK](scope2b, "SERVICE-LAZY-SHUTDOWN", func(i Injector) (*lazyTestShutdownerOK, error) { return &lazyTestShutdownerOK{}, nil })
	ProvideNamedValue[int](scope1a, "SERVICE-EAGER-VALUE", 1)
	_ = AsNamed[*lazyTestHeathcheckerOK, Healthchecker](scope2a, "SERVICE-LAZY-HEALTH", "SERVICE-ALIAS-HEALTH")
	_, _ = InvokeNamed[int](scope1a, "SERVICE-D")
	_, _ = InvokeNamed[*lazyTestHeathcheckerOK](scope2a, "SERVICE-LAZY-HEALTH")
	_, _ = InvokeNamed[*lazyTestShutdownerOK](scope2b, "SERVICE-LAZY-SHUTDOWN")

	// from root POV
	expected := `Scope ID: scope-id-root
Scope name: [root]

DAG:
 |
  \_ [root] (ID: scope-id-root)
      * 😴 SERVICE-A1
      * 😴 SERVICE-A2
      * 😴 SERVICE-B
      |
      |
       \_ scope-0 (ID: scope-id-0)
           |
           |
           |\_ scope-1a (ID: scope-id-1a)
           |    * 😴 SERVICE-C1
           |    * 😴 SERVICE-C2
           |    * 😴 SERVICE-D
           |    * 😴 SERVICE-E
           |    * 🔁 SERVICE-EAGER-VALUE
           |    |
           |    |
           |    |\_ scope-2a (ID: scope-id-2a)
           |    |    * 🔗 SERVICE-ALIAS-HEALTH 🫀
           |    |    * 😴 SERVICE-LAZY-HEALTH 🫀
           |    |    * 🏭 SERVICE-TRANSIENT-SIMPLE
           |    |
           |    |
           |     \_ scope-2b (ID: scope-id-2b)
           |         * 😴 SERVICE-LAZY-SHUTDOWN 🙅
           |
           |
            \_ scope-1b (ID: scope-id-1b)
                * 😴 SERVICE-F
`
	output := ExplainInjector(i)
	is.Equal(expected, output.String())

	// from scope0 POV
	expected = `Scope ID: scope-id-0
Scope name: scope-0

DAG:
 |
  \_ [root] (ID: scope-id-root)
      * 😴 SERVICE-A1
      * 😴 SERVICE-A2
      * 😴 SERVICE-B
      |
      |
       \_ scope-0 (ID: scope-id-0)
           |
           |
           |\_ scope-1a (ID: scope-id-1a)
           |    * 😴 SERVICE-C1
           |    * 😴 SERVICE-C2
           |    * 😴 SERVICE-D
           |    * 😴 SERVICE-E
           |    * 🔁 SERVICE-EAGER-VALUE
           |    |
           |    |
           |    |\_ scope-2a (ID: scope-id-2a)
           |    |    * 🔗 SERVICE-ALIAS-HEALTH 🫀
           |    |    * 😴 SERVICE-LAZY-HEALTH 🫀
           |    |    * 🏭 SERVICE-TRANSIENT-SIMPLE
           |    |
           |    |
           |     \_ scope-2b (ID: scope-id-2b)
           |         * 😴 SERVICE-LAZY-SHUTDOWN 🙅
           |
           |
            \_ scope-1b (ID: scope-id-1b)
                * 😴 SERVICE-F
`
	output = ExplainInjector(scope0)
	is.Equal(expected, output.String())

	// from scope1a POV
	expected = `Scope ID: scope-id-1a
Scope name: scope-1a

DAG:
 |
  \_ [root] (ID: scope-id-root)
      * 😴 SERVICE-A1
      * 😴 SERVICE-A2
      * 😴 SERVICE-B
      |
      |
       \_ scope-0 (ID: scope-id-0)
           |
           |
            \_ scope-1a (ID: scope-id-1a)
                * 😴 SERVICE-C1
                * 😴 SERVICE-C2
                * 😴 SERVICE-D
                * 😴 SERVICE-E
                * 🔁 SERVICE-EAGER-VALUE
                |
                |
                |\_ scope-2a (ID: scope-id-2a)
                |    * 🔗 SERVICE-ALIAS-HEALTH 🫀
                |    * 😴 SERVICE-LAZY-HEALTH 🫀
                |    * 🏭 SERVICE-TRANSIENT-SIMPLE
                |
                |
                 \_ scope-2b (ID: scope-id-2b)
                     * 😴 SERVICE-LAZY-SHUTDOWN 🙅
`
	output = ExplainInjector(scope1a)
	is.Equal(expected, output.String())

	// from scope1b POV
	expected = `Scope ID: scope-id-1b
Scope name: scope-1b

DAG:
 |
  \_ [root] (ID: scope-id-root)
      * 😴 SERVICE-A1
      * 😴 SERVICE-A2
      * 😴 SERVICE-B
      |
      |
       \_ scope-0 (ID: scope-id-0)
           |
           |
            \_ scope-1b (ID: scope-id-1b)
                * 😴 SERVICE-F
`
	output = ExplainInjector(scope1b)
	is.Equal(expected, output.String())

	// from scope2a POV
	expected = `Scope ID: scope-id-2a
Scope name: scope-2a

DAG:
 |
  \_ [root] (ID: scope-id-root)
      * 😴 SERVICE-A1
      * 😴 SERVICE-A2
      * 😴 SERVICE-B
      |
      |
       \_ scope-0 (ID: scope-id-0)
           |
           |
            \_ scope-1a (ID: scope-id-1a)
                * 😴 SERVICE-C1
                * 😴 SERVICE-C2
                * 😴 SERVICE-D
                * 😴 SERVICE-E
                * 🔁 SERVICE-EAGER-VALUE
                |
                |
                 \_ scope-2a (ID: scope-id-2a)
                     * 🔗 SERVICE-ALIAS-HEALTH 🫀
                     * 😴 SERVICE-LAZY-HEALTH 🫀
                     * 🏭 SERVICE-TRANSIENT-SIMPLE
`
	output = ExplainInjector(scope2a)
	is.Equal(expected, output.String())

	// from scope2b POV
	expected = `Scope ID: scope-id-2b
Scope name: scope-2b

DAG:
 |
  \_ [root] (ID: scope-id-root)
      * 😴 SERVICE-A1
      * 😴 SERVICE-A2
      * 😴 SERVICE-B
      |
      |
       \_ scope-0 (ID: scope-id-0)
           |
           |
            \_ scope-1a (ID: scope-id-1a)
                * 😴 SERVICE-C1
                * 😴 SERVICE-C2
                * 😴 SERVICE-D
                * 😴 SERVICE-E
                * 🔁 SERVICE-EAGER-VALUE
                |
                |
                 \_ scope-2b (ID: scope-id-2b)
                     * 😴 SERVICE-LAZY-SHUTDOWN 🙅
`
	output = ExplainInjector(scope2b)
	is.Equal(expected, output.String())
}
