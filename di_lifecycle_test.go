package do

import (
	"context"
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
	is := assert.New(t)

	i := New()
	ctx := context.Background()

	Provide(i, func(i Injector) (*lazyTestHeathcheckerKOCtx, error) {
		return &lazyTestHeathcheckerKOCtx{foobar: "foobar"}, nil
	})
	is.Nil(HealthCheckWithContext[*lazyTestHeathcheckerKOCtx](ctx, i))
	_, _ = Invoke[*lazyTestHeathcheckerKOCtx](i)

	is.Error(assert.AnError, HealthCheckWithContext[*lazyTestHeathcheckerKOCtx](ctx, i))
}

func TestHealthCheckNamed(t *testing.T) {
	is := assert.New(t)

	i := New()

	ProvideNamed(i, "foobar", func(i Injector) (*lazyTestHeathcheckerKO, error) { return &lazyTestHeathcheckerKO{}, nil })
	is.Nil(HealthCheckNamed(i, "foobar"))
	_, _ = InvokeNamed[*lazyTestHeathcheckerKO](i, "foobar")

	is.Error(assert.AnError, HealthCheckNamed(i, "foobar"))
}

func TestHealthCheckNamedWithContext(t *testing.T) {
	is := assert.New(t)

	i := New()
	ctx := context.Background()

	ProvideNamed(i, "foobar", func(i Injector) (*lazyTestHeathcheckerKOCtx, error) {
		return &lazyTestHeathcheckerKOCtx{foobar: "foobar"}, nil
	})
	is.Nil(HealthCheckNamedWithContext(ctx, i, "foobar"))
	_, _ = InvokeNamed[*lazyTestHeathcheckerKOCtx](i, "foobar")

	is.Error(assert.AnError, HealthCheckNamedWithContext(ctx, i, "foobar"))
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
	is := assert.New(t)

	i := New()
	ctx := context.Background()

	Provide(i, func(i Injector) (*lazyTestShutdownerKOCtx, error) {
		return &lazyTestShutdownerKOCtx{foobar: "foobar"}, nil
	})
	is.Nil(ShutdownWithContext[*lazyTestShutdownerKOCtx](ctx, i))
	_, _ = Invoke[*lazyTestShutdownerKOCtx](i)

	is.Error(assert.AnError, ShutdownWithContext[*lazyTestShutdownerKOCtx](ctx, i))
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
	is := assert.New(t)

	i := New()
	ctx := context.Background()

	Provide(i, func(i Injector) (*lazyTestShutdownerKOCtx, error) {
		return &lazyTestShutdownerKOCtx{foobar: "foobar"}, nil
	})
	is.NotPanics(func() {
		MustShutdownWithContext[*lazyTestShutdownerKOCtx](ctx, i)
	})
	_, _ = Invoke[*lazyTestShutdownerKOCtx](i)

	is.PanicsWithError("DI: could not find service `*github.com/samber/do/v2.lazyTestShutdownerKOCtx`, no service available", func() {
		MustShutdownWithContext[*lazyTestShutdownerKOCtx](ctx, i)
	})
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
	is := assert.New(t)

	i := New()
	ctx := context.Background()

	ProvideNamed(i, "foobar", func(i Injector) (*lazyTestShutdownerKOCtx, error) {
		return &lazyTestShutdownerKOCtx{foobar: "foobar"}, nil
	})
	is.Nil(ShutdownNamedWithContext(ctx, i, "foobar"))
	_, _ = Invoke[*lazyTestShutdownerKOCtx](i)

	is.EqualError(ShutdownNamedWithContext(ctx, i, "foobar"), "DI: could not find service `foobar`, no service available")
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
	is := assert.New(t)

	i := New()
	ctx := context.Background()

	ProvideNamed(i, "foobar", func(i Injector) (*lazyTestShutdownerKOCtx, error) {
		return &lazyTestShutdownerKOCtx{foobar: "foobar"}, nil
	})
	is.NotPanics(func() {
		MustShutdownNamedWithContext(ctx, i, "foobar")
	})
	_, _ = InvokeNamed[*lazyTestShutdownerKOCtx](i, "foobar")

	is.PanicsWithError("DI: could not find service `foobar`, no service available", func() {
		MustShutdownNamedWithContext(ctx, i, "foobar")
	})
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

	is.PanicsWithError("DI: service `*github.com/samber/do/v2.test` has already been declared", func() {
		Provide(i, func(i Injector) (*test, error) {
			return &test{}, nil
		})
	})

	is.PanicsWithError("DI: service `*github.com/samber/do/v2.test` has already been declared", func() {
		ProvideValue(i, &test{})
	})

	is.PanicsWithError("DI: service `*github.com/samber/do/v2.test` has already been declared", func() {
		ProvideNamed(i, "*github.com/samber/do/v2.test", func(i Injector) (*test, error) {
			return &test{}, nil
		})
	})

	is.PanicsWithError("DI: service `*github.com/samber/do/v2.test` has already been declared", func() {
		ProvideNamedValue(i, "*github.com/samber/do/v2.test", &test{})
	})
}
