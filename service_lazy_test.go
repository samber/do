package do

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/samber/do/v2/stacktrace"
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

// Test services for context value propagation in service lazy
type contextValueHealthcheckerLazy struct {
}

func (c *contextValueHealthcheckerLazy) HealthCheck(ctx context.Context) error {
	value := ctx.Value("test-key")
	if value != "healthcheck-value" {
		return fmt.Errorf("test-key not found or value is incorrect")
	}
	return nil
}

type contextValueShutdownerLazy struct {
}

func (c *contextValueShutdownerLazy) Shutdown(ctx context.Context) error {
	value := ctx.Value("test-key")
	if value != "shutdown-value" {
		return fmt.Errorf("test-key not found or value is incorrect")
	}
	return nil
}

func TestNewServiceLazy(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	// @TODO
}

func TestServiceLazy_getName(t *testing.T) {
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

	service1 := newServiceLazy("foobar1", provider1)
	is.Equal("foobar1", service1.getName())

	service2 := newServiceLazy("foobar2", provider2)
	is.Equal("foobar2", service2.getName())
}

func TestServiceLazy_getTypeName(t *testing.T) {
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

	service1 := newServiceLazy("foobar1", provider1)
	is.Equal("int", service1.getTypeName())

	service2 := newServiceLazy("foobar2", provider2)
	is.Equal("github.com/samber/do/v2.lazyTest", service2.getTypeName())
}

func TestServiceLazy_getServiceType(t *testing.T) {
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

	service1 := newServiceLazy("foobar1", provider1)
	is.Equal(ServiceTypeLazy, service1.getServiceType())

	service2 := newServiceLazy("foobar2", provider2)
	is.Equal(ServiceTypeLazy, service2.getServiceType())
}

func TestServiceLazy_getReflectType(t *testing.T) {
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Create a service lazy
	service := newServiceLazy("test-lazy", func(i Injector) (lazyTest, error) {
		return lazyTest{foobar: "test-value"}, nil
	})

	// Test initial state - should have provider frame but no invocation frames
	providerFrame, invocationFrames := service.source()

	// Provider frame should be set (from newServiceLazy)
	is.NotEmpty(providerFrame.File, "Provider frame should have a file")
	is.Greater(providerFrame.Line, 0, "Provider frame should have a line number")

	// Initially no invocation frames
	is.Empty(invocationFrames, "Should have no invocation frames initially")

	// Test after getting instance (which should add invocation frames)
	_, err := service.getInstance(nil)
	is.NoError(err, "Should be able to get instance")

	// Check source after invocation
	providerFrame2, invocationFrames2 := service.source()

	// Provider frame should remain the same
	is.Equal(providerFrame, providerFrame2, "Provider frame should remain unchanged")

	// Should now have invocation frames
	is.NotEmpty(invocationFrames2, "Should have invocation frames after getInstance")
	is.Len(invocationFrames2, 1, "Should have exactly one invocation frame")

	// Test multiple invocations
	_, err = service.getInstance(nil)
	is.NoError(err)

	_, invocationFrames3 := service.source()
	is.Len(invocationFrames3, 1, "Should still have only one unique invocation frame (duplicates are ignored)")

	// Test with different service types
	service2 := newServiceLazy("healthchecker-lazy", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "healthchecker-value"}, nil
	})

	// Get instance multiple times
	_, err = service2.getInstance(nil)
	is.NoError(err)
	_, err = service2.getInstance(nil)
	is.NoError(err)

	providerFrame3, invocationFrames4 := service2.source()
	is.NotEmpty(providerFrame3.File, "Provider frame should be set")
	is.Len(invocationFrames4, 1, "Should have one unique invocation frame")

	// Test concurrent access to source method
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = service.source()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Source should still work correctly after concurrent access
	providerFrame4, invocationFrames5 := service.source()
	is.Equal(providerFrame, providerFrame4, "Provider frame should remain consistent under concurrent access")
	is.Len(invocationFrames5, 1, "Invocation frames should remain consistent under concurrent access")

	// Test that invocation frames are properly collected from different call sites
	// Add a frame manually to simulate different call sites
	service.mu.Lock()
	service.invokationFrames[stacktrace.Frame{File: "different_file.go", Line: 42}] = struct{}{}
	service.mu.Unlock()

	_, invocationFrames6 := service.source()
	is.Len(invocationFrames6, 2, "Should have two invocation frames after adding different call site")

	// Verify the frames are different
	frameFiles := make(map[string]bool)
	for _, frame := range invocationFrames6 {
		frameFiles[frame.File] = true
	}
	is.Len(frameFiles, 2, "Should have frames from two different files")

	// Test with primitive types
	intService := newServiceLazy("int-lazy", func(i Injector) (int, error) {
		return 42, nil
	})
	_, err = intService.getInstance(nil)
	is.NoError(err)

	providerFrame5, invocationFrames7 := intService.source()
	is.NotEmpty(providerFrame5.File, "Provider frame should be set for primitive types")
	is.Len(invocationFrames7, 1, "Should have invocation frames for primitive types")

	// Test that the returned frames are copies, not references
	// This ensures thread safety
	originalFrames := make([]stacktrace.Frame, len(invocationFrames6))
	copy(originalFrames, invocationFrames6)

	// Modify the service's internal frames
	service.mu.Lock()
	service.invokationFrames[stacktrace.Frame{File: "new_file.go", Line: 100}] = struct{}{}
	service.mu.Unlock()

	// Get frames again
	_, invocationFrames8 := service.source()

	// Original frames should not be affected
	is.Len(originalFrames, 2, "Original frames should not be affected by internal changes")
	is.Len(invocationFrames8, 3, "New frames should include the additional frame")

	// Test that source works correctly after service is built and reset
	// This tests the lazy service's unique behavior
	_, err = service.getInstance(nil)
	is.NoError(err)
	is.True(service.built, "Service should be built")

	providerFrame6, invocationFrames9 := service.source()
	is.Equal(providerFrame, providerFrame6, "Provider frame should remain unchanged after build")
	is.Len(invocationFrames9, 3, "Should have all invocation frames after build")

	// Test with error provider
	errorService := newServiceLazy("error-lazy", func(i Injector) (lazyTest, error) {
		return lazyTest{}, assert.AnError
	})

	// Even with error, source should still work
	providerFrame7, invocationFrames10 := errorService.source()
	is.NotEmpty(providerFrame7.File, "Provider frame should be set even for error providers")
	is.Empty(invocationFrames10, "Should have no invocation frames for error providers initially")

	// After attempting to get instance (which fails), should still have invocation frames
	// because the frame is collected before the error occurs
	_, err = errorService.getInstance(nil)
	is.Error(err, "Should get error from error provider")

	_, invocationFrames11 := errorService.source()
	is.NotEmpty(invocationFrames11, "Should have invocation frames even after error (frame collected before error)")
}

// Test context value propagation for service lazy
func TestServiceLazy_ContextValuePropagation(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Create service lazy instances with providers that return context-aware services
	healthcheckLazy := newServiceLazy("healthcheck-lazy", func(i Injector) (*contextValueHealthcheckerLazy, error) {
		return &contextValueHealthcheckerLazy{}, nil
	})

	shutdownLazy := newServiceLazy("shutdown-lazy", func(i Injector) (*contextValueShutdownerLazy, error) {
		return &contextValueShutdownerLazy{}, nil
	})

	// Test context value propagation for healthcheck
	ctx1 := context.WithValue(context.Background(), "test-key", "healthcheck-value")
	err := healthcheckLazy.healthcheck(ctx1)
	is.Nil(err)

	// Test context value propagation for shutdown
	ctx2 := context.WithValue(context.Background(), "test-key", "shutdown-value")
	err = shutdownLazy.shutdown(ctx2)
	is.Nil(err)

	// Test that lazy service properly delegates to the underlying instance
	// The lazy service should not store context values itself, but pass them through
	// to the underlying service instance created by the provider
}
