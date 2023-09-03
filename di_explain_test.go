package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExplainService(t *testing.T) {
	is := assert.New(t)

	rootScope := New()
	child := rootScope.Scope("child")

	scopeIDRoot := rootScope.ID()
	scopeIDChild := child.ID()

	Provide(rootScope, func(i Injector) (*eagerTest, error) {
		return &eagerTest{foobar: "foobar"}, nil
	})
	Provide(child, func(i Injector) (*lazyTest, error) {
		_, err := Invoke[*eagerTest](i)
		if err != nil {
			return nil, err
		}
		return &lazyTest{foobar: "foobar"}, nil
	})
	_ = MustInvoke[*lazyTest](child)

	// from root POV
	a, b, ok := ExplainService[*eagerTest](rootScope)
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{{scopeIDChild, "child", "*github.com/samber/do.lazyTest"}}, b)
	is.True(ok)
	a, b, ok = ExplainService[*lazyTest](rootScope)
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{}, b)
	is.False(ok)
	a, b, ok = ExplainService[int](rootScope)
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{}, b)
	is.False(ok)

	// from child POV
	a, b, ok = ExplainService[*eagerTest](child)
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{{scopeIDChild, "child", "*github.com/samber/do.lazyTest"}}, b)
	is.True(ok)
	a, b, ok = ExplainService[*lazyTest](child)
	is.ElementsMatch([]EdgeService{{scopeIDRoot, "[root]", "*github.com/samber/do.eagerTest"}}, a)
	is.ElementsMatch([]EdgeService{}, b)
	is.True(ok)
	a, b, ok = ExplainService[int](child)
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{}, b)
	is.False(ok)
}

func TestExplainNamedService(t *testing.T) {
	is := assert.New(t)

	rootScope := New()
	child := rootScope.Scope("child")

	scopeIDRoot := rootScope.ID()
	scopeIDChild := child.ID()

	ProvideNamed(rootScope, "eager", func(i Injector) (*eagerTest, error) {
		return &eagerTest{foobar: "foobar"}, nil
	})
	ProvideNamed(child, "lazy", func(i Injector) (*lazyTest, error) {
		_, err := InvokeNamed[*eagerTest](i, "eager")
		if err != nil {
			return nil, err
		}
		return &lazyTest{foobar: "foobar"}, nil
	})
	_ = MustInvokeNamed[*lazyTest](child, "lazy")

	// from root POV
	a, b, ok := ExplainNamedService(rootScope, "eager")
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{{scopeIDChild, "child", "lazy"}}, b)
	is.True(ok)
	a, b, ok = ExplainNamedService(rootScope, "lazy")
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{}, b)
	is.False(ok)
	a, b, ok = ExplainNamedService(rootScope, "foobar")
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{}, b)
	is.False(ok)

	// from child POV
	a, b, ok = ExplainNamedService(child, "eager")
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{{scopeIDChild, "child", "lazy"}}, b)
	is.True(ok)
	a, b, ok = ExplainNamedService(child, "lazy")
	is.ElementsMatch([]EdgeService{{scopeIDRoot, "[root]", "eager"}}, a)
	is.ElementsMatch([]EdgeService{}, b)
	is.True(ok)
	a, b, ok = ExplainNamedService(child, "foobar")
	is.ElementsMatch([]EdgeService{}, a)
	is.ElementsMatch([]EdgeService{}, b)
	is.False(ok)
}
