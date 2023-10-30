package do

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServiceAlias(t *testing.T) {
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal("foobar1", service1.name)
	is.Equal(i, service1.scope)
	is.Equal("foobar2", service1.targetName)
}

func TestServiceAlias_getName(t *testing.T) {
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal("foobar1", service1.getName())
}

func TestServiceAlias_getType(t *testing.T) {
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal(ServiceTypeAlias, service1.getType())
}

func TestServiceAlias_getInstance(t *testing.T) {
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerOK, Healthchecker](i))

	// basic type
	service1 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("*github.com/samber/do/v2.Healthchecker", i, "*github.com/samber/do/v2.lazyTestHeathcheckerOK")
	instance1, err1 := service1.getInstance(i)
	is.Nil(err1)
	is.EqualValues(&lazyTestHeathcheckerOK{foobar: "foobar"}, instance1)

	// target service not found
	service2 := newServiceAlias[*lazyTestHeathcheckerOK, int]("*github.com/samber/do/v2.Healthchecker", i, "int")
	instance2, err2 := service2.getInstance(i)
	is.EqualError(err2, "DI: could not find service `int`, available services: `*github.com/samber/do/v2.Healthchecker`, `*github.com/samber/do/v2.lazyTestHeathcheckerOK`")
	is.EqualValues(0, instance2)

	Provide(i, func(i Injector) (int, error) {
		return 42, nil
	})

	// target service found but not convertible type
	service3 := newServiceAlias[*lazyTestHeathcheckerOK, int]("*github.com/samber/do/v2.Healthchecker", i, "int")
	instance3, err3 := service3.getInstance(i)
	is.EqualError(err3, "DI: could not find service `int`, available services: `*github.com/samber/do/v2.Healthchecker`, `*github.com/samber/do/v2.lazyTestHeathcheckerOK`, `int`")
	is.EqualValues(0, instance3)

	// @TODO: missing test with child scopes
	// @TODO: missing test with stacktrace
}

func TestServiceAlias_isHealthchecker(t *testing.T) {
	is := assert.New(t)

	// no healthcheck
	i1 := New()
	Provide(i1, func(i Injector) (*lazyTest, error) {
		return &lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTest, any](i1))
	service1, _ := i1.serviceGet("interface {}")
	is.False(service1.(Service[any]).isHealthchecker())

	// healthcheck ok
	i2 := New()
	Provide(i2, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerOK, Healthchecker](i2))
	service2, _ := i2.serviceGet("*github.com/samber/do/v2.Healthchecker")
	is.False(service2.(Service[Healthchecker]).isHealthchecker())
	_, _ = service2.(Service[Healthchecker]).getInstance(nil)
	is.True(service2.(Service[Healthchecker]).isHealthchecker())

	// healthcheck ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerKO, Healthchecker](i3))
	service3, _ := i3.serviceGet("*github.com/samber/do/v2.Healthchecker")
	is.False(service3.(Service[Healthchecker]).isHealthchecker())
	_, _ = service3.(Service[Healthchecker]).getInstance(nil)
	is.True(service3.(Service[Healthchecker]).isHealthchecker())
}

// @TODO: missing tests for context
func TestServiceAlias_healthcheck(t *testing.T) {
	is := assert.New(t)

	ctx := context.Background()

	// no healthcheck
	i1 := New()
	Provide(i1, func(i Injector) (*lazyTest, error) {
		return &lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTest, any](i1))
	service1, _ := i1.serviceGet("interface {}")
	is.Nil(service1.(Service[any]).healthcheck(ctx))

	// healthcheck ok
	i2 := New()
	Provide(i2, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerOK, Healthchecker](i2))
	service2, _ := i2.serviceGet("*github.com/samber/do/v2.Healthchecker")
	is.Nil(service2.(Service[Healthchecker]).healthcheck(ctx))
	_, _ = service2.(Service[Healthchecker]).getInstance(nil)
	is.Nil(service2.(Service[Healthchecker]).healthcheck(ctx))

	// healthcheck ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerKO, Healthchecker](i3))
	service3, _ := i3.serviceGet("*github.com/samber/do/v2.Healthchecker")
	is.Nil(service3.(Service[Healthchecker]).healthcheck(ctx))
	_, _ = service3.(Service[Healthchecker]).getInstance(nil)
	is.Equal(assert.AnError, service3.(Service[Healthchecker]).healthcheck(ctx))
}

// @TODO: missing tests for context
func TestServiceAlias_isShutdowner(t *testing.T) {
	is := assert.New(t)

	// no shutdown
	i1 := New()
	Provide(i1, func(i Injector) (*lazyTest, error) {
		return &lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTest, any](i1))
	service1, _ := i1.serviceGet("interface {}")
	is.False(service1.(Service[any]).isShutdowner())

	// shutdown ok
	i2 := New()
	Provide(i2, func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestShutdownerOK, ShutdownerWithContextAndError](i2))
	service2, _ := i2.serviceGet("*github.com/samber/do/v2.ShutdownerWithContextAndError")
	is.False(service2.(Service[ShutdownerWithContextAndError]).isShutdowner())
	_, _ = service2.(Service[ShutdownerWithContextAndError]).getInstance(nil)
	is.True(service2.(Service[ShutdownerWithContextAndError]).isShutdowner())

	// shutdown ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestShutdownerKO, ShutdownerWithError](i3))
	service3, _ := i3.serviceGet("*github.com/samber/do/v2.ShutdownerWithError")
	is.False(service3.(Service[ShutdownerWithError]).isShutdowner())
	_, _ = service3.(Service[ShutdownerWithError]).getInstance(nil)
	is.True(service3.(Service[ShutdownerWithError]).isShutdowner())
}

func TestServiceAlias_shutdown(t *testing.T) {
	is := assert.New(t)

	ctx := context.Background()

	// no shutdown
	i1 := New()
	Provide(i1, func(i Injector) (*lazyTest, error) {
		return &lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTest, any](i1))
	service1, _ := i1.serviceGet("interface {}")
	is.Nil(service1.(Service[any]).shutdown(ctx))

	// shutdown ok
	i2 := New()
	Provide(i2, func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestShutdownerOK, ShutdownerWithContextAndError](i2))
	service2, _ := i2.serviceGet("*github.com/samber/do/v2.ShutdownerWithContextAndError")
	is.Nil(service2.(Service[ShutdownerWithContextAndError]).shutdown(ctx))
	_, _ = service2.(Service[ShutdownerWithContextAndError]).getInstance(nil)
	is.Nil(service2.(Service[ShutdownerWithContextAndError]).shutdown(ctx))

	// shutdown ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestShutdownerKO, ShutdownerWithError](i3))
	service3, _ := i3.serviceGet("*github.com/samber/do/v2.ShutdownerWithError")
	is.Nil(service3.(Service[ShutdownerWithError]).shutdown(ctx))
	_, _ = service3.(Service[ShutdownerWithError]).getInstance(nil)
	is.Equal(assert.AnError, service3.(Service[ShutdownerWithError]).shutdown(ctx))
}

func TestServiceAlias_clone(t *testing.T) {
	// @TODO
}

func TestServiceAlias_source(t *testing.T) {
	// @TODO
}
