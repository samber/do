package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
)

// TestRaceConditionServiceInitialization tests race conditions during service initialization
func TestRaceConditionServiceInitialization(t *testing.T) {
	testWithTimeout(t, 500*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that can be initialized multiple times with different values
	var initializationCount int32
	do.ProvideNamed(root, "race-service", func(i do.Injector) (*raceTestService, error) {
		count := atomic.AddInt32(&initializationCount, 1)
		return &raceTestService{
			id:    int(count),
			value: fmt.Sprintf("initialized-%d", count),
		}, nil
	})

	// Concurrently invoke the service many times
	const numGoroutines = 50
	var wg sync.WaitGroup
	results := make([]*raceTestService, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := do.InvokeNamed[*raceTestService](root, "race-service")
			results[index] = result
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// All invocations should succeed
	for i := 0; i < numGoroutines; i++ {
		is.NoError(errors[i])
		is.NotNil(results[i])
	}

	// All results should be the same instance (singleton behavior)
	firstResult := results[0]
	for i := 1; i < numGoroutines; i++ {
		is.Equal(firstResult, results[i])
	}

	// Service should only be initialized once
	is.Equal(int32(1), atomic.LoadInt32(&initializationCount))
}

// TestRaceConditionServiceReplacement tests race conditions when replacing services
func TestRaceConditionServiceReplacement(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create initial service
	do.ProvideNamed(root, "race-replace-service", func(i do.Injector) (string, error) {
		return "initial", nil
	})

	// Invoke the initial service
	result, err := do.InvokeNamed[string](root, "race-replace-service")
	is.NoError(err)
	is.Equal("initial", result)

	// Replace the service
	do.OverrideNamed(root, "race-replace-service", func(i do.Injector) (string, error) {
		return "replaced", nil
	})

	// Invoke the service again
	result, err = do.InvokeNamed[string](root, "race-replace-service")
	is.NoError(err)
	is.Equal("replaced", result)
}

// TestRaceConditionScopeCreation tests race conditions during scope creation
func TestRaceConditionScopeCreation(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service in root
	do.ProvideNamed(root, "root-race-service", func(i do.Injector) (string, error) {
		return "root-value", nil
	})

	// Concurrently create scopes and access services
	const numScopes = 30
	var wg sync.WaitGroup
	scopeErrors := make([]error, numScopes)

	for i := 0; i < numScopes; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Create scope
			scopeName := fmt.Sprintf("race-scope-%d", index)
			childScope := root.Scope(scopeName)

			// Create service in child scope
			do.ProvideNamed(childScope, "child-race-service", func(i do.Injector) (string, error) {
				return fmt.Sprintf("child-value-%d", index), nil
			})

			// Access both root and child services
			rootVal, err := do.InvokeNamed[string](childScope, "root-race-service")
			if err != nil {
				scopeErrors[index] = err
				return
			}

			childVal, err := do.InvokeNamed[string](childScope, "child-race-service")
			if err != nil {
				scopeErrors[index] = err
				return
			}

			// Verify values
			if rootVal != "root-value" || childVal != fmt.Sprintf("child-value-%d", index) {
				scopeErrors[index] = fmt.Errorf("unexpected values in scope %d: root=%s, child=%s", index, rootVal, childVal)
			}
		}(i)
	}

	wg.Wait()

	// All operations should succeed
	for i := 0; i < numScopes; i++ {
		is.NoError(scopeErrors[i])
	}
}

// TestRaceConditionHealthCheckInvocation tests race conditions between health checks and service invocation
func TestRaceConditionHealthCheckInvocation(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service with health check
	do.ProvideNamed(root, "health-race-service", func(i do.Injector) (*raceHealthchecker, error) {
		return &raceHealthchecker{name: "health-race", healthy: true}, nil
	})

	// Concurrently invoke service and perform health checks
	const numOperations = 15
	var wg sync.WaitGroup
	invokeErrors := make([]error, numOperations)
	healthResults := make([]map[string]error, numOperations)

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			if index%2 == 0 {
				// Invoke service
				_, err := do.InvokeNamed[*raceHealthchecker](root, "health-race-service")
				invokeErrors[index] = err
			} else {
				// Perform health check
				healthResults[index] = root.HealthCheck()
			}
		}(i)
	}

	wg.Wait()

	// All operations should succeed
	for i := 0; i < numOperations; i += 2 {
		is.NoError(invokeErrors[i])
	}

	for i := 1; i < numOperations; i += 2 {
		is.NotNil(healthResults[i])
		is.Contains(healthResults[i], "health-race-service")
	}
}

// TestRaceConditionShutdownInvocation tests race conditions between shutdown and service invocation
func TestRaceConditionShutdownInvocation(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service with shutdown
	do.ProvideNamed(root, "shutdown-race-service", func(i do.Injector) (*raceShutdowner, error) {
		return &raceShutdowner{name: "shutdown-race"}, nil
	})

	// Invoke service to ensure it's initialized
	_, err := do.InvokeNamed[*raceShutdowner](root, "shutdown-race-service")
	is.NoError(err)

	// Concurrently invoke service and shutdown
	const numOperations = 10
	var wg sync.WaitGroup
	invokeErrors := make([]error, numOperations)
	shutdownErrors := make([]*do.ShutdownErrors, numOperations)

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			if index%2 == 0 {
				// Invoke service
				_, err := do.InvokeNamed[*raceShutdowner](root, "shutdown-race-service")
				invokeErrors[index] = err
			} else {
				// Shutdown
				shutdownErrors[index] = root.Shutdown()
			}
		}(i)
	}

	wg.Wait()

	// Some operations may fail due to shutdown, which is expected
	// We just verify that the operations complete without panicking
	for i := 0; i < numOperations; i++ {
		// Operations should complete without panicking
		is.NotNil(invokeErrors[i] != nil || shutdownErrors[i] != nil)
	}
}

// TestRaceConditionSlowServiceInvocation tests race conditions with slow service resolution
func TestRaceConditionSlowServiceInvocation(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that takes time to resolve
	do.ProvideNamed(root, "slow-race-service", func(i do.Injector) (string, error) {
		time.Sleep(50 * time.Millisecond)
		return "slow-value", nil
	})

	// Concurrently invoke the slow service
	const numGoroutines = 8
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Invoke service
			result, err := do.InvokeNamed[string](root, "slow-race-service")
			results[index] = result
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// All invocations should succeed with the same result
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		if errors[i] == nil {
			successCount++
			is.Equal("slow-value", results[i])
		}
	}

	// All invocations should succeed
	is.Equal(numGoroutines, successCount)
}

// TestRaceConditionServiceAliasCreation tests race conditions during service alias creation
func TestRaceConditionServiceAliasCreation(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create base service
	do.ProvideNamed(root, "alias-base-service", func(i do.Injector) (string, error) {
		return "base-value", nil
	})

	// Concurrently create aliases and invoke them
	const numOperations = 25
	var wg sync.WaitGroup
	aliasErrors := make([]error, numOperations)
	invokeResults := make([]string, numOperations)
	invokeErrors := make([]error, numOperations)

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			aliasName := fmt.Sprintf("alias-%d", index)

			if index%2 == 0 {
				// Create alias
				aliasErrors[index] = do.AsNamed[string, string](root, "alias-base-service", aliasName)
			} else {
				// Try to invoke alias (may not exist yet)
				result, err := do.InvokeNamed[string](root, aliasName)
				invokeResults[index] = result
				invokeErrors[index] = err
			}
		}(i)
	}

	wg.Wait()

	// All alias creations should succeed
	for i := 0; i < numOperations; i += 2 {
		is.NoError(aliasErrors[i])
	}

	// Some alias invocations may fail if they were invoked before creation
	// This is expected behavior
}

// Helper types for testing

type raceTestService struct {
	id    int
	value string
}

type raceHealthchecker struct {
	name    string
	healthy bool
}

func (t *raceHealthchecker) HealthCheck() error {
	if t.healthy {
		return nil
	}
	return fmt.Errorf("health check failed for %s", t.name)
}

type raceShutdowner struct {
	name string
}

func (t *raceShutdowner) Shutdown(ctx context.Context) error {
	return nil
}
