package do

import (
	"fmt"
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
Service: SERVICE-E
Service type: lazy
Invoked: ` + dirname + `/di_describe_test.go:fakeProvider5:38

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
	output, ok := DescribeNamedService[int](scope, "SERVICE-E")
	is.True(ok)
	is.Equal(expected, output)
	fmt.Println(output)

	// service not found
	output, ok = DescribeNamedService[int](scope, "not_found")
	is.False(ok)
	is.Equal("", output)

	// wrong service type
	output, ok = DescribeNamedService[string](i, "SERVICE-A1")
	is.False(ok)
	is.Equal("", output)
}

/////////////////////////////////////////////////////////////////////////////
// 							Describe scopes
/////////////////////////////////////////////////////////////////////////////

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
	ProvideNamedTransiant[*lazyTest](scope2a, "SERVICE-TRANSIANT-SIMPLE", func(i Injector) (*lazyTest, error) { return &lazyTest{}, nil })
	ProvideNamed[*lazyTestHeathcheckerOK](scope2a, "SERVICE-LAZY-HEALTH", func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })
	ProvideNamed[*lazyTestShutdownerOK](scope2b, "SERVICE-LAZY-SHUTDOWN", func(i Injector) (*lazyTestShutdownerOK, error) { return &lazyTestShutdownerOK{}, nil })
	ProvideNamedValue[int](scope1a, "SERVICE-EAGER-VALUE", 1)
	_, _ = InvokeNamed[int](scope1a, "SERVICE-D")
	_, _ = InvokeNamed[*lazyTestHeathcheckerOK](scope2a, "SERVICE-LAZY-HEALTH")
	_, _ = InvokeNamed[*lazyTestShutdownerOK](scope2b, "SERVICE-LAZY-SHUTDOWN")

	// from root POV
	expected := `
Scope ID: scope-id-root
Scope name: [root]

DAG:

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
         |    |    * 😴 SERVICE-LAZY-HEALTH 🏥
         |    |    * 🏭 SERVICE-TRANSIANT-SIMPLE
         |    |     
         |    |
         |     \_ scope-2b (ID: scope-id-2b)
         |         * 😴 SERVICE-LAZY-SHUTDOWN 🙅
         |          
         |
          \_ scope-1b (ID: scope-id-1b)
              * 😴 SERVICE-F
               
`
	output, ok := DescribeInjector(i)
	is.True(ok)
	is.Equal(expected, output)

	// from scope0 POV
	expected = `
Scope ID: scope-id-0
Scope name: scope-0

DAG:

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
         |    |    * 😴 SERVICE-LAZY-HEALTH 🏥
         |    |    * 🏭 SERVICE-TRANSIANT-SIMPLE
         |    |     
         |    |
         |     \_ scope-2b (ID: scope-id-2b)
         |         * 😴 SERVICE-LAZY-SHUTDOWN 🙅
         |          
         |
          \_ scope-1b (ID: scope-id-1b)
              * 😴 SERVICE-F
               
`
	output, ok = DescribeInjector(scope0)
	is.True(ok)
	is.Equal(expected, output)

	// from scope1a POV
	expected = `
Scope ID: scope-id-1a
Scope name: scope-1a

DAG:

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
              |    * 😴 SERVICE-LAZY-HEALTH 🏥
              |    * 🏭 SERVICE-TRANSIANT-SIMPLE
              |     
              |
               \_ scope-2b (ID: scope-id-2b)
                   * 😴 SERVICE-LAZY-SHUTDOWN 🙅
                    
`
	output, ok = DescribeInjector(scope1a)
	is.True(ok)
	is.Equal(expected, output)

	// from scope1b POV
	expected = `
Scope ID: scope-id-1b
Scope name: scope-1b

DAG:

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
	output, ok = DescribeInjector(scope1b)
	is.True(ok)
	is.Equal(expected, output)

	// from scope2a POV
	expected = `
Scope ID: scope-id-2a
Scope name: scope-2a

DAG:

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
                   * 😴 SERVICE-LAZY-HEALTH 🏥
                   * 🏭 SERVICE-TRANSIANT-SIMPLE
                    
`
	output, ok = DescribeInjector(scope2a)
	is.True(ok)
	is.Equal(expected, output)

	// from scope2b POV
	expected = `
Scope ID: scope-id-2b
Scope name: scope-2b

DAG:

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
	output, ok = DescribeInjector(scope2b)
	is.True(ok)
	is.Equal(expected, output)
}
