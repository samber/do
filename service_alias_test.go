package do

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/samber/do/v2/stacktrace"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceAlias(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal("foobar1", service1.name)
	is.Equal(i, service1.scope)
	is.Equal("foobar2", service1.targetName)
}

func TestServiceAlias_getName(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal("foobar1", service1.getName())
}

func TestServiceAlias_getTypeName(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, int]("foobar1", i, "foobar2")
	is.Equal("int", service1.getTypeName())
}

func TestServiceAlias_getServiceType(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := New()

	service1 := newServiceAlias[string, string]("foobar1", i, "foobar2")
	is.Equal(ServiceTypeAlias, service1.getServiceType())
}

func TestServiceAlias_getReflectType(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	testWithTimeout(t, 100*time.Millisecond)
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
	_, _ = service2.(serviceWrapperGetInstanceAny).getInstanceAny(i2)
	is.True(service2.(serviceWrapperIsHealthchecker).isHealthchecker())

	// healthcheck ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerKO, Healthchecker](i3))
	service3, _ := i3.serviceGet("github.com/samber/do/v2.Healthchecker")
	is.False(service3.(serviceWrapperIsHealthchecker).isHealthchecker())
	_, _ = service3.(serviceWrapperGetInstanceAny).getInstanceAny(i3)
	is.True(service3.(serviceWrapperIsHealthchecker).isHealthchecker())

	// service not found (wrong type)
	i4 := New()
	service4 := newServiceAlias[*lazyTestHeathcheckerKO, Healthchecker]("github.com/samber/do/v2.Healthchecker", i4, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.False(service4.isHealthchecker())
	_, err4 := service4.getInstanceAny(i4)
	is.Error(err4)
	is.False(service4.isHealthchecker())

	// service not found (wrong name)
	i5 := New()
	service5 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("github.com/samber/do/v2.Healthchecker", i5, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.False(service5.isHealthchecker())
	_, err5 := service5.getInstanceAny(i5)
	is.Error(err5)
	is.False(service5.isHealthchecker())
}

// @TODO: missing tests for context
func TestServiceAlias_healthcheck(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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
	_, _ = service2.(serviceWrapper[Healthchecker]).getInstance(i2)
	is.Nil(service2.(serviceWrapper[Healthchecker]).healthcheck(ctx))

	// healthcheck ko
	i3 := New()
	Provide(i3, func(i Injector) (*lazyTestHeathcheckerKO, error) {
		return &lazyTestHeathcheckerKO{foobar: "foobar"}, nil
	})
	is.Nil(As[*lazyTestHeathcheckerKO, Healthchecker](i3))
	service3, _ := i3.serviceGet("github.com/samber/do/v2.Healthchecker")
	is.Nil(service3.(serviceWrapper[Healthchecker]).healthcheck(ctx))
	_, _ = service3.(serviceWrapper[Healthchecker]).getInstance(i3)
	is.Equal(assert.AnError, service3.(serviceWrapper[Healthchecker]).healthcheck(ctx))

	// service not found (wrong type)
	i4 := New()
	service4 := newServiceAlias[*lazyTestHeathcheckerKO, Healthchecker]("github.com/samber/do/v2.Healthchecker", i4, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.Nil(service4.healthcheck(ctx))
	_, _ = service4.getInstanceAny(i4)
	is.Nil(service4.healthcheck(ctx))

	// service not found (wrong name)
	i5 := New()
	service5 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("github.com/samber/do/v2.Healthchecker", i5, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.Nil(service5.healthcheck(ctx))
	_, _ = service5.getInstanceAny(i5)
	is.Nil(service5.healthcheck(ctx))
}

// @TODO: missing tests for context
func TestServiceAlias_isShutdowner(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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
	_, _ = service2.(serviceWrapper[ShutdownerWithContextAndError]).getInstance(i2)
	is.True(service2.(serviceWrapper[ShutdownerWithContextAndError]).isShutdowner())

	// shutdown ko
	i3 := New()
	Provide(i3, func(i Injector) (*contextValueShutdownerAlias, error) {
		return &contextValueShutdownerAlias{}, nil
	})
	is.Nil(As[*contextValueShutdownerAlias, ShutdownerWithContextAndError](i3))
	service3, _ := i3.serviceGet("github.com/samber/do/v2.ShutdownerWithContextAndError")
	is.False(service3.(serviceWrapper[ShutdownerWithContextAndError]).isShutdowner())
	_, _ = service3.(serviceWrapper[ShutdownerWithContextAndError]).getInstance(i3)
	is.True(service3.(serviceWrapper[ShutdownerWithContextAndError]).isShutdowner())

	// service not found (wrong type)
	i4 := New()
	service4 := newServiceAlias[*lazyTestShutdownerKO, Healthchecker]("*github.com/samber/do/v2.Healthchecker", i4, "*github.com/samber/do/v2.lazyTestShutdownerKO")
	is.False(service4.isShutdowner())
	_, _ = service4.getInstanceAny(i4)
	is.False(service4.isShutdowner())

	// service not found (wrong name)
	i5 := New()
	service5 := newServiceAlias[*lazyTestShutdownerOK, Healthchecker]("*github.com/samber/do/v2.Healthchecker", i5, "*github.com/samber/do/v2.lazyTestShutdownerKO")
	is.False(service5.isShutdowner())
	_, _ = service5.getInstanceAny(i5)
	is.False(service5.isShutdowner())
}

func TestServiceAlias_shutdown(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
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
	_, _ = service2.(serviceWrapper[ShutdownerWithContextAndError]).getInstance(i2)
	is.Nil(service2.(serviceWrapper[ShutdownerWithContextAndError]).shutdown(ctx))

	// shutdown ko
	i3 := New()
	Provide(i3, func(i Injector) (*contextValueShutdownerAlias, error) {
		return &contextValueShutdownerAlias{}, nil
	})
	is.Nil(As[*contextValueShutdownerAlias, ShutdownerWithContextAndError](i3))
	service3, _ := i3.serviceGet("github.com/samber/do/v2.ShutdownerWithContextAndError")
	is.Nil(service3.(serviceWrapper[ShutdownerWithContextAndError]).shutdown(ctx))
	_, _ = service3.(serviceWrapper[ShutdownerWithContextAndError]).getInstance(i3)
	is.Error(service3.(serviceWrapper[ShutdownerWithContextAndError]).shutdown(ctx))

	// service not found (wrong type)
	i4 := New()
	service4 := newServiceAlias[*lazyTestShutdownerKO, Healthchecker]("github.com/samber/do/v2.Healthchecker", i4, "*github.com/samber/do/v2.lazyTestShutdownerKO")
	is.Nil(service4.shutdown(ctx))
	_, _ = service4.getInstanceAny(i4)
	is.Nil(service4.shutdown(ctx))

	// service not found (wrong name)
	i5 := New()
	service5 := newServiceAlias[*lazyTestShutdownerOK, Healthchecker]("github.com/samber/do/v2.Healthchecker", i5, "*github.com/samber/do/v2.lazyTestHeathcheckerKO")
	is.Nil(service5.shutdown(ctx))
	_, _ = service5.getInstanceAny(i5)
	is.Nil(service5.shutdown(ctx))
}

func TestServiceAlias_clone(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Create original service alias
	originalScope := New()
	newScope := New()

	originalService := newServiceAlias[string, int]("test-alias", originalScope, "target-service")

	// Set some invocation frames to test that they are reset
	originalService.mu.Lock()
	originalService.invokationFrames[stacktrace.Frame{File: "test.go", Line: 42}] = struct{}{}
	originalService.invokationFrames[stacktrace.Frame{File: "test2.go", Line: 100}] = struct{}{}
	originalService.invokationFramesCounter = 5
	originalService.mu.Unlock()

	// Clone the service
	clonedServiceAny := originalService.clone(newScope)
	clonedService, ok := clonedServiceAny.(*serviceAlias[string, int])
	is.True(ok, "Clone should return the correct type")

	// Test that all fields are properly copied
	is.Equal(originalService.name, clonedService.name, "Name should be copied")
	is.Equal(originalService.typeName, clonedService.typeName, "Type name should be copied")
	is.Equal(originalService.targetName, clonedService.targetName, "Target name should be copied")
	is.Equal(originalService.providerFrame, clonedService.providerFrame, "Provider frame should be copied")

	// Test that scope is replaced with new scope
	is.Equal(newScope, clonedService.scope, "Scope should be replaced with new scope")
	is.NotEqual(originalService.scope, clonedService.scope, "Original and cloned scopes should be different")

	// Test that invocation frames are reset
	is.Empty(clonedService.invokationFrames, "Invocation frames should be reset to empty")
	is.Equal(uint32(0), clonedService.invokationFramesCounter, "Invocation frames counter should be reset to 0")

	// Test that original service is not affected
	is.NotEmpty(originalService.invokationFrames, "Original service invocation frames should remain unchanged")
	is.Equal(uint32(5), originalService.invokationFramesCounter, "Original service counter should remain unchanged")

	// Test that cloned service has a new mutex (not shared with original)
	originalService.mu.Lock()
	clonedService.mu.Lock()
	// If mutexes were shared, this would deadlock
	clonedService.mu.Unlock()
	originalService.mu.Unlock()

	// Test clone with different types
	originalService2 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("healthchecker-alias", originalScope, "healthchecker-target")
	clonedService2Any := originalService2.clone(newScope)
	clonedService2, ok := clonedService2Any.(*serviceAlias[*lazyTestHeathcheckerOK, Healthchecker])
	is.True(ok, "Clone should return the correct type for different generic types")
	is.Equal(originalService2.name, clonedService2.name)
	is.Equal(originalService2.typeName, clonedService2.typeName)
	is.Equal(newScope, clonedService2.scope)

	// Test clone with nil scope
	originalService3 := newServiceAlias[bool, string]("bool-alias", nil, "bool-target")
	clonedService3Any := originalService3.clone(newScope)
	clonedService3, ok := clonedService3Any.(*serviceAlias[bool, string])
	is.True(ok)
	is.Equal(newScope, clonedService3.scope, "Clone should work even when original scope is nil")

	// Test that cloned service can be used independently
	// This tests that the clone is a completely independent copy
	clonedService.mu.Lock()
	clonedService.invokationFrames[stacktrace.Frame{File: "cloned.go", Line: 1}] = struct{}{}
	clonedService.invokationFramesCounter = 10
	clonedService.mu.Unlock()

	// Original should not be affected
	is.NotContains(originalService.invokationFrames, stacktrace.Frame{File: "cloned.go", Line: 1})
	is.Equal(uint32(5), originalService.invokationFramesCounter)

	// Cloned should have the new data
	is.Contains(clonedService.invokationFrames, stacktrace.Frame{File: "cloned.go", Line: 1})
	is.Equal(uint32(10), clonedService.invokationFramesCounter)
}

func TestServiceAlias_source(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Create a service alias
	scope := New()
	service := newServiceAlias[string, string]("test-alias", scope, "string")

	// Test initial state - should have provider frame but no invocation frames
	providerFrame, invocationFrames := service.source()

	// Provider frame should be set (from newServiceAlias)
	is.NotEmpty(providerFrame.File, "Provider frame should have a file")
	is.Greater(providerFrame.Line, 0, "Provider frame should have a line number")

	// Initially no invocation frames
	is.Empty(invocationFrames, "Should have no invocation frames initially")

	// Test after getting instance (which should add invocation frames)
	// First, we need to provide the target service
	Provide(scope, func(i Injector) (string, error) {
		return "test-value", nil
	})

	// Get instance to trigger invocation frame collection
	_, err := service.getInstance(scope)
	is.NoError(err, "Should be able to get instance")

	// Check source after invocation
	providerFrame2, invocationFrames2 := service.source()

	// Provider frame should remain the same
	is.Equal(providerFrame, providerFrame2, "Provider frame should remain unchanged")

	// Should now have invocation frames
	is.NotEmpty(invocationFrames2, "Should have invocation frames after getInstance")
	is.Len(invocationFrames2, 1, "Should have exactly one invocation frame")

	// Test multiple invocations
	_, err = service.getInstance(scope)
	is.NoError(err)

	_, invocationFrames3 := service.source()
	is.Len(invocationFrames3, 1, "Should still have only one unique invocation frame (duplicates are ignored)")

	// Test with different service types
	service2 := newServiceAlias[*lazyTestHeathcheckerOK, Healthchecker]("healthchecker-alias", scope, "healthchecker-target")

	// Provide the target service
	Provide(scope, func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "test"}, nil
	})
	ProvideNamed(scope, "healthchecker-target", func(i Injector) (*lazyTestHeathcheckerOK, error) {
		return &lazyTestHeathcheckerOK{foobar: "target"}, nil
	})

	// Get instance multiple times
	_, err = service2.getInstance(scope)
	is.NoError(err)
	_, err = service2.getInstance(scope)
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
	// This tests the frame collection mechanism

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
}

// Test services for context value propagation in service alias
type contextValueHealthcheckerAlias struct {
}

func (c *contextValueHealthcheckerAlias) HealthCheck(ctx context.Context) error {
	value := ctx.Value("test-key")
	if value != "healthcheck-value" {
		return fmt.Errorf("test-key not found or value is incorrect")
	}
	return nil
}

type contextValueShutdownerAlias struct {
}

func (c *contextValueShutdownerAlias) Shutdown(ctx context.Context) error {
	value := ctx.Value("test-key")
	if value != "shutdown-value" {
		return fmt.Errorf("test-key not found or value is incorrect")
	}
	return nil
}

// Test context value propagation for service alias
func TestServiceAlias_ContextValuePropagation(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Create injector and services
	injector := New()

	// Create target services
	healthcheckService := &contextValueHealthcheckerAlias{}
	shutdownService := &contextValueShutdownerAlias{}

	// Provide target services
	ProvideNamedValue(injector, "target-healthcheck", healthcheckService)
	ProvideNamedValue(injector, "target-shutdown", shutdownService)

	// Create service aliases
	healthcheckAlias := newServiceAlias[*contextValueHealthcheckerAlias, *contextValueHealthcheckerAlias]("healthcheck-alias", injector, "target-healthcheck")
	shutdownAlias := newServiceAlias[*contextValueShutdownerAlias, *contextValueShutdownerAlias]("shutdown-alias", injector, "target-shutdown")

	// Invoke services to make them healthcheckable/shutdownable
	_, err1 := healthcheckAlias.getInstance(injector)
	_, err2 := shutdownAlias.getInstance(injector)
	is.Nil(err1)
	is.Nil(err2)

	// Test context value propagation for healthcheck
	ctx1 := context.WithValue(context.Background(), "test-key", "healthcheck-value")
	err := healthcheckAlias.healthcheck(ctx1)
	is.Nil(err)

	// Test context value propagation for shutdown
	ctx2 := context.WithValue(context.Background(), "test-key", "shutdown-value")
	err = shutdownAlias.shutdown(ctx2)
	is.Nil(err)

	// Test that alias properly delegates to target service
	// The alias should not store context values itself, but pass them through
	// to the underlying target service
}
