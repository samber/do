package do

import (
	"os"
	"testing"

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
// 							Describe services
/////////////////////////////////////////////////////////////////////////////

func TestDescriptionService_String(t *testing.T) {
	// @TODO
}

func TestDescriptionServiceDependency_String(t *testing.T) {
	is := assert.New(t)

	a1 := DescriptionServiceDependency{
		ScopeID:   "1234",
		ScopeName: "scope-a",
		Service:   "service-a1",
		Recursive: []DescriptionServiceDependency{},
	}
	a2 := DescriptionServiceDependency{
		ScopeID:   "1234",
		ScopeName: "scope-a",
		Service:   "service-a2",
		Recursive: []DescriptionServiceDependency{},
	}
	b := DescriptionServiceDependency{
		ScopeID:   "1234",
		ScopeName: "scope-a",
		Service:   "service-b",
		Recursive: []DescriptionServiceDependency{a1, a2},
	}
	c := DescriptionServiceDependency{
		ScopeID:   "1234",
		ScopeName: "scope-a",
		Service:   "service-c",
		Recursive: []DescriptionServiceDependency{b},
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

func TestDescribeService(t *testing.T) {
	// @TODO
}

func TestDescribeNamedService(t *testing.T) {
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

	// full describe
	expected := `
Scope ID: scope-id-123
Scope name: scope-child

Service name: SERVICE-E
Service type: lazy
Invoked: ` + dirname + `/di_describe_test.go:fakeProvider5:37

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
	output, ok := DescribeNamedService(scope, "SERVICE-E")
	is.True(ok)
	is.Equal(expected, output.String())

	// service not found
	output, ok = DescribeNamedService(scope, "not_found")
	is.False(ok)
	is.Equal(DescriptionService{}, output)
}

/////////////////////////////////////////////////////////////////////////////
// 							Describe scopes
/////////////////////////////////////////////////////////////////////////////

func TestDescriptionInjector_String(t *testing.T) {
	// @TODO
}

func TestDescriptionInjectorScope_String(t *testing.T) {
	// @TODO
}

func TestDescriptionInjectorService_String(t *testing.T) {
	is := assert.New(t)

	svc := DescriptionInjectorService{ServiceName: "service-name", ServiceType: ServiceTypeLazy, IsHealthchecker: true, IsShutdowner: true}
	expected := ` * 😴 service-name 🫀 🙅`
	is.Equal(expected, svc.String())

	svc = DescriptionInjectorService{ServiceName: "service-name", ServiceType: ServiceTypeEager, IsHealthchecker: true, IsShutdowner: false}
	expected = ` * 🔁 service-name 🫀`
	is.Equal(expected, svc.String())

	svc = DescriptionInjectorService{ServiceName: "service-name", IsHealthchecker: true, IsShutdowner: true}
	expected = ` * ❓ service-name`
	is.Equal(expected, svc.String())
}

func TestDescribeInjector(t *testing.T) {
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
	output := DescribeInjector(i)
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
	output = DescribeInjector(scope0)
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
	output = DescribeInjector(scope1a)
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
	output = DescribeInjector(scope1b)
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
	output = DescribeInjector(scope2a)
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
	output = DescribeInjector(scope2b)
	is.Equal(expected, output.String())
}
