package do

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type lazyTest struct {
	foobar string
}

// Alias to HealthChecker.
// Used for testing interface implements interface.
type iTestHeathchecker interface {
	HealthCheck() error
}

var _ Healthchecker = (*lazyTestHeathcheckerOK)(nil)

type lazyTestHeathcheckerOK struct {
	foobar string
}

func (t *lazyTestHeathcheckerOK) HealthCheck() error {
	return nil
}

var _ Healthchecker = (*lazyTestHeathcheckerKO)(nil)

type lazyTestHeathcheckerKO struct {
	foobar string
}

func (t *lazyTestHeathcheckerKO) HealthCheck() error {
	return assert.AnError
}

var _ HealthcheckerWithContext = (*lazyTestHeathcheckerKOCtx)(nil)

type lazyTestHeathcheckerKOCtx struct {
	foobar string
}

func (t *lazyTestHeathcheckerKOCtx) HealthCheck(ctx context.Context) error {
	return assert.AnError
}

var _ Healthchecker = (*lazyTestHeathcheckerOKTimeout)(nil)

type lazyTestHeathcheckerOKTimeout struct {
	foobar string
}

func (t *lazyTestHeathcheckerOKTimeout) HealthCheck() error {
	time.Sleep(20 * time.Millisecond)
	return nil
}

var _ ShutdownerWithContextAndError = (*lazyTestShutdownerOK)(nil)

type lazyTestShutdownerOK struct {
	foobar string
}

func (t *lazyTestShutdownerOK) Shutdown(ctx context.Context) error {
	return nil
}

var _ ShutdownerWithError = (*lazyTestShutdownerKO)(nil)

type lazyTestShutdownerKO struct {
	foobar string
}

func (t *lazyTestShutdownerKO) Shutdown() error {
	return assert.AnError
}

var _ ShutdownerWithContextAndError = (*lazyTestShutdownerKOCtx)(nil)

type lazyTestShutdownerKOCtx struct {
	foobar string
}

func (t *lazyTestShutdownerKOCtx) Shutdown(ctx context.Context) error {
	return assert.AnError
}

func TestNewServiceLazy(t *testing.T) {
	// @TODO
}

func TestServiceLazy_getName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := lazyTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (lazyTest, error) {
		return test, nil
	}

	service1 := newServiceLazy("foobar1", provider1)
	is.Equal("foobar1", service1.getName())

	service2 := newServiceLazy("foobar2", provider2)
	is.Equal("foobar2", service2.getName())
}

func TestServiceLazy_getTypeName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := lazyTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (lazyTest, error) {
		return test, nil
	}

	service1 := newServiceLazy("foobar1", provider1)
	is.Equal("int", service1.getTypeName())

	service2 := newServiceLazy("foobar2", provider2)
	is.Equal("github.com/samber/do/v2.lazyTest", service2.getTypeName())
}

func TestServiceLazy_getServiceType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := lazyTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (lazyTest, error) {
		return test, nil
	}

	service1 := newServiceLazy("foobar1", provider1)
	is.Equal(ServiceTypeLazy, service1.getServiceType())

	service2 := newServiceLazy("foobar2", provider2)
	is.Equal(ServiceTypeLazy, service2.getServiceType())
}

func TestServiceLazy_getReflectType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := lazyTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (lazyTest, error) {
		return test, nil
	}
	provider3 := func(i Injector) (Healthchecker, error) {
		return nil, nil
	}
	provider4 := func(i Injector) (*lazyTest, error) {
		return &test, nil
	}

	service1 := newServiceLazy("foobar1", provider1)
	is.Equal("int", service1.getReflectType().String())

	service2 := newServiceLazy("foobar2", provider2)
	is.Equal(pkgName+".lazyTest", service2.getReflectType().String())

	service3 := newServiceLazy("foobar3", provider3)
	is.Equal(pkgName+".Healthchecker", service3.getReflectType().String())

	service4 := newServiceLazy("foobar1", provider4)
	is.Equal("*"+pkgName+".lazyTest", service4.getReflectType().String())
}

func TestServiceLazy_getInstanceAny(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := lazyTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (lazyTest, error) {
		return test, nil
	}
	provider3 := func(i Injector) (int, error) {
		panic("error")
	}
	provider4 := func(i Injector) (int, error) {
		panic(fmt.Errorf("error"))
	}
	provider5 := func(i Injector) (int, error) {
		return 42, fmt.Errorf("error")
	}

	i := New()

	// basic type
	service1 := newServiceLazy("foobar", provider1)
	instance1, err1 := service1.getInstanceAny(i)
	is.Nil(err1)
	is.Equal(42, instance1)

	// struct
	service2 := newServiceLazy("hello", provider2)
	instance2, err2 := service2.getInstanceAny(i)
	is.Nil(err2)
	is.Equal(test, instance2)

	// provider panics, but panic is catched by getInstanceAny
	is.NotPanics(func() {
		service3 := newServiceLazy("baz", provider3)
		_, _ = service3.getInstanceAny(i)
	})

	// provider panics, but panic is catched by getInstanceAny
	is.NotPanics(func() {
		service4 := newServiceLazy("plop", provider4)
		instance4, err4 := service4.getInstanceAny(i)
		is.NotNil(err4)
		is.Empty(instance4)
		expected := fmt.Errorf("error")
		is.Equal(expected, err4)
	})

	// provider returning error
	is.NotPanics(func() {
		service5 := newServiceLazy("plop", provider5)
		instance5, err5 := service5.getInstanceAny(i)
		is.NotNil(err5)
		is.Empty(instance5)
		expected := fmt.Errorf("error")
		is.Equal(expected, err5)
	})
}

func TestServiceLazy_getInstance(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := lazyTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (lazyTest, error) {
		return test, nil
	}
	provider3 := func(i Injector) (int, error) {
		panic("error")
	}
	provider4 := func(i Injector) (int, error) {
		panic(fmt.Errorf("error"))
	}
	provider5 := func(i Injector) (int, error) {
		return 42, fmt.Errorf("error")
	}

	i := New()

	// basic type
	service1 := newServiceLazy("foobar", provider1)
	instance1, err1 := service1.getInstance(i)
	is.Nil(err1)
	is.Equal(42, instance1)

	// struct
	service2 := newServiceLazy("hello", provider2)
	instance2, err2 := service2.getInstance(i)
	is.Nil(err2)
	is.Equal(test, instance2)

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service3 := newServiceLazy("baz", provider3)
		_, _ = service3.getInstance(i)
	})

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service4 := newServiceLazy("plop", provider4)
		instance4, err4 := service4.getInstance(i)
		is.NotNil(err4)
		is.Empty(instance4)
		expected := fmt.Errorf("error")
		is.Equal(expected, err4)
	})

	// provider returning error
	is.NotPanics(func() {
		service5 := newServiceLazy("plop", provider5)
		instance5, err5 := service5.getInstance(i)
		is.NotNil(err5)
		is.Empty(instance5)
		expected := fmt.Errorf("error")
		is.Equal(expected, err5)
	})
}

func TestServiceLazy_build(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := lazyTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (lazyTest, error) {
		return test, nil
	}
	provider3 := func(i Injector) (int, error) {
		panic("error")
	}
	provider4 := func(i Injector) (int, error) {
		panic(fmt.Errorf("error"))
	}
	provider5 := func(i Injector) (int, error) {
		return 42, fmt.Errorf("error")
	}

	i := New()

	// basic type
	service1 := newServiceLazy("foobar", provider1)
	is.False(service1.built)
	is.Empty(service1.buildTime)
	err1 := service1.build(i)
	is.True(service1.built)
	is.NotEmpty(service1.buildTime)
	is.Nil(err1)

	// struct
	service2 := newServiceLazy("hello", provider2)
	is.False(service2.built)
	is.Empty(service2.buildTime)
	err2 := service2.build(i)
	is.True(service2.built)
	is.NotEmpty(service2.buildTime)
	is.Nil(err2)

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service3 := newServiceLazy("baz", provider3)
		is.False(service3.built)
		is.Empty(service3.buildTime)
		_ = service3.build(i)
		is.False(service3.built)
		is.Empty(service3.buildTime)
	})

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service4 := newServiceLazy("plop", provider4)
		is.False(service4.built)
		is.Empty(service4.buildTime)
		err4 := service4.build(i)
		is.False(service4.built)
		is.Empty(service4.buildTime)
		is.NotNil(err4)
		is.Equal(fmt.Errorf("error"), err4)
	})

	// provider returning error
	is.NotPanics(func() {
		service5 := newServiceLazy("plop", provider5)
		is.False(service5.built)
		is.Empty(service5.buildTime)
		err5 := service5.build(i)
		is.False(service5.built)
		is.Empty(service5.buildTime)
		is.NotNil(err5)
		is.Equal(fmt.Errorf("error"), err5)
	})
}

func TestServiceLazy_isHealthchecker(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// no healthcheck
	service1 := newServiceLazy("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.False(service1.isHealthchecker())

	// healthcheck ok
	service2 := newServiceLazy("foobar", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.False(service2.isHealthchecker())
	_, _ = service2.getInstance(nil)
	is.True(service2.isHealthchecker())

	// healthcheck ko
	service3 := newServiceLazy("foobar", func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.False(service3.isHealthchecker())
	_, _ = service3.getInstance(nil)
	is.True(service3.isHealthchecker())
}

// @TODO: missing tests for context
func TestServiceLazy_healthcheck(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ctx := context.Background()

	// no healthcheck
	service1 := newServiceLazy("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(service1.healthcheck(ctx))
	_, _ = service1.getInstance(nil)
	is.Nil(service1.healthcheck(ctx))

	// healthcheck ok
	service2 := newServiceLazy("foobar", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(service2.healthcheck(ctx))
	_, _ = service2.getInstance(nil)
	is.Nil(service2.healthcheck(ctx))

	// healthcheck ko
	service3 := newServiceLazy("foobar", func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.Nil(service3.healthcheck(ctx))
	_, _ = service3.getInstance(nil)
	is.Equal(assert.AnError, service3.healthcheck(ctx))
}

// @TODO: missing tests for context
func TestServiceLazy_isShutdowner(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// no shutdown
	service1 := newServiceLazy("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.False(service1.isShutdowner())

	// shutdown ok
	service2 := newServiceLazy("foobar", func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	})
	is.False(service2.isShutdowner())
	_, _ = service2.getInstance(nil)
	is.True(service2.isShutdowner())

	// shutdown ko
	service3 := newServiceLazy("foobar", func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	})
	is.False(service3.isShutdowner())
	_, _ = service3.getInstance(nil)
	is.True(service3.isShutdowner())
}

// @TODO: missing tests for context
func TestServiceLazy_shutdown(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ctx := context.Background()

	// no shutdown
	service1 := newServiceLazy("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.False(service1.built)
	is.Nil(service1.shutdown(ctx))
	_, _ = service1.getInstance(nil)
	is.True(service1.built)
	is.Nil(service1.shutdown(ctx))
	is.False(service1.built)

	// shutdown ok
	service2 := newServiceLazy("foobar", func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	})
	is.False(service2.built)
	is.Nil(service2.shutdown(ctx))
	_, _ = service2.getInstance(nil)
	is.True(service2.built)
	is.Nil(service2.shutdown(ctx))
	is.False(service2.built)

	// shutdown ko
	service3 := newServiceLazy("foobar", func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	})
	is.False(service3.built)
	is.Nil(service3.shutdown(ctx))
	_, _ = service3.getInstance(nil)
	is.True(service3.built)
	is.Equal(assert.AnError, service3.shutdown(ctx))
	is.False(service3.built)
}

func TestServiceLazy_clone(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// initial
	service1 := newServiceLazy("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.Equal("foobar", service1.getName())
	is.False(service1.built)
	_, _ = service1.getInstance(nil)
	is.True(service1.built)

	// clone
	service2, ok := service1.clone(nil).(*serviceLazy[lazyTest])
	is.True(ok)
	is.Equal("foobar", service2.getName())
	is.Empty(service2.instance)
	is.False(service2.built)
	_, _ = service2.getInstance(nil)
	is.NotEmpty(service2.instance)
	is.True(service2.built)

	// change initial and check clone
	service1.name = "baz"
	is.Equal("baz", service1.getName())
	is.Equal("foobar", service2.getName())
}

func TestServiceLazy_source(t *testing.T) {
	// @TODO
}
