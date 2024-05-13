package do

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type transientTest struct {
	foobar string
}

// nolint:unused
type transientTestHeathcheckerOK struct {
	foobar string
}

// nolint:unused
func (t *transientTestHeathcheckerOK) HealthCheck() error {
	return nil
}

// nolint:unused
type transientTestHeathcheckerKO struct {
	foobar string
}

// nolint:unused
func (t *transientTestHeathcheckerKO) HealthCheck() error {
	return assert.AnError
}

// nolint:unused
type transientTestShutdownerOK struct {
	foobar string
}

// nolint:unused
func (t *transientTestShutdownerOK) Shutdown() error {
	return nil
}

// nolint:unused
type transientTestShutdownerKO struct {
	foobar string
}

// nolint:unused
func (t *transientTestShutdownerKO) Shutdown() error {
	return assert.AnError
}

func TestNewServiceTransient(t *testing.T) {
	// @TODO
}

func TestServiceTransient_getName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := transientTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transientTest, error) {
		return test, nil
	}

	service1 := newServiceTransient("foobar1", provider1)
	is.Equal("foobar1", service1.getName())

	service2 := newServiceTransient("foobar2", provider2)
	is.Equal("foobar2", service2.getName())
}

func TestServiceTransient_getTypeName(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := transientTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transientTest, error) {
		return test, nil
	}

	service1 := newServiceTransient("foobar1", provider1)
	is.Equal("int", service1.getTypeName())

	service2 := newServiceTransient("foobar2", provider2)
	is.Equal("github.com/samber/do/v2.transientTest", service2.getTypeName())
}

func TestServiceTransient_getServiceType(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := transientTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transientTest, error) {
		return test, nil
	}

	service1 := newServiceTransient("foobar1", provider1)
	is.Equal(ServiceTypeTransient, service1.getServiceType())

	service2 := newServiceTransient("foobar2", provider2)
	is.Equal(ServiceTypeTransient, service2.getServiceType())
}

func TestServiceTransient_getEmptyInstance(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	svc := newServiceTransient("foobar", func(i Injector) (*transientTest, error) {
		return &transientTest{foobar: "foobar"}, nil
	})
	is.Empty(svc.getEmptyInstance())
	is.EqualValues((*transientTest)(nil), svc.getEmptyInstance())
}

func TestServiceTransient_getInstanceAny(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := transientTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transientTest, error) {
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
	service1 := newServiceTransient("foobar", provider1)
	instance1, err1 := service1.getInstanceAny(i)
	is.Nil(err1)
	is.Equal(42, instance1)

	// struct
	service2 := newServiceTransient("hello", provider2)
	instance2, err2 := service2.getInstanceAny(i)
	is.Nil(err2)
	is.Equal(test, instance2)

	// provider panics, but panic is catched by getInstanceAny
	is.NotPanics(func() {
		service3 := newServiceTransient("baz", provider3)
		_, _ = service3.getInstanceAny(i)
	})

	// provider panics, but panic is catched by getInstanceAny
	is.NotPanics(func() {
		service4 := newServiceTransient("plop", provider4)
		instance4, err4 := service4.getInstanceAny(i)
		is.NotNil(err4)
		is.Empty(instance4)
		expected := fmt.Errorf("error")
		is.Equal(expected, err4)
	})

	// provider returning error
	is.NotPanics(func() {
		service5 := newServiceTransient("plop", provider5)
		instance5, err5 := service5.getInstanceAny(i)
		is.NotNil(err5)
		is.Empty(instance5)
		expected := fmt.Errorf("error")
		is.Equal(expected, err5)
	})
}

func TestServiceTransient_getInstance(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	test := transientTest{foobar: "foobar"}

	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	provider2 := func(i Injector) (transientTest, error) {
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
	service1 := newServiceTransient("foobar", provider1)
	instance1, err1 := service1.getInstance(i)
	is.Nil(err1)
	is.Equal(42, instance1)

	// struct
	service2 := newServiceTransient("hello", provider2)
	instance2, err2 := service2.getInstance(i)
	is.Nil(err2)
	is.Equal(test, instance2)

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service3 := newServiceTransient("baz", provider3)
		_, _ = service3.getInstance(i)
	})

	// provider panics, but panic is catched by getInstance
	is.NotPanics(func() {
		service4 := newServiceTransient("plop", provider4)
		instance4, err4 := service4.getInstance(i)
		is.NotNil(err4)
		is.Empty(instance4)
		expected := fmt.Errorf("error")
		is.Equal(expected, err4)
	})

	// provider returning error
	is.NotPanics(func() {
		service5 := newServiceTransient("plop", provider5)
		instance5, err5 := service5.getInstance(i)
		is.NotNil(err5)
		is.Empty(instance5)
		expected := fmt.Errorf("error")
		is.Equal(expected, err5)
	})
}

// @TODO: missing tests for context
func TestServiceTransient_isHealthchecker(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// no healthcheck
	service1 := newServiceTransient("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.False(service1.isHealthchecker())

	// healthcheck ok
	service2 := newServiceTransient("foobar", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.False(service2.isHealthchecker())
	_, _ = service2.getInstance(nil)
	is.False(service2.isHealthchecker())

	// healthcheck ko
	service3 := newServiceTransient("foobar", func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.False(service3.isHealthchecker())
	_, _ = service3.getInstance(nil)
	is.False(service3.isHealthchecker())
}

// @TODO: missing tests for context
func TestServiceTransient_healthcheck(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ctx := context.Background()

	// no healthcheck
	service1 := newServiceTransient("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(service1.healthcheck(ctx))
	_, _ = service1.getInstance(nil)
	is.Nil(service1.healthcheck(ctx))

	// healthcheck ok
	service2 := newServiceTransient("foobar", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "foobar"}, nil
	})
	is.Nil(service2.healthcheck(ctx))
	_, _ = service2.getInstance(nil)
	is.Nil(service2.healthcheck(ctx))

	// healthcheck ko
	service3 := newServiceTransient("foobar", func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.Nil(service3.healthcheck(ctx))
	_, _ = service3.getInstance(nil)
	is.Nil(service3.healthcheck(ctx))
}

// @TODO: missing tests for context
func TestServiceTransient_isShutdowner(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	// no shutdown
	service1 := newServiceTransient("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.False(service1.isShutdowner())

	// shutdown ok
	service2 := newServiceTransient("foobar", func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	})
	is.False(service2.isShutdowner())
	_, _ = service2.getInstance(nil)
	is.False(service2.isShutdowner())

	// shutdown ko
	service3 := newServiceTransient("foobar", func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	})
	is.False(service3.isShutdowner())
	_, _ = service3.getInstance(nil)
	is.False(service3.isShutdowner())
}

// @TODO: missing tests for context
func TestServiceTransient_shutdown(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	ctx := context.Background()

	// no shutdown
	service1 := newServiceTransient("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(service1.shutdown(ctx))
	_, _ = service1.getInstance(nil)
	is.Nil(service1.shutdown(ctx))

	// shutdown ok
	service2 := newServiceTransient("foobar", func(i Injector) (*lazyTestShutdownerOK, error) {
		return &lazyTestShutdownerOK{foobar: "foobar"}, nil
	})
	is.Nil(service2.shutdown(ctx))
	_, _ = service2.getInstance(nil)
	is.Nil(service2.shutdown(ctx))

	// shutdown ko
	service3 := newServiceTransient("foobar", func(i Injector) (*lazyTestShutdownerKO, error) {
		return &lazyTestShutdownerKO{foobar: "foobar"}, nil
	})
	is.Nil(service3.shutdown(ctx))
	_, _ = service3.getInstance(nil)
	is.Nil(service3.shutdown(ctx))
}

func TestServiceTransient_clone(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	// @TODO

	// initial
	service1 := newServiceTransient("foobar", func(i Injector) (transientTest, error) {
		return transientTest{foobar: "foobar"}, nil
	})
	is.Equal("foobar", service1.getName())

	// clone
	service2, ok := service1.clone().(*serviceTransient[transientTest])
	is.True(ok)
	is.Equal("foobar", service2.getName())

	// change initial and check clone
	service1.name = "baz"
	is.Equal("baz", service1.getName())
	is.Equal("foobar", service2.getName())
}

func TestServiceTransient_source(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	service1 := newServiceTransient("foobar", func(i Injector) (transientTest, error) {
		return transientTest{foobar: "foobar"}, nil
	})
	_, _ = service1.getInstance(nil)

	a, b := service1.source()
	is.Empty(a)
	is.Empty(b)
}
