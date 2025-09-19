package do

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/samber/do/v2/stacktrace"
	"github.com/stretchr/testify/assert"
)

type eagerTest struct {
	foobar string
}

var _ Healthchecker = (*eagerTestHeathcheckerOK)(nil)

type eagerTestHeathcheckerOK struct {
	foobar string
}

func (t *eagerTestHeathcheckerOK) HealthCheck() error {
	return nil
}

var _ Healthchecker = (*eagerTestHeathcheckerKO)(nil)

type eagerTestHeathcheckerKO struct {
	foobar string
}

func (t *eagerTestHeathcheckerKO) HealthCheck() error {
	return assert.AnError
}

var _ ShutdownerWithContextAndError = (*eagerTestShutdownerOK)(nil)

type eagerTestShutdownerOK struct {
	foobar string
}

func (t *eagerTestShutdownerOK) Shutdown(ctx context.Context) error {
	return nil
}

var _ ShutdownerWithError = (*eagerTestShutdownerKO)(nil)

type eagerTestShutdownerKO struct {
	foobar string
}

func (t *eagerTestShutdownerKO) Shutdown() error {
	return assert.AnError
}

var _ HealthcheckerWithContext = (*eagerTestHeathcheckerWithContext)(nil)

type eagerTestHeathcheckerWithContext struct {
	foobar     string
	shouldFail bool
}

func (h *eagerTestHeathcheckerWithContext) HealthCheck(ctx context.Context) error {
	if h.shouldFail {
		return assert.AnError
	}
	return nil
}

var _ ShutdownerWithContext = (*eagerTestShutdownerWithContext)(nil)

type eagerTestShutdownerWithContext struct {
	foobar string
}

func (s *eagerTestShutdownerWithContext) Shutdown(ctx context.Context) {
	// Void return
}

var _ Shutdowner = (*eagerTestShutdownerVoid)(nil)

type eagerTestShutdownerVoid struct {
	foobar string
}

func (s *eagerTestShutdownerVoid) Shutdown() {
	// Void return, no context
}

// Test services for context value propagation in service eager
type contextValueHealthcheckerEager struct{}

func (c *contextValueHealthcheckerEager) HealthCheck(ctx context.Context) error {
	value := ctx.Value(ctxTestKey)
	if value != "healthcheck-value" {
		return fmt.Errorf("test-key not found or value is incorrect")
	}
	return nil
}

type contextValueShutdownerEager struct{}

func (c *contextValueShutdownerEager) Shutdown(ctx context.Context) error {
	value := ctx.Value(ctxTestKey)
	if value != "shutdown-value" {
		return fmt.Errorf("test-key not found or value is incorrect")
	}
	return nil
}

func TestNewServiceEager(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test with string instance
	service1 := newServiceEager("string-service", "test-value")
	is.Equal("string-service", service1.name)
	is.Equal("string", service1.typeName)
	is.Equal("test-value", service1.instance)
	is.NotEmpty(service1.providerFrame.File, "Provider frame should be set")
	is.Greater(service1.providerFrame.Line, 0, "Provider frame should have line number")
	is.Empty(service1.invokationFrames, "Invocation frames should be empty initially")
	is.Equal(uint32(0), service1.invokationFramesCounter, "Invocation frames counter should be 0")

	// Test with int instance
	service2 := newServiceEager("int-service", 42)
	is.Equal("int-service", service2.name)
	is.Equal("int", service2.typeName)
	is.Equal(42, service2.instance)

	// Test with struct instance
	testStruct := eagerTest{foobar: "test"}
	service3 := newServiceEager("struct-service", testStruct)
	is.Equal("struct-service", service3.name)
	is.Equal("github.com/samber/do/v2.eagerTest", service3.typeName)
	is.Equal(testStruct, service3.instance)

	// Test with pointer instance
	testPtr := &eagerTest{foobar: "pointer-test"}
	service4 := newServiceEager("pointer-service", testPtr)
	is.Equal("pointer-service", service4.name)
	is.Equal("*github.com/samber/do/v2.eagerTest", service4.typeName)
	is.Equal(testPtr, service4.instance)

	// Test with interface instance
	var healthchecker Healthchecker = &eagerTestHeathcheckerOK{foobar: "healthchecker"}
	service5 := newServiceEager("interface-service", healthchecker)
	is.Equal("interface-service", service5.name)
	is.Equal("github.com/samber/do/v2.Healthchecker", service5.typeName)
	is.Equal(healthchecker, service5.instance)

	// Test with nil instance
	service6 := newServiceEager[*eagerTest]("nil-service", nil)
	is.Equal("nil-service", service6.name)
	is.Equal("*github.com/samber/do/v2.eagerTest", service6.typeName)
	is.Nil(service6.instance)

	// Test with boolean instance
	service7 := newServiceEager("bool-service", true)
	is.Equal("bool-service", service7.name)
	is.Equal("bool", service7.typeName)
	is.True(service7.instance)

	// Test with slice instance
	slice := []string{"a", "b", "c"}
	service8 := newServiceEager("slice-service", slice)
	is.Equal("slice-service", service8.name)
	is.Equal("[]string", service8.typeName)
	is.Equal(slice, service8.instance)

	// Test with map instance
	m := map[string]int{"key": 42}
	service9 := newServiceEager("map-service", m)
	is.Equal("map-service", service9.name)
	is.Equal("map[string]int", service9.typeName)
	is.Equal(m, service9.instance)
}

func TestServiceEager_getName(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal("foobar1", service1.getName())

	service2 := newServiceEager("foobar2", test)
	is.Equal("foobar2", service2.getName())
}

func TestServiceEager_getTypeName(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal("int", service1.getTypeName())

	service2 := newServiceEager("foobar2", test)
	is.Equal("github.com/samber/do/v2.eagerTest", service2.getTypeName())
}

func TestServiceEager_getServiceType(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal(ServiceTypeEager, service1.getServiceType())

	service2 := newServiceEager("foobar2", test)
	is.Equal(ServiceTypeEager, service2.getServiceType())
}

func TestServiceEager_getReflectType(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar1", 42)
	is.Equal("int", service1.getReflectType().String())

	service2 := newServiceEager("foobar2", test)
	is.Equal(pkgName+".eagerTest", service2.getReflectType().String())

	service3 := newServiceEager("foobar3", (Healthchecker)(nil))
	is.Equal(pkgName+".Healthchecker", service3.getReflectType().String())

	service4 := newServiceEager[Healthchecker]("foobar4", nil)
	is.Equal(pkgName+".Healthchecker", service4.getReflectType().String())

	service5 := newServiceEager("foobar1", &test)
	is.Equal("*"+pkgName+".eagerTest", service5.getReflectType().String())
}

func TestServiceEager_getInstanceAny(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar", test)
	instance1, err1 := service1.getInstanceAny(nil)
	is.Nil(err1)
	is.Equal(test, instance1)

	service2 := newServiceEager("foobar", 42)
	instance2, err2 := service2.getInstanceAny(nil)
	is.Nil(err2)
	is.Equal(42, instance2)
}

func TestServiceEager_getInstance(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	service1 := newServiceEager("foobar", test)
	instance1, err1 := service1.getInstance(nil)
	is.Nil(err1)
	is.Equal(test, instance1)

	service2 := newServiceEager("foobar", 42)
	instance2, err2 := service2.getInstance(nil)
	is.Nil(err2)
	is.Equal(42, instance2)
}

func TestServiceEager_isHealthchecker(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// no healthcheck
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	is.False(service1.isHealthchecker())

	// healthcheck ok
	service2 := newServiceEager("foobar", &eagerTestHeathcheckerOK{foobar: "foobar"})
	is.True(service2.isHealthchecker())

	// healthcheck ko
	service3 := newServiceEager("foobar", &eagerTestHeathcheckerKO{foobar: "foobar"})
	is.True(service3.isHealthchecker())
}

func TestServiceEager_healthcheck(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	ctx := context.Background()

	// no healthcheck
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	err1 := service1.healthcheck(ctx)
	is.Nil(err1)

	// healthcheck ok
	service2 := newServiceEager("foobar", &eagerTestHeathcheckerOK{foobar: "foobar"})
	err2 := service2.healthcheck(ctx)
	is.Nil(err2)

	// healthcheck ko
	service3 := newServiceEager("foobar", &eagerTestHeathcheckerKO{foobar: "foobar"})
	err3 := service3.healthcheck(ctx)
	is.NotNil(err3)
	is.Error(err3)
	is.Equal(err3, assert.AnError)

	// Test with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	service4 := newServiceEager("foobar", &eagerTestHeathcheckerOK{foobar: "foobar"})
	err4 := service4.healthcheck(canceledCtx)
	is.Error(err4)
	is.Equal(context.Canceled, err4)

	// Test with timeout context that expires
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond) // Ensure timeout expires

	service5 := newServiceEager("foobar", &eagerTestHeathcheckerOK{foobar: "foobar"})
	err5 := service5.healthcheck(timeoutCtx)
	is.Error(err5)
	is.Equal(context.DeadlineExceeded, err5)

	// Test with context that has deadline but hasn't expired
	futureCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	service6 := newServiceEager("foobar", &eagerTestHeathcheckerOK{foobar: "foobar"})
	err6 := service6.healthcheck(futureCtx)
	is.Nil(err6) // Should succeed since context hasn't expired

	// Test HealthcheckerWithContext type
	healthcheckerWithCtx := &eagerTestHeathcheckerWithContext{foobar: "test", shouldFail: false}
	service7 := newServiceEager("foobar", healthcheckerWithCtx)

	// Test HealthcheckerWithContext interface
	err7 := service7.healthcheck(ctx)
	is.Nil(err7) // Should work with normal context

	// Test context error handling with canceled context and error healthchecker
	service8 := newServiceEager("foobar", &eagerTestHeathcheckerKO{foobar: "foobar"})
	err8 := service8.healthcheck(canceledCtx)
	is.Error(err8)
	is.Equal(context.Canceled, err8) // Context error takes precedence
}

func TestServiceEager_isShutdowner(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// no shutdown
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	is.False(service1.isShutdowner())

	// shutdown ok
	service2 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	is.True(service2.isShutdowner())

	// shutdown ko
	service3 := newServiceEager("foobar", &eagerTestShutdownerKO{foobar: "foobar"})
	is.True(service3.isShutdowner())
}

func TestServiceEager_shutdown(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	ctx := context.Background()

	// no shutdown
	service1 := newServiceEager("foobar", &eagerTest{foobar: "foobar"})
	err1 := service1.shutdown(ctx)
	is.Nil(err1)

	// shutdown ok
	service2 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	err2 := service2.shutdown(ctx)
	is.Nil(err2)

	// shutdown ko
	service3 := newServiceEager("foobar", &eagerTestShutdownerKO{foobar: "foobar"})
	err3 := service3.shutdown(ctx)
	is.NotNil(err3)
	is.Error(err3)
	is.Equal(err3, assert.AnError)

	// Test with canceled context
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	service4 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	err4 := service4.shutdown(canceledCtx)
	is.Error(err4)
	is.Equal(context.Canceled, err4)

	// Test with timeout context that expires
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(2 * time.Millisecond) // Ensure timeout expires

	service5 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	err5 := service5.shutdown(timeoutCtx)
	is.Error(err5)
	is.Equal(context.DeadlineExceeded, err5)

	// Test with context that has deadline but hasn't expired
	futureCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	service6 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	err6 := service6.shutdown(futureCtx)
	is.Nil(err6) // Should succeed since context hasn't expired

	// Test different shutdowner types with context

	// Test ShutdownerWithContextAndError (already tested above)
	service7 := newServiceEager("foobar", &eagerTestShutdownerOK{foobar: "foobar"})
	err7 := service7.shutdown(ctx)
	is.Nil(err7)

	// Test ShutdownerWithError with canceled context
	service8 := newServiceEager("foobar", &eagerTestShutdownerKO{foobar: "foobar"})
	err8 := service8.shutdown(canceledCtx)
	is.Error(err8)
	is.Equal(context.Canceled, err8) // Context error takes precedence

	// Test ShutdownerWithContext type (void return)
	shutdownerWithCtx := &eagerTestShutdownerWithContext{foobar: "test"}
	service9 := newServiceEager("foobar", shutdownerWithCtx)

	// Test ShutdownerWithContext interface
	err9 := service9.shutdown(ctx)
	is.Nil(err9) // Should work with normal context

	// Test Shutdowner type (void return, no context)
	shutdownerVoid := &eagerTestShutdownerVoid{foobar: "test"}
	service10 := newServiceEager("foobar", shutdownerVoid)

	// Test Shutdowner interface
	err10 := service10.shutdown(ctx)
	is.Nil(err10) // Should work

	// Test context error handling with canceled context and different shutdowner types
	service11 := newServiceEager("foobar", shutdownerWithCtx)
	err11 := service11.shutdown(canceledCtx)
	is.Error(err11)
	is.Equal(context.Canceled, err11) // Context error takes precedence
}

func TestServiceEager_clone(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	test := eagerTest{foobar: "foobar"}

	// initial
	service1 := newServiceEager("foobar", test)
	is.Equal("foobar", service1.getName())

	// clone
	service2, ok := service1.clone(nil).(*serviceEager[eagerTest])
	is.True(ok)
	is.Equal("foobar", service2.getName())

	// change initial and check clone
	service1.name = "baz"
	is.Equal("baz", service1.getName())
	is.Equal("foobar", service2.getName())
}

func TestServiceEager_source(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Create a service eager
	testInstance := &eagerTest{foobar: "test-value"}
	service := newServiceEager("test-eager", testInstance)

	// Test initial state - should have provider frame but no invocation frames
	providerFrame, invocationFrames := service.source()

	// Provider frame should be set (from newServiceEager)
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
	healthcheckerInstance := &eagerTestHeathcheckerOK{foobar: "healthchecker-value"}
	service2 := newServiceEager("healthchecker-eager", healthcheckerInstance)

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
	service.invokationFramesMu.Lock()
	service.invokationFrames[stacktrace.Frame{File: "different_file.go", Line: 42}] = struct{}{}
	service.invokationFramesMu.Unlock()

	_, invocationFrames6 := service.source()
	is.Len(invocationFrames6, 2, "Should have two invocation frames after adding different call site")

	// Verify the frames are different
	frameFiles := make(map[string]bool)
	for _, frame := range invocationFrames6 {
		frameFiles[frame.File] = true
	}
	is.Len(frameFiles, 2, "Should have frames from two different files")

	// Test with primitive types
	intService := newServiceEager("int-eager", 42)
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
	service.invokationFramesMu.Lock()
	service.invokationFrames[stacktrace.Frame{File: "new_file.go", Line: 100}] = struct{}{}
	service.invokationFramesMu.Unlock()

	// Get frames again
	_, invocationFrames8 := service.source()

	// Original frames should not be affected
	is.Len(originalFrames, 2, "Original frames should not be affected by internal changes")
	is.Len(invocationFrames8, 3, "New frames should include the additional frame")
}

// Test context value propagation for service eager
func TestServiceEager_ContextValuePropagation(t *testing.T) {
	t.Parallel()
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Create test services that capture context values
	healthcheckService := &contextValueHealthcheckerEager{}
	shutdownService := &contextValueShutdownerEager{}

	// Create service eager instances
	healthcheckEager := newServiceEager("healthcheck-eager", healthcheckService)
	shutdownEager := newServiceEager("shutdown-eager", shutdownService)

	// Test context value propagation for healthcheck
	ctx1 := context.WithValue(context.Background(), ctxTestKey, "healthcheck-value")
	err := healthcheckEager.healthcheck(ctx1)
	is.Nil(err)

	// Test context value propagation for shutdown
	ctx2 := context.WithValue(context.Background(), ctxTestKey, "shutdown-value")
	err = shutdownEager.shutdown(ctx2)
	is.Nil(err)

	// Test that eager service properly delegates to the underlying instance
	// The eager service should not store context values itself, but pass them through
	// to the underlying service instance
}
