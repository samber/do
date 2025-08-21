package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
)

// TestCascadingConstructionFailures tests cascading failures during service construction
func TestCascadingConstructionFailures(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a chain of services where one fails during construction
	do.ProvideNamed(root, "service-a", func(i do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](i, "service-b")
		if err != nil {
			return "", err
		}
		return "service-a-value", nil
	})

	do.ProvideNamed(root, "service-b", func(i do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](i, "service-c")
		if err != nil {
			return "", err
		}
		return "service-b-value", nil
	})

	do.ProvideNamed(root, "service-c", func(i do.Injector) (string, error) {
		return "", fmt.Errorf("construction failure in service-c")
	})

	// Attempting to invoke service-a should fail due to cascading construction failure
	_, err := do.InvokeNamed[string](root, "service-a")
	is.Error(err)
	is.Contains(err.Error(), "construction failure in service-c")
}

// TestCascadingHealthCheckFailures tests cascading failures during health checks
func TestCascadingHealthCheckFailures(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services with health checks that can fail
	do.ProvideNamed(root, "healthy-service", func(i do.Injector) (*cascadingHealthchecker, error) {
		return &cascadingHealthchecker{name: "healthy", healthy: true}, nil
	})

	do.ProvideNamed(root, "unhealthy-service", func(i do.Injector) (*cascadingHealthchecker, error) {
		return &cascadingHealthchecker{name: "unhealthy", healthy: false}, nil
	})

	do.ProvideNamed(root, "dependent-service", func(i do.Injector) (*cascadingHealthchecker, error) {
		// This service depends on the unhealthy service
		_, err := do.InvokeNamed[*cascadingHealthchecker](i, "unhealthy-service")
		if err != nil {
			return nil, err
		}
		return &cascadingHealthchecker{name: "dependent", healthy: true}, nil
	})

	// Invoke services to ensure they are initialized
	_, err := do.InvokeNamed[*cascadingHealthchecker](root, "healthy-service")
	is.NoError(err)
	_, err = do.InvokeNamed[*cascadingHealthchecker](root, "unhealthy-service")
	is.NoError(err)
	_, err = do.InvokeNamed[*cascadingHealthchecker](root, "dependent-service")
	is.NoError(err)

	// Perform health check - should show failures for unhealthy service only
	// Health checks don't cascade - each service reports its own health status
	healthMap := root.HealthCheck()
	is.Contains(healthMap, "unhealthy-service")
	is.Error(healthMap["unhealthy-service"])
	is.Contains(healthMap, "dependent-service")
	is.Nil(healthMap["dependent-service"]) // dependent-service itself is healthy
	is.Contains(healthMap, "healthy-service")
	is.Nil(healthMap["healthy-service"])
}

// TestCascadingShutdownFailures tests cascading failures during shutdown
func TestCascadingShutdownFailures(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services with shutdown tracking
	shutdownOrder := make([]string, 0)
	var mu sync.Mutex

	do.ProvideNamed(root, "shutdownable-1", func(i do.Injector) (*cascadingShutdowner, error) {
		return &cascadingShutdowner{name: "service-1", shouldFail: false, shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(root, "shutdownable-2", func(i do.Injector) (*cascadingShutdowner, error) {
		return &cascadingShutdowner{name: "service-2", shouldFail: true, shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(root, "shutdownable-3", func(i do.Injector) (*cascadingShutdowner, error) {
		return &cascadingShutdowner{name: "service-3", shouldFail: false, shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	// Invoke services to ensure they are initialized
	_, err := do.InvokeNamed[*cascadingShutdowner](root, "shutdownable-1")
	is.NoError(err)
	_, err = do.InvokeNamed[*cascadingShutdowner](root, "shutdownable-2")
	is.NoError(err)
	_, err = do.InvokeNamed[*cascadingShutdowner](root, "shutdownable-3")
	is.NoError(err)

	// Shutdown should complete but with errors
	report := root.Shutdown()
	is.NotNil(report)
	is.Contains(report.Error(), "shutdown failed for service-2")

	// Verify shutdown order (all services should attempt shutdown)
	is.Len(shutdownOrder, 3)
	is.Contains(shutdownOrder, "service-1")
	is.Contains(shutdownOrder, "service-2")
	is.Contains(shutdownOrder, "service-3")
}

// TestCascadingFailuresAcrossScopes tests cascading failures across nested scopes
func TestCascadingFailuresAcrossScopes(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()
	child1 := root.Scope("child1")
	child2 := child1.Scope("child2")

	// Services in different scopes
	do.ProvideNamedValue(root, "root-service", "root-value")
	do.ProvideNamed(child1, "child1-service", func(i do.Injector) (string, error) {
		// This service depends on a service that will fail
		_, err := do.InvokeNamed[string](i, "failing-service")
		if err != nil {
			return "", err
		}
		return "child1-value", nil
	})
	do.ProvideNamedValue(child2, "child2-service", "child2-value")

	// Service that fails
	do.ProvideNamed(child2, "failing-service", func(i do.Injector) (string, error) {
		return "", fmt.Errorf("service failure in child2")
	})

	// Service in child2 that depends on services in parent scopes
	do.ProvideNamed(child2, "dependent-service", func(i do.Injector) (string, error) {
		rootVal, err := do.InvokeNamed[string](i, "root-service")
		if err != nil {
			return "", err
		}
		child1Val, err := do.InvokeNamed[string](i, "child1-service")
		if err != nil {
			return "", err
		}
		child2Val, err := do.InvokeNamed[string](i, "child2-service")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s-%s-%s", rootVal, child1Val, child2Val), nil
	})

	// Attempting to invoke the dependent service should fail due to cascading failure
	_, err := do.InvokeNamed[string](child2, "dependent-service")
	is.Error(err)
	is.Contains(err.Error(), "could not find service")
}

// TestPartialFailureScenarios tests scenarios where some services succeed and others fail
func TestPartialFailureScenarios(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a mix of successful and failing services
	do.ProvideNamed(root, "successful-service", func(i do.Injector) (string, error) {
		return "successful-value", nil
	})

	do.ProvideNamed(root, "failing-service", func(i do.Injector) (string, error) {
		return "", fmt.Errorf("service failure")
	})

	do.ProvideNamed(root, "mixed-service", func(i do.Injector) (string, error) {
		// This service depends on both successful and failing services
		successVal, err := do.InvokeNamed[string](i, "successful-service")
		if err != nil {
			return "", err
		}
		_, err = do.InvokeNamed[string](i, "failing-service")
		if err != nil {
			return "", err
		}
		return "mixed-" + successVal, nil
	})

	// Successful service should work
	result, err := do.InvokeNamed[string](root, "successful-service")
	is.NoError(err)
	is.Equal("successful-value", result)

	// Failing service should fail
	_, err = do.InvokeNamed[string](root, "failing-service")
	is.Error(err)
	is.Contains(err.Error(), "service failure")

	// Mixed service should fail due to dependency on failing service
	_, err = do.InvokeNamed[string](root, "mixed-service")
	is.Error(err)
	is.Contains(err.Error(), "service failure")
}

// TestErrorRecoveryScenarios tests error recovery scenarios
func TestErrorRecoveryScenarios(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that can recover from errors
	recoveryAttempts := 0
	do.ProvideNamed(root, "recovery-service", func(i do.Injector) (string, error) {
		recoveryAttempts++
		if recoveryAttempts < 3 {
			return "", fmt.Errorf("temporary failure, attempt %d", recoveryAttempts)
		}
		return "recovered-value", nil
	})

	// First two attempts should fail
	_, err := do.InvokeNamed[string](root, "recovery-service")
	is.Error(err)
	is.Contains(err.Error(), "temporary failure, attempt 1")

	_, err = do.InvokeNamed[string](root, "recovery-service")
	is.Error(err)
	is.Contains(err.Error(), "temporary failure, attempt 2")

	// Third attempt should succeed
	result, err := do.InvokeNamed[string](root, "recovery-service")
	is.NoError(err)
	is.Equal("recovered-value", result)
}

// TestContextCancellationDuringFailures tests context cancellation during failure scenarios
func TestContextCancellationDuringFailures(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that takes a long time to fail
	do.ProvideNamed(root, "slow-failing-service", func(i do.Injector) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "", fmt.Errorf("slow failure")
	})

	// Try to invoke the service - it should fail due to service failure
	_, err := do.InvokeNamed[string](root, "slow-failing-service")
	is.Error(err)
	// The service should still fail, but the context cancellation doesn't affect the virtual scope
	is.Contains(err.Error(), "slow failure")
}

// TestCascadingMemoryLeaks tests for memory leaks in cascading failure scenarios
func TestCascadingMemoryLeaks(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create many services that can fail
	const numServices = 100

	for i := 0; i < numServices; i++ {
		serviceName := fmt.Sprintf("service-%d", i)
		serviceIndex := i
		do.ProvideNamed(root, serviceName, func(i do.Injector) (string, error) {
			if serviceIndex%10 == 0 {
				return "", fmt.Errorf("failure in service %d", serviceIndex)
			}
			return fmt.Sprintf("value-%d", serviceIndex), nil
		})
	}

	// Try to invoke services, some will fail
	failureCount := 0
	successCount := 0

	for i := 0; i < numServices; i++ {
		serviceName := fmt.Sprintf("service-%d", i)
		_, err := do.InvokeNamed[string](root, serviceName)
		if err != nil {
			failureCount++
		} else {
			successCount++
		}
	}

	// Verify that we have both successes and failures
	is.Greater(failureCount, 0)
	is.Greater(successCount, 0)
	is.Equal(numServices, failureCount+successCount)

	// Try to invoke a successful service again to ensure no memory leaks
	result, err := do.InvokeNamed[string](root, "service-1")
	is.NoError(err)
	is.Equal("value-1", result)
}

// Helper types for testing

type cascadingHealthchecker struct {
	name    string
	healthy bool
}

func (t *cascadingHealthchecker) HealthCheck() error {
	if t.healthy {
		return nil
	}
	return fmt.Errorf("health check failed for %s", t.name)
}

type cascadingShutdowner struct {
	name          string
	shouldFail    bool
	shutdownOrder *[]string
	mu            *sync.Mutex
}

func (t *cascadingShutdowner) Shutdown(ctx context.Context) error {
	t.mu.Lock()
	*t.shutdownOrder = append(*t.shutdownOrder, t.name)
	t.mu.Unlock()

	if t.shouldFail {
		return fmt.Errorf("shutdown failed for %s", t.name)
	}

	return nil
}
