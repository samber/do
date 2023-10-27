package do

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	is := assert.New(t)

	i := New()

	Provide(i, func(i Injector) (*lazyTest, error) { return &lazyTest{}, nil })
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) { return &lazyTestHeathcheckerOK{}, nil })
	Provide(i, func(i Injector) (*lazyTestHeathcheckerKO, error) { return &lazyTestHeathcheckerKO{}, nil })

	is.Nil(HealthCheck[*lazyTest](i))
	is.Nil(HealthCheck[*lazyTestHeathcheckerOK](i))
	is.Nil(HealthCheck[*lazyTestHeathcheckerKO](i))

	_, _ = Invoke[*lazyTest](i)
	_, _ = Invoke[*lazyTestHeathcheckerOK](i)
	_, _ = Invoke[*lazyTestHeathcheckerKO](i)

	is.Nil(HealthCheck[*lazyTest](i))
	is.Nil(HealthCheck[*lazyTestHeathcheckerOK](i))
	is.Error(assert.AnError, HealthCheck[*lazyTestHeathcheckerKO](i))
}

func TestHealthCheckWithContext(t *testing.T) {
	// @TODO
}

func TestHealthCheckNamed(t *testing.T) {
	// @TODO
}

func TestHealthCheckNamedWithContext(t *testing.T) {
	// @TODO
}

func TestShutdown(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}

	i := New()

	Provide(i, func(i Injector) (test, error) {
		return test{foobar: "foobar"}, nil
	})

	instance, err := Invoke[test](i)
	is.Equal(test{foobar: "foobar"}, instance)
	is.Nil(err)

	err = Shutdown[test](i)
	is.Nil(err)

	instance, err = Invoke[test](i)
	is.Empty(instance)
	is.NotNil(err)

	err = Shutdown[test](i)
	is.NotNil(err)
}

func TestShutdownWithContext(t *testing.T) {
	// @TODO
}

func TestMustShutdown(t *testing.T) {
	is := assert.New(t)

	type test struct {
		foobar string
	}

	i := New()

	Provide(i, func(i Injector) (test, error) {
		return test{foobar: "foobar"}, nil
	})

	instance, err := Invoke[test](i)
	is.Equal(test{foobar: "foobar"}, instance)
	is.Nil(err)

	is.NotPanics(func() {
		MustShutdown[test](i)
	})

	instance, err = Invoke[test](i)
	is.Empty(instance)
	is.NotNil(err)

	is.Panics(func() {
		MustShutdown[test](i)
	})
}

func TestMustShutdownWithContext(t *testing.T) {
	// @TODO
}

func TestShutdownNamed(t *testing.T) {
	is := assert.New(t)

	i := New()

	ProvideNamedValue(i, "foobar", 42)

	instance, err := InvokeNamed[int](i, "foobar")
	is.Equal(42, instance)
	is.Nil(err)

	err = ShutdownNamed(i, "foobar")
	is.Nil(err)

	instance, err = InvokeNamed[int](i, "foobar")
	is.Empty(instance)
	is.NotNil(err)

	err = ShutdownNamed(i, "foobar")
	is.NotNil(err)
}

func TestShutdownNamedWithContext(t *testing.T) {
	// @TODO
}

func TestMustShutdownNamed(t *testing.T) {
	is := assert.New(t)

	i := New()

	ProvideNamedValue(i, "foobar", 42)

	instance, err := InvokeNamed[int](i, "foobar")
	is.Equal(42, instance)
	is.Nil(err)

	is.NotPanics(func() {
		MustShutdownNamed(i, "foobar")
	})

	instance, err = InvokeNamed[int](i, "foobar")
	is.Empty(instance)
	is.NotNil(err)

	is.Panics(func() {
		MustShutdownNamed(i, "foobar")
	})
}

func TestMustShutdownNamedWithContext(t *testing.T) {
	// @TODO
}

func TestDoubleInjection(t *testing.T) {
	is := assert.New(t)

	type test struct{}

	i := New()

	is.NotPanics(func() {
		Provide(i, func(i Injector) (*test, error) {
			return &test{}, nil
		})
	})

	is.PanicsWithError("DI: service `*github.com/samber/do.test` has already been declared", func() {
		Provide(i, func(i Injector) (*test, error) {
			return &test{}, nil
		})
	})

	is.PanicsWithError("DI: service `*github.com/samber/do.test` has already been declared", func() {
		ProvideValue(i, &test{})
	})

	is.PanicsWithError("DI: service `*github.com/samber/do.test` has already been declared", func() {
		ProvideNamed(i, "*github.com/samber/do.test", func(i Injector) (*test, error) {
			return &test{}, nil
		})
	})

	is.PanicsWithError("DI: service `*github.com/samber/do.test` has already been declared", func() {
		ProvideNamedValue(i, "*github.com/samber/do.test", &test{})
	})
}
