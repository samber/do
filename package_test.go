package do

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackage(t *testing.T) {
	is := assert.New(t)

	type test struct{}
	type iTest interface{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	pkg := Package(
		Lazy(provider1),
		Eager(test{}),
		Bind[*test, iTest](),
	)

	root := New()
	pkg(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())
	svc3 := newEdgeService(root.ID(), root.Name(), NameOf[iTest]())

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_ = MustInvoke[test](root)
		_ = MustInvoke[iTest](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListInvokedServices())
}

func TestNewWithPackage(t *testing.T) {
	is := assert.New(t)

	type test struct{}
	type iTest interface{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	pkg := Package(
		Lazy(provider1),
		Eager(test{}),
	)

	root := New(
		pkg,
		Bind[*test, iTest](),
	)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())
	svc3 := newEdgeService(root.ID(), root.Name(), NameOf[iTest]())

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_ = MustInvoke[test](root)
		_ = MustInvoke[iTest](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListInvokedServices())
}

func TestLazy(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	Lazy(provider1)(root)
	Lazy(provider2)(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_, _ = Invoke[test](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1}, root.ListInvokedServices())
}

func TestLazyNamed(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	LazyNamed("p1", provider1)(root)
	LazyNamed("p2", provider2)(root)

	svc1 := newEdgeService(root.ID(), root.Name(), "p1")
	svc2 := newEdgeService(root.ID(), root.Name(), "p2")

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvokeNamed[*test](root, "p1")
		_, _ = InvokeNamed[test](root, "p2")
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1}, root.ListInvokedServices())
}

func TestEager(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	root := New()
	Eager(&test{})(root)
	Eager(test{})(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_ = MustInvoke[test](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListInvokedServices())
}

func TestEagerNamed(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	root := New()
	EagerNamed("p1", &test{})(root)
	EagerNamed("p2", test{})(root)

	svc1 := newEdgeService(root.ID(), root.Name(), "p1")
	svc2 := newEdgeService(root.ID(), root.Name(), "p2")

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvokeNamed[*test](root, "p1")
		_ = MustInvokeNamed[test](root, "p2")
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListInvokedServices())
}

func TestTransient(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	Transient(provider1)(root)
	Transient(provider2)(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_, _ = Invoke[test](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1}, root.ListInvokedServices())
}

func TestTransientNamed(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	TransientNamed("p1", provider1)(root)
	TransientNamed("p2", provider2)(root)

	svc1 := newEdgeService(root.ID(), root.Name(), "p1")
	svc2 := newEdgeService(root.ID(), root.Name(), "p2")

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvokeNamed[*test](root, "p1")
		_, _ = InvokeNamed[test](root, "p2")
	})

	is.ElementsMatch([]EdgeService{svc1, svc2}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1}, root.ListInvokedServices())
}

func TestBind(t *testing.T) {
	is := assert.New(t)

	type test struct{}
	type iTest interface{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	Lazy(provider1)(root)
	Lazy(provider2)(root)
	Bind[*test, iTest]()(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())
	svc3 := newEdgeService(root.ID(), root.Name(), NameOf[iTest]())

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_, _ = Invoke[test](root)
		_ = MustInvoke[iTest](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc3}, root.ListInvokedServices())
}

func TestBindNamed(t *testing.T) {
	is := assert.New(t)

	type test struct{}
	type iTest interface{}

	provider1 := func(i Injector) (*test, error) {
		return &test{}, nil
	}

	provider2 := func(i Injector) (test, error) {
		return test{}, fmt.Errorf("error")
	}

	root := New()
	Lazy(provider1)(root)
	Lazy(provider2)(root)
	BindNamed[*test, iTest](NameOf[*test](), NameOf[iTest]())(root)

	svc1 := newEdgeService(root.ID(), root.Name(), NameOf[*test]())
	svc2 := newEdgeService(root.ID(), root.Name(), NameOf[test]())
	svc3 := newEdgeService(root.ID(), root.Name(), NameOf[iTest]())

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{}, root.ListInvokedServices())

	is.NotPanics(func() {
		_ = MustInvoke[*test](root)
		_, _ = Invoke[test](root)
		_ = MustInvoke[iTest](root)
	})

	is.ElementsMatch([]EdgeService{svc1, svc2, svc3}, root.ListProvidedServices())
	is.ElementsMatch([]EdgeService{svc1, svc3}, root.ListInvokedServices())
}
