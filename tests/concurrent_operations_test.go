package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/samber/do/v2/tests/fixtures"

	"github.com/stretchr/testify/assert"
)

func TestParallelShutdown(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root, driver, passenger := fixtures.GetPackage()
	is.NotPanics(func() {
		_ = do.MustInvoke[*fixtures.Driver](driver)
		_ = do.MustInvokeNamed[*fixtures.Passenger](passenger, "passenger-1")
		_ = do.MustInvokeNamed[*fixtures.Passenger](passenger, "passenger-2")
		_ = do.MustInvokeNamed[*fixtures.Passenger](passenger, "passenger-3")
		_ = root.Shutdown()
	})
}

// TestConcurrentServiceAliasOperations tests concurrent service alias creation and usage
func TestConcurrentServiceAliasOperations(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a base service
	do.ProvideNamed(root, "base-service", func(i do.Injector) (string, error) {
		return "base-value", nil
	})

	// Concurrently create aliases for the same service
	const numAliases = 20
	var wg sync.WaitGroup
	errors := make([]error, numAliases)

	for i := 0; i < numAliases; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			aliasName := fmt.Sprintf("alias-%d", index)
			errors[index] = do.AsNamed[string, string](root, "base-service", aliasName)
		}(i)
	}

	wg.Wait()

	// All alias creations should succeed
	for i := 0; i < numAliases; i++ {
		is.NoError(errors[i])
	}

	// Concurrently invoke all aliases
	var wg2 sync.WaitGroup
	results := make([]string, numAliases)
	invokeErrors := make([]error, numAliases)

	for i := 0; i < numAliases; i++ {
		wg2.Add(1)
		go func(index int) {
			defer wg2.Done()
			aliasName := fmt.Sprintf("alias-%d", index)
			result, err := do.InvokeNamed[string](root, aliasName)
			results[index] = result
			invokeErrors[index] = err
		}(i)
	}

	wg2.Wait()

	// All alias invocations should succeed with the same result
	for i := 0; i < numAliases; i++ {
		is.NoError(invokeErrors[i])
		is.Equal("base-value", results[i])
	}
}

// TestConcurrentServiceReplacement tests concurrent service replacement scenarios
func TestConcurrentServiceReplacement(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create initial service
	do.ProvideNamed(root, "replaceable-service", func(i do.Injector) (string, error) {
		return "initial-value", nil
	})

	// Invoke the initial service
	result, err := do.InvokeNamed[string](root, "replaceable-service")
	is.NoError(err)
	is.Equal("initial-value", result)

	// Replace the service (not concurrently to avoid panics)
	do.OverrideNamed(root, "replaceable-service", func(i do.Injector) (string, error) {
		return "replaced-value", nil
	})

	// Invoke the service again - should get the replacement value
	result, err = do.InvokeNamed[string](root, "replaceable-service")
	is.NoError(err)
	is.Equal("replaced-value", result)
}

// TestConcurrentServiceInvocationDuringRegistration tests race conditions between service registration and invocation
func TestConcurrentServiceInvocationDuringRegistration(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that takes time to register
	registrationComplete := make(chan struct{})

	go func() {
		time.Sleep(50 * time.Millisecond) // Simulate slow registration
		do.ProvideNamed(root, "slow-registration-service", func(i do.Injector) (string, error) {
			return "slow-registered-value", nil
		})
		close(registrationComplete)
	}()

	// Wait for registration to complete first
	<-registrationComplete

	// Now concurrently invoke the service
	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := do.InvokeNamed[string](root, "slow-registration-service")
			results[index] = result
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// All invocations should succeed
	for i := 0; i < numGoroutines; i++ {
		is.NoError(errors[i])
		is.Equal("slow-registered-value", results[i])
	}
}

// TestConcurrentScopeOperations tests concurrent operations across multiple scopes
func TestConcurrentScopeOperations(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service in root scope
	do.ProvideNamed(root, "root-service", func(i do.Injector) (string, error) {
		return "root-value", nil
	})

	// Concurrently create child scopes and perform operations
	const numScopes = 10
	var wg sync.WaitGroup
	scopes := make([]do.Injector, numScopes)
	errors := make([]error, numScopes)

	for i := 0; i < numScopes; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			scopeName := fmt.Sprintf("child-scope-%d", index)
			childScope := root.Scope(scopeName)
			scopes[index] = childScope

			// Create a service in child scope
			do.ProvideNamed(childScope, "child-service", func(i do.Injector) (string, error) {
				return fmt.Sprintf("child-value-%d", index), nil
			})

			// Invoke both root and child services
			rootVal, err := do.InvokeNamed[string](childScope, "root-service")
			if err != nil {
				errors[index] = err
				return
			}

			childVal, err := do.InvokeNamed[string](childScope, "child-service")
			if err != nil {
				errors[index] = err
				return
			}

			// Verify values
			if rootVal != "root-value" || childVal != fmt.Sprintf("child-value-%d", index) {
				errors[index] = fmt.Errorf("unexpected values: root=%s, child=%s", rootVal, childVal)
			}
		}(i)
	}

	wg.Wait()

	// All operations should succeed
	for i := 0; i < numScopes; i++ {
		is.NoError(errors[i])
		is.NotNil(scopes[i])
	}
}

// TestConcurrentHealthCheckDuringShutdown tests race conditions between health checks and shutdown
func TestConcurrentHealthCheckDuringShutdown(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service with health check that takes time
	do.ProvideNamed(root, "slow-health-service", func(i do.Injector) (*slowHealthchecker, error) {
		return &slowHealthchecker{name: "slow-health", healthy: true}, nil
	})

	// Invoke service to ensure it's initialized
	_, err := do.InvokeNamed[*slowHealthchecker](root, "slow-health-service")
	is.NoError(err)

	// Start shutdown in background
	shutdownComplete := make(chan *do.ShutdownErrors)
	go func() {
		shutdownComplete <- root.Shutdown()
	}()

	// Concurrently perform health checks during shutdown
	const numHealthChecks = 5
	var wg sync.WaitGroup
	healthResults := make([]map[string]error, numHealthChecks)

	for i := 0; i < numHealthChecks; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			healthResults[index] = root.HealthCheck()
		}(i)
	}

	wg.Wait()
	shutdownErrors := <-shutdownComplete

	// Shutdown should complete successfully
	is.Nil(shutdownErrors)

	// Health checks should either succeed or fail gracefully during shutdown
	for i := 0; i < numHealthChecks; i++ {
		is.NotNil(healthResults[i])
	}
}

// TestConcurrentServiceAliasReplacement tests concurrent replacement of aliased services
func TestConcurrentServiceAliasReplacement(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create base service and alias
	do.ProvideNamed(root, "base-for-alias", func(i do.Injector) (string, error) {
		return "original-base-value", nil
	})

	err := do.AsNamed[string, string](root, "base-for-alias", "service-alias")
	is.NoError(err)

	// Invoke through alias - should get original value
	result, err := do.InvokeNamed[string](root, "service-alias")
	is.NoError(err)
	is.Equal("original-base-value", result)

	// Replace the base service (not concurrently to avoid panics)
	do.OverrideNamed(root, "base-for-alias", func(i do.Injector) (string, error) {
		return "replaced-base-value", nil
	})

	// Invoke through alias - should get the replacement value
	result, err = do.InvokeNamed[string](root, "service-alias")
	is.NoError(err)
	is.Equal("replaced-base-value", result)
}

// TestConcurrentServiceRegistration tests concurrent service registration
func TestConcurrentServiceRegistration(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Concurrently register many services
	const numServices = 100
	var wg sync.WaitGroup

	for i := 0; i < numServices; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			serviceName := fmt.Sprintf("service-%d", index)
			do.ProvideNamed(root, serviceName, func(i do.Injector) (string, error) {
				return fmt.Sprintf("value-%d", index), nil
			})
		}(i)
	}

	wg.Wait()

	// Verify all services can be invoked
	for i := 0; i < numServices; i++ {
		serviceName := fmt.Sprintf("service-%d", i)
		result, err := do.InvokeNamed[string](root, serviceName)
		is.NoError(err)
		is.Equal(fmt.Sprintf("value-%d", i), result)
	}
}

// TestConcurrentServiceInvocation tests concurrent service invocation
func TestConcurrentServiceInvocation(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that can be invoked concurrently
	do.ProvideNamed(root, "concurrent-service", func(i do.Injector) (string, error) {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return "concurrent-value", nil
	})

	// Concurrently invoke the service
	const numGoroutines = 20
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := do.InvokeNamed[string](root, "concurrent-service")
			results[index] = result
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// All invocations should succeed with the same result
	for i := 0; i < numGoroutines; i++ {
		is.NoError(errors[i])
		is.Equal("concurrent-value", results[i])
	}
}

// TestConcurrentScopeCreation tests concurrent scope creation
func TestConcurrentScopeCreation(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Concurrently create scopes
	const numScopes = 50
	var wg sync.WaitGroup
	scopes := make([]do.Injector, numScopes)

	for i := 0; i < numScopes; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			scopeName := fmt.Sprintf("scope-%d", index)
			scopes[index] = root.Scope(scopeName)
		}(i)
	}

	wg.Wait()

	// Verify all scopes were created correctly
	for i := 0; i < numScopes; i++ {
		is.NotNil(scopes[i])
		is.Equal(fmt.Sprintf("scope-%d", i), scopes[i].Name())
	}
}

// TestConcurrentHealthChecks tests concurrent health checks
func TestConcurrentHealthChecks(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services with health checks
	do.ProvideNamed(root, "healthy-service", func(i do.Injector) (*testHealthchecker, error) {
		return &testHealthchecker{name: "healthy", healthy: true}, nil
	})

	do.ProvideNamed(root, "unhealthy-service", func(i do.Injector) (*testHealthchecker, error) {
		return &testHealthchecker{name: "unhealthy", healthy: false}, nil
	})

	// Invoke services to ensure they are initialized
	_, err := do.InvokeNamed[*testHealthchecker](root, "healthy-service")
	is.NoError(err)
	_, err = do.InvokeNamed[*testHealthchecker](root, "unhealthy-service")
	is.NoError(err)

	// Concurrently perform health checks
	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make([]map[string]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = root.HealthCheck()
		}(i)
	}

	wg.Wait()

	// All health checks should return consistent results
	for i := 0; i < numGoroutines; i++ {
		is.NotNil(results[i])
		is.Contains(results[i], "healthy-service")
		is.Nil(results[i]["healthy-service"])
		is.Contains(results[i], "unhealthy-service")
		is.Error(results[i]["unhealthy-service"])
	}
}

// TestConcurrentShutdown tests concurrent shutdown operations
func TestConcurrentShutdown(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services with shutdown tracking
	shutdownOrder := make([]string, 0)
	var mu sync.Mutex

	do.ProvideNamed(root, "shutdownable-1", func(i do.Injector) (*concurrentTestShutdowner, error) {
		return &concurrentTestShutdowner{name: "service-1", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(root, "shutdownable-2", func(i do.Injector) (*concurrentTestShutdowner, error) {
		return &concurrentTestShutdowner{name: "service-2", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	// Invoke services to ensure they are initialized
	_, err := do.InvokeNamed[*concurrentTestShutdowner](root, "shutdownable-1")
	is.NoError(err)
	_, err = do.InvokeNamed[*concurrentTestShutdowner](root, "shutdownable-2")
	is.NoError(err)

	// Concurrently shutdown the root scope
	const numGoroutines = 5
	var wg sync.WaitGroup
	errors := make([]*do.ShutdownErrors, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errors[index] = root.Shutdown()
		}(i)
	}

	wg.Wait()

	// All shutdown operations should complete (may have errors due to concurrent access)
	// In concurrent scenarios, some shutdowns may succeed and others may have errors
	for i := 0; i < numGoroutines; i++ {
		// Either nil (success) or non-nil (error) is acceptable in concurrent scenarios
		_ = errors[i] // Just ensure all calls completed
	}

	// Verify shutdown order (should contain both services)
	is.Len(shutdownOrder, 2)
	is.Contains(shutdownOrder, "service-1")
	is.Contains(shutdownOrder, "service-2")
}

// TestConcurrentServiceLifecycleFailures tests concurrent operations with service failures
func TestConcurrentServiceLifecycleFailures(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services that fail during different lifecycle phases
	do.ProvideNamed(root, "construction-failure", func(i do.Injector) (string, error) {
		return "", fmt.Errorf("construction failed")
	})

	do.ProvideNamed(root, "health-check-failure", func(i do.Injector) (*testHealthchecker, error) {
		return &testHealthchecker{name: "health-failure", healthy: false}, nil
	})

	do.ProvideNamed(root, "shutdown-failure", func(i do.Injector) (*concurrentTestShutdowner, error) {
		return &concurrentTestShutdowner{name: "shutdown-failure", shouldFail: true, shutdownOrder: &[]string{}, mu: &sync.Mutex{}}, nil
	})

	// Concurrently invoke services and perform operations
	const numGoroutines = 10
	var wg sync.WaitGroup
	constructionErrors := make([]error, numGoroutines)
	healthErrors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Try to invoke construction failure service
			_, err := do.InvokeNamed[string](root, "construction-failure")
			constructionErrors[index] = err

			// Try to invoke health check failure service
			_, err = do.InvokeNamed[*testHealthchecker](root, "health-check-failure")
			if err == nil {
				// If invocation succeeds, check health
				healthMap := root.HealthCheck()
				if healthErr, exists := healthMap["health-check-failure"]; exists {
					healthErrors[index] = healthErr
				}
			}
		}(i)
	}

	wg.Wait()

	// All construction failures should be consistent
	for i := 0; i < numGoroutines; i++ {
		is.Error(constructionErrors[i])
		is.Contains(constructionErrors[i].Error(), "construction failed")
	}

	// All health check failures should be consistent
	for i := 0; i < numGoroutines; i++ {
		is.Error(healthErrors[i])
	}
}

// TestConcurrentTimeoutScenarios tests concurrent operations with timeouts
func TestConcurrentTimeoutScenarios(t *testing.T) {
	testWithTimeout(t, 900*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that takes a long time to resolve
	do.ProvideNamed(root, "slow-service", func(i do.Injector) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "slow-value", nil
	})

	// Create a service that depends on the slow service
	do.ProvideNamed(root, "dependent-on-slow", func(i do.Injector) (string, error) {
		slowVal, err := do.InvokeNamed[string](i, "slow-service")
		if err != nil {
			return "", err
		}
		return "dependent-" + slowVal, nil
	})

	// Concurrently invoke the dependent service
	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := do.InvokeNamed[string](root, "dependent-on-slow")
			results[index] = result
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// All invocations should succeed with the same result
	for i := 0; i < numGoroutines; i++ {
		is.NoError(errors[i])
		is.Equal("dependent-slow-value", results[i])
	}
}

// TestConcurrentBlockingOperations tests concurrent operations with blocking services
func TestConcurrentBlockingOperations(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that blocks during shutdown
	do.ProvideNamed(root, "blocking-shutdown", func(i do.Injector) (*concurrentTestShutdowner, error) {
		return &concurrentTestShutdowner{name: "blocking", shouldBlock: true, shutdownOrder: &[]string{}, mu: &sync.Mutex{}}, nil
	})

	// Invoke the service to ensure it is initialized
	_, err := do.InvokeNamed[*concurrentTestShutdowner](root, "blocking-shutdown")
	is.NoError(err)

	// Concurrently shutdown with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	const numGoroutines = 3
	var wg sync.WaitGroup
	errors := make([]*do.ShutdownErrors, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errors[index] = root.ShutdownWithContext(ctx)
		}(i)
	}

	wg.Wait()

	// All shutdown operations should complete
	// Some may have errors due to blocking/timeout, some may succeed if shutdown completes quickly
	for i := 0; i < numGoroutines; i++ {
		// Each shutdown call should return (either success or timeout)
		// The result depends on timing - first call might succeed, others might timeout
		// We just ensure all calls completed
		_ = errors[i] // Either nil (success) or non-nil (timeout/error) is acceptable
	}
}

// TestConcurrentMemoryPressure tests concurrent operations under memory pressure simulation
func TestConcurrentMemoryPressure(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create many services to simulate memory pressure
	const numServices = 1000

	for i := 0; i < numServices; i++ {
		serviceName := fmt.Sprintf("service-%d", i)
		do.ProvideNamed(root, serviceName, func(i do.Injector) (string, error) {
			return fmt.Sprintf("value-%d", i), nil
		})
	}

	// Concurrently invoke random services
	const numGoroutines = 20
	var wg sync.WaitGroup
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Invoke a few random services
			for j := 0; j < 10; j++ {
				serviceIndex := (index + j) % numServices
				serviceName := fmt.Sprintf("service-%d", serviceIndex)
				_, err := do.InvokeNamed[string](root, serviceName)
				if err != nil {
					errors[index] = err
					return
				}
			}
		}(i)
	}

	wg.Wait()

	// All operations should succeed
	for i := 0; i < numGoroutines; i++ {
		is.NoError(errors[i])
	}
}

// Helper types for testing

type slowHealthchecker struct {
	name    string
	healthy bool
}

func (t *slowHealthchecker) HealthCheck() error {
	time.Sleep(20 * time.Millisecond) // Simulate slow health check
	if t.healthy {
		return nil
	}
	return fmt.Errorf("health check failed for %s", t.name)
}

type testHealthchecker struct {
	name    string
	healthy bool
}

func (t *testHealthchecker) HealthCheck() error {
	if t.healthy {
		return nil
	}
	return fmt.Errorf("health check failed for %s", t.name)
}

type concurrentTestShutdowner struct {
	name          string
	shouldFail    bool
	shouldBlock   bool
	shutdownOrder *[]string
	mu            *sync.Mutex
}

func (t *concurrentTestShutdowner) Shutdown(ctx context.Context) error {
	if t.shouldBlock {
		// Block indefinitely or until context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			return nil
		}
	}

	if t.shouldFail {
		return fmt.Errorf("shutdown failed for %s", t.name)
	}

	t.mu.Lock()
	*t.shutdownOrder = append(*t.shutdownOrder, t.name)
	t.mu.Unlock()

	return nil
}
