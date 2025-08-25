package do

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type transientTest struct {
	foobar string
}

//nolint:unused
type transientTestHeathcheckerOK struct {
	foobar string
}

//nolint:unused
func (t *transientTestHeathcheckerOK) HealthCheck() error {
	return nil
}

//nolint:unused
type transientTestHeathcheckerKO struct {
	foobar string
}

//nolint:unused
func (t *transientTestHeathcheckerKO) HealthCheck() error {
	return assert.AnError
}

//nolint:unused
type transientTestShutdownerOK struct {
	foobar string
}

//nolint:unused
func (t *transientTestShutdownerOK) Shutdown() error {
	return nil
}

//nolint:unused
type transientTestShutdownerKO struct {
	foobar string
}

//nolint:unused
func (t *transientTestShutdownerKO) Shutdown() error {
	return assert.AnError
}

func TestNewServiceTransient(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test with int provider
	provider1 := func(i Injector) (int, error) {
		return 42, nil
	}
	service1 := newServiceTransient("int-service", provider1)
	is.Equal("int-service", service1.name)
	is.Equal("int", service1.typeName)
	is.NotNil(service1.provider)

	// Test with string provider
	provider2 := func(i Injector) (string, error) {
		return "test-value", nil
	}
	service2 := newServiceTransient("string-service", provider2)
	is.Equal("string-service", service2.name)
	is.Equal("string", service2.typeName)
	is.NotNil(service2.provider)

	// Test with struct provider
	testStruct := transientTest{foobar: "test"}
	provider3 := func(i Injector) (transientTest, error) {
		return testStruct, nil
	}
	service3 := newServiceTransient("struct-service", provider3)
	is.Equal("struct-service", service3.name)
	is.Equal("github.com/samber/do/v2.transientTest", service3.typeName)
	is.NotNil(service3.provider)

	// Test with pointer provider
	provider4 := func(i Injector) (*transientTest, error) {
		return &testStruct, nil
	}
	service4 := newServiceTransient("pointer-service", provider4)
	is.Equal("pointer-service", service4.name)
	is.Equal("*github.com/samber/do/v2.transientTest", service4.typeName)
	is.NotNil(service4.provider)

	// Test with interface provider
	provider5 := func(i Injector) (Healthchecker, error) {
		return &lazyTestHeathcheckerOK{foobar: "healthchecker"}, nil
	}
	service5 := newServiceTransient("interface-service", provider5)
	is.Equal("interface-service", service5.name)
	is.Equal("github.com/samber/do/v2.Healthchecker", service5.typeName)
	is.NotNil(service5.provider)

	// Test with slice provider
	provider6 := func(i Injector) ([]string, error) {
		return []string{"a", "b", "c"}, nil
	}
	service6 := newServiceTransient("slice-service", provider6)
	is.Equal("slice-service", service6.name)
	is.Equal("[]string", service6.typeName)
	is.NotNil(service6.provider)

	// Test with map provider
	provider7 := func(i Injector) (map[string]int, error) {
		return map[string]int{"key": 42}, nil
	}
	service7 := newServiceTransient("map-service", provider7)
	is.Equal("map-service", service7.name)
	is.Equal("map[string]int", service7.typeName)
	is.NotNil(service7.provider)

	// Test with boolean provider
	provider8 := func(i Injector) (bool, error) {
		return true, nil
	}
	service8 := newServiceTransient("bool-service", provider8)
	is.Equal("bool-service", service8.name)
	is.Equal("bool", service8.typeName)
	is.NotNil(service8.provider)

	// Test that all services are properly initialized
	is.NotNil(service1.provider, "Service1 provider should be set")
	is.NotNil(service2.provider, "Service2 provider should be set")
	is.NotNil(service3.provider, "Service3 provider should be set")
	is.NotNil(service4.provider, "Service4 provider should be set")
	is.NotNil(service5.provider, "Service5 provider should be set")
	is.NotNil(service6.provider, "Service6 provider should be set")
	is.NotNil(service7.provider, "Service7 provider should be set")
	is.NotNil(service8.provider, "Service8 provider should be set")
}

func TestServiceTransient_getName(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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

func TestServiceTransient_getReflectType(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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

	service1 := newServiceTransient("foobar1", provider1)
	is.Equal("int", service1.getReflectType().String())

	service2 := newServiceTransient("foobar2", provider2)
	is.Equal(pkgName+".lazyTest", service2.getReflectType().String())

	service3 := newServiceTransient("foobar3", provider3)
	is.Equal(pkgName+".Healthchecker", service3.getReflectType().String())

	service4 := newServiceTransient("foobar1", provider4)
	is.Equal("*"+pkgName+".lazyTest", service4.getReflectType().String())
}

func TestServiceTransient_getInstanceAny(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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

func TestServiceTransient_isHealthchecker(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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

	// Test that transient services don't support healthchecking regardless of provider type
	is.False(service1.isHealthchecker())
	is.False(service2.isHealthchecker())
	is.False(service3.isHealthchecker())

	// Test with HealthcheckerWithContext provider
	service4 := newServiceTransient("foobar", func(i Injector) (*eagerTestHeathcheckerWithContext, error) {
		return &eagerTestHeathcheckerWithContext{foobar: "test", shouldFail: false}, nil
	})
	is.False(service4.isHealthchecker())
	_, _ = service4.getInstance(nil)
	is.False(service4.isHealthchecker())
}

func TestServiceTransient_healthcheck(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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

	// Test with different context scenarios - transient services always return nil
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	service4 := newServiceTransient("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(service4.healthcheck(canceledCtx))

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)

	is.Nil(service4.healthcheck(timeoutCtx))

	futureCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	is.Nil(service4.healthcheck(futureCtx))

	// Test with HealthcheckerWithContext provider
	service5 := newServiceTransient("foobar", func(i Injector) (*eagerTestHeathcheckerWithContext, error) {
		return &eagerTestHeathcheckerWithContext{foobar: "test", shouldFail: true}, nil
	})
	is.Nil(service5.healthcheck(ctx))
	_, _ = service5.getInstance(nil)
	is.Nil(service5.healthcheck(ctx)) // Should still return nil even with failing healthchecker
}

func TestServiceTransient_isShutdowner(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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

	// Test that transient services don't support shutdowning regardless of provider type
	is.False(service1.isShutdowner())
	is.False(service2.isShutdowner())
	is.False(service3.isShutdowner())

	// Test with different shutdowner types
	service4 := newServiceTransient("foobar", func(i Injector) (*eagerTestShutdownerWithContext, error) {
		return &eagerTestShutdownerWithContext{foobar: "test"}, nil
	})
	is.False(service4.isShutdowner())
	_, _ = service4.getInstance(nil)
	is.False(service4.isShutdowner())

	service5 := newServiceTransient("foobar", func(i Injector) (*eagerTestShutdownerVoid, error) {
		return &eagerTestShutdownerVoid{foobar: "test"}, nil
	})
	is.False(service5.isShutdowner())
	_, _ = service5.getInstance(nil)
	is.False(service5.isShutdowner())
}

func TestServiceTransient_shutdown(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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

	// Test with different context scenarios - transient services always return nil
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	service4 := newServiceTransient("foobar", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "foobar"}, nil
	})
	is.Nil(service4.shutdown(canceledCtx))

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond)

	is.Nil(service4.shutdown(timeoutCtx))

	futureCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	is.Nil(service4.shutdown(futureCtx))

	// Test with different shutdowner types
	service5 := newServiceTransient("foobar", func(i Injector) (*eagerTestShutdownerWithContext, error) {
		return &eagerTestShutdownerWithContext{foobar: "test"}, nil
	})
	is.Nil(service5.shutdown(ctx))
	_, _ = service5.getInstance(nil)
	is.Nil(service5.shutdown(ctx)) // Should still return nil even with context shutdowner

	service6 := newServiceTransient("foobar", func(i Injector) (*eagerTestShutdownerVoid, error) {
		return &eagerTestShutdownerVoid{foobar: "test"}, nil
	})
	is.Nil(service6.shutdown(ctx))
	_, _ = service6.getInstance(nil)
	is.Nil(service6.shutdown(ctx)) // Should still return nil even with void shutdowner
}

func TestServiceTransient_clone(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// initial
	service1 := newServiceTransient("foobar", func(i Injector) (transientTest, error) {
		return transientTest{foobar: "foobar"}, nil
	})
	is.Equal("foobar", service1.getName())
	is.Equal("github.com/samber/do/v2.transientTest", service1.getTypeName())
	is.NotNil(service1.provider)

	// clone
	service2, ok := service1.clone(nil).(*serviceTransient[transientTest])
	is.True(ok)
	is.Equal("foobar", service2.getName())
	is.Equal("github.com/samber/do/v2.transientTest", service2.getTypeName())
	is.NotNil(service2.provider)

	// change initial and check clone
	service1.name = "baz"
	is.Equal("baz", service1.getName())
	is.Equal("foobar", service2.getName())

	// Test that clone is independent
	service1.typeName = "changed"
	is.Equal("changed", service1.getTypeName())
	is.Equal("github.com/samber/do/v2.transientTest", service2.getTypeName())

	// Test with different types
	service3 := newServiceTransient("int-service", func(i Injector) (int, error) {
		return 42, nil
	})
	service4, ok := service3.clone(nil).(*serviceTransient[int])
	is.True(ok)
	is.Equal("int-service", service4.getName())
	is.Equal("int", service4.getTypeName())

	// Test with interface types
	service5 := newServiceTransient("interface-service", func(i Injector) (Healthchecker, error) {
		return &lazyTestHeathcheckerOK{foobar: "healthchecker"}, nil
	})
	service6, ok := service5.clone(nil).(*serviceTransient[Healthchecker])
	is.True(ok)
	is.Equal("interface-service", service6.getName())
	is.Equal("github.com/samber/do/v2.Healthchecker", service6.getTypeName())

	// Test that cloned services can be used independently
	i := New()
	instance1, err1 := service1.getInstance(i)
	is.Nil(err1)
	is.Equal("foobar", instance1.foobar)

	instance2, err2 := service2.getInstance(i)
	is.Nil(err2)
	is.Equal("foobar", instance2.foobar)

	instance3, err3 := service3.getInstance(i)
	is.Nil(err3)
	is.Equal(42, instance3)

	instance4, err4 := service4.getInstance(i)
	is.Nil(err4)
	is.Equal(42, instance4)
}

func TestServiceTransient_source(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	service1 := newServiceTransient("foobar", func(i Injector) (transientTest, error) {
		return transientTest{foobar: "foobar"}, nil
	})
	_, _ = service1.getInstance(nil)

	a, b := service1.source()
	is.Empty(a)
	is.Empty(b)
}

// Test services for context value propagation in service transient
type contextValueHealthcheckerTransient struct {
	ID int
}

func (c *contextValueHealthcheckerTransient) HealthCheck(ctx context.Context) error {
	value := ctx.Value("test-key")
	if value != "healthcheck-value" {
		return fmt.Errorf("test-key not found or value is incorrect")
	}
	return nil
}

type contextValueShutdownerTransient struct {
	ID int
}

func (c *contextValueShutdownerTransient) Shutdown(ctx context.Context) error {
	value := ctx.Value("test-key")
	if value != "shutdown-value" {
		return fmt.Errorf("test-key not found or value is incorrect")
	}
	return nil
}

// Test context value propagation for service transient
// Note: Transient services don't support healthcheck/shutdown, so we test that behavior
func TestServiceTransient_ContextValuePropagation(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Create service transient instances with providers that return context-aware services
	callCount := 0
	healthcheckTransient := newServiceTransient("healthcheck-transient", func(i Injector) (*contextValueHealthcheckerTransient, error) {
		callCount++
		return &contextValueHealthcheckerTransient{ID: callCount}, nil
	})

	shutdownTransient := newServiceTransient("shutdown-transient", func(i Injector) (*contextValueShutdownerTransient, error) {
		return &contextValueShutdownerTransient{ID: 1}, nil
	})

	// Test that transient services don't support healthcheck (should return nil)
	ctx1 := context.WithValue(context.Background(), "test-key", "healthcheck-value")
	err := healthcheckTransient.healthcheck(ctx1)
	is.Nil(err) // Transient services return nil for healthcheck

	// Test that transient services don't support shutdown (should return nil)
	ctx2 := context.WithValue(context.Background(), "test-key", "shutdown-value")
	err = shutdownTransient.shutdown(ctx2)
	is.Nil(err) // Transient services return nil for shutdown

	// Verify that transient services are created fresh each time
	instance1, err := healthcheckTransient.getInstance(nil)
	is.Nil(err)

	instance2, err := healthcheckTransient.getInstance(nil)
	is.Nil(err)

	// These should be different instances (transient behavior)
	is.NotEqual(instance1, instance2)

	// Test that transient services properly ignore context values
	// This is the expected behavior for transient services - they don't participate
	// in healthcheck/shutdown lifecycle, so context values are not propagated
}
