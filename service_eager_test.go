package do

import (
	"context"
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

func TestNewServiceEager(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	// @TODO
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

// @TODO: missing tests for context
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

// @TODO: missing tests for context
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
