package do

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServiceAlias(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal("foobar1", service1.name)
	is.Equal(i, service1.scope)
	is.Equal("foobar2", service1.targetName)
}

func TestServiceAlias_getName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal("foobar1", service1.getName())
}

func TestServiceAlias_getTypeName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, int]("foobar1", i, "foobar2")
	is.Equal("int", service1.getTypeName())
}

func TestServiceAlias_getServiceType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal(ServiceTypeAlias, service1.getServiceType())
}

func TestServiceAlias_getReflectType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	service1 := newServiceAlias[string, int]("foobar1", nil, "foobar2")
	is.Equal("int", service1.getReflectType().String())

	service2 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("foobar2", nil, "foobar3")
	is.Equal(pkgName+".Healthchecker", service2.getReflectType().String())

	service3 := newServiceAlias[iTestHeathchecker, Healthchecker]("foobar3", nil, "foobar4")
	is.Equal(pkgName+".Healthchecker", service3.getReflectType().String())
}

func TestServiceAlias_getInstanceAny(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerOK, Healthchecker](i))

	// basic type
	service1 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("github.com/samber/do/v2.Healthchecker", i, "*github.com/samber/do/v2.lazyTestHeathcheckerOK")
	instance1, err1 := service1.getInstanceAny(i)
	is.Nil(err1)
	is.EqualValues(&lazyTestHeathcheckerOK{foobar: "foobar"}, instance1)

	// target service not found
	service2 := newServiceAlias[*lazyTestHeathcheckerOK, int]("github.com/samber/do/v2.Healthchecker", i, "int")
	instance2, err2 := service2.getInstanceAny(i)
	is.EqualError(err2, "DI: could not find service `int`, available services: `*github.com/samber/do/v2.lazyTestHeathcheckerOK`, `github.com/samber/do/v2.Healthchecker`")
	is.EqualValues(0, instance2)

	Provide(i, func(i Injector) (int, error) {
		return 42, nil
	})

	// target service found but not convertible type
	service3 := newServiceAlias[*lazyTestHeathcheckerOK, int]("github.com/samber/do/v2.Healthchecker", i, "int")
	instance3, err3 := service3.getInstanceAny(i)
	is.EqualError(err3, "DI: service found, but type mismatch: invoking `*github.com/samber/do/v2.lazyTestHeathcheckerOK` but registered `int`")
	is.EqualValues(0, instance3)

	// @TODO: missing test with child scopes
	// @TODO: missing test with stacktrace
}

func TestServiceAlias_getInstance(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	i := New()
	Provide(i, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerOK, Healthchecker](i))

	// basic type
	service1 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("github.com/samber/do/v2.Healthchecker", i, "*github.com/samber/do/v2.lazyTestHeathcheckerOK")
	instance1, err1 := service1.getInstance(i)
	is.Nil(err1)
	is.EqualValues(&lazyTestHeathcheckerOK{foobar: "foobar"}, instance1)

	// target service not found
	service2 := newServiceAlias[*lazyTestHeathcheckerOK, int]("github.com/samber/do/v2.Healthchecker", i, "int")
	instance2, err2 := service2.getInstance(i)
	is.EqualError(err2, "DI: could not find service `int`, available services: `*github.com/samber/do/v2.lazyTestHeathcheckerOK`, `github.com/samber/do/v2.Healthchecker`")
	is.EqualValues(0, instance2)

	Provide(i, func(i Injector) (int, error) {
		return 42, nil
	})

	// target service found but not convertible type
	service3 := newServiceAlias[*lazyTestHeathcheckerOK, int]("github.com/samber/do/v2.Healthchecker", i, "int")
	instance3, err3 := service3.getInstance(i)
	is.EqualError(err3, "DI: service found, but type mismatch: invoking `*github.com/samber/do/v2.lazyTestHeathcheckerOK` but registered `int`")
	is.EqualValues(0, instance3)

	// @TODO: missing test with child scopes
	// @TODO: missing test with stacktrace
}

func TestServiceAlias_isHealthchecker(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// no healthcheck
	i1 := New()
	Provide(i1, func(i Injector) (*lazyTest, error) {
		return &lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTest, any](i1))
	service1, _ := i1.serviceGet("interface {}")
	is.False(service1.(serviceWrapper[any]).isHealthchecker())

	// healthcheck ok
	i2 := New()
	Provide(i2, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerOK, Healthchecker](i2))
	service2, _ := i2.serviceGet("github.com/samber/do/v2.Healthchecker")
	is.False(service2.(serviceWrapperIsHealthchecker).isHealthchecker())
	_, _ = service2.(serviceWrapperGetInstanceAny).getInstanceAny(nil)
	is.True(service2.(serviceWrapperIsHealthchecker).isHealthchecker())

	// healthcheck ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerKO, Healthchecker](i3))
	service3, _ := i3.serviceGet("github.com/samber/do/v2.Healthchecker")
	is.False(service3.(serviceWrapperIsHealthchecker).isHealthchecker())
	_, _ = service3.(serviceWrapperGetInstanceAny).getInstanceAny(nil)
	is.True(service3.(serviceWrapperIsHealthchecker).isHealthchecker())

	// service not found (wrong type)
	i4 := New()
	service4 := newServiceAlias[*lazyTestHeathcheckerKO, Healthchecker]("github.com/samber/do/v2.Healthchecker", i4, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.False(service4.isHealthchecker())
	_, _ = service4.getInstanceAny(nil)
	is.False(service4.isHealthchecker())

	// service not found (wrong name)
	i5 := New()
	service5 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("github.com/samber/do/v2.Healthchecker", i5, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.False(service5.isHealthchecker())
	_, _ = service5.getInstanceAny(nil)
	is.False(service5.isHealthchecker())
}

// @TODO: missing tests for context
func TestServiceAlias_healthcheck(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ctx := context.Background()

	// no healthcheck
	i1 := New()
	Provide(i1, func(i Injector) (*lazyTest, error) {
		return &lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTest, any](i1))
	service1, _ := i1.serviceGet("interface {}")
	is.Nil(service1.(serviceWrapper[any]).healthcheck(ctx))

	// healthcheck ok
	i2 := New()
	Provide(i2, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerOK, Healthchecker](i2))
	service2, _ := i2.serviceGet("github.com/samber/do/v2.Healthchecker")
	is.Nil(service2.(serviceWrapper[Healthchecker]).healthcheck(ctx))
	_, _ = service2.(serviceWrapper[Healthchecker]).getInstance(nil)
	is.Nil(service2.(serviceWrapper[Healthchecker]).healthcheck(ctx))

	// healthcheck ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerKO, Healthchecker](i3))
	service3, _ := i3.serviceGet("github.com/samber/do/v2.Healthchecker")
	is.Nil(service3.(serviceWrapper[Healthchecker]).healthcheck(ctx))
	_, _ = service3.(serviceWrapper[Healthchecker]).getInstance(nil)
	is.Equal(assert.AnError, service3.(serviceWrapper[Healthchecker]).healthcheck(ctx))

	// service not found (wrong type)
	i4 := New()
	service4 := newServiceAlias[*lazyTestHeathcheckerKO, Healthchecker]("github.com/samber/do/v2.Healthchecker", i4, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.Nil(service4.healthcheck(ctx))
	_, _ = service4.getInstanceAny(nil)
	is.Nil(service4.healthcheck(ctx))

	// service not found (wrong name)
	i5 := New()
	service5 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("github.com/samber/do/v2.Healthchecker", i5, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.Nil(service5.healthcheck(ctx))
	_, _ = service5.getInstanceAny(nil)
	is.Nil(service5.healthcheck(ctx))
}

// @TODO: missing tests for context
func TestServiceAlias_isShutdowner(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// no shutdown
	i1 := New()
	Provide(i1, func(i Injector) (*lazyTest, error) {
		return &lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTest, any](i1))
	service1, _ := i1.serviceGet("interface {}")
	is.False(service1.(serviceWrapper[any]).isShutdowner())

	// shutdown ok
	i2 := New()
	Provide(i2, func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestShutdownerOK, ShutdownerWithContextAndError](i2))
	service2, _ := i2.serviceGet("github.com/samber/do/v2.ShutdownerWithContextAndError")
	is.False(service2.(serviceWrapper[ShutdownerWithContextAndError]).isShutdowner())
	_, _ = service2.(serviceWrapper[ShutdownerWithContextAndError]).getInstance(nil)
	is.True(service2.(serviceWrapper[ShutdownerWithContextAndError]).isShutdowner())

	// shutdown ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestShutdownerKO, ShutdownerWithError](i3))
	service3, _ := i3.serviceGet("github.com/samber/do/v2.ShutdownerWithError")
	is.False(service3.(serviceWrapper[ShutdownerWithError]).isShutdowner())
	_, _ = service3.(serviceWrapper[ShutdownerWithError]).getInstance(nil)
	is.True(service3.(serviceWrapper[ShutdownerWithError]).isShutdowner())

	// service not found (wrong type)
	i4 := New()
	service4 := newServiceAlias[*lazyTestShutdownerKO, Healthchecker]("*github.com/samber/do/v2.Healthchecker", i4, "*github.com/samber/do/v2.lazyTestShutdownerKO")
	is.False(service4.isShutdowner())
	_, _ = service4.getInstanceAny(nil)
	is.False(service4.isShutdowner())

	// service not found (wrong name)
	i5 := New()
	service5 := newServiceAlias[*lazyTestShutdownerOK, Healthchecker]("*github.com/samber/do/v2.Healthchecker", i5, "*github.com/samber/do/v2.lazyTestShutdownerKO")
	is.False(service5.isShutdowner())
	_, _ = service5.getInstanceAny(nil)
	is.False(service5.isShutdowner())
}

func TestServiceAlias_shutdown(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ctx := context.Background()

	// no shutdown
	i1 := New()
	Provide(i1, func(i Injector) (*lazyTest, error) {
		return &lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTest, any](i1))
	service1, _ := i1.serviceGet("interface {}")
	is.Nil(service1.(serviceWrapper[any]).shutdown(ctx))

	// shutdown ok
	i2 := New()
	Provide(i2, func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestShutdownerOK, ShutdownerWithContextAndError](i2))
	service2, _ := i2.serviceGet("github.com/samber/do/v2.ShutdownerWithContextAndError")
	is.Nil(service2.(serviceWrapper[ShutdownerWithContextAndError]).shutdown(ctx))
	_, _ = service2.(serviceWrapper[ShutdownerWithContextAndError]).getInstance(nil)
	is.Nil(service2.(serviceWrapper[ShutdownerWithContextAndError]).shutdown(ctx))

	// shutdown ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestShutdownerKO, ShutdownerWithError](i3))
	service3, _ := i3.serviceGet("github.com/samber/do/v2.ShutdownerWithError")
	is.Nil(service3.(serviceWrapper[ShutdownerWithError]).shutdown(ctx))
	_, _ = service3.(serviceWrapper[ShutdownerWithError]).getInstance(nil)
	is.Equal(assert.AnError, service3.(serviceWrapper[ShutdownerWithError]).shutdown(ctx))

	// service not found (wrong type)
	i4 := New()
	service4 := newServiceAlias[*lazyTestShutdownerKO, Healthchecker]("github.com/samber/do/v2.Healthchecker", i4, "*github.com/samber/do/v2.lazyTestShutdownerKO")
	is.Nil(service4.shutdown(ctx))
	_, _ = service4.getInstanceAny(nil)
	is.Nil(service4.shutdown(ctx))

	// service not found (wrong name)
	i5 := New()
	service5 := newServiceAlias[*lazyTestShutdownerOK, Healthchecker]("github.com/samber/do/v2.Healthchecker", i5, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.Nil(service5.shutdown(ctx))
	_, _ = service5.getInstanceAny(nil)
	is.Nil(service5.shutdown(ctx))
}

func TestServiceAlias_clone(t *testing.T) {
	// @TODO
}

func TestServiceAlias_source(t *testing.T) {
	// @TODO
}
