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

// TestMixedServiceTypesComplexDependencies tests complex dependencies between different service types
func TestMixedServiceTypesComplexDependencies(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services of different types with complex dependencies
	// Eager service (immediate initialization)
	do.ProvideNamedValue(root, "eager-base", "eager-base-value")

	// Lazy service that depends on eager service
	do.ProvideNamed(root, "lazy-dependent", func(i do.Injector) (string, error) {
		eagerVal, err := do.InvokeNamed[string](i, "eager-base")
		if err != nil {
			return "", err
		}
		return "lazy-" + eagerVal, nil
	})

	// Transient service that depends on lazy service
	do.ProvideNamedTransient(root, "transient-dependent", func(i do.Injector) (string, error) {
		lazyVal, err := do.InvokeNamed[string](i, "lazy-dependent")
		if err != nil {
			return "", err
		}
		return "transient-" + lazyVal + "-" + fmt.Sprintf("%d", time.Now().UnixNano()), nil
	})

	// Alias service that points to transient service
	err := do.AsNamed[string, string](root, "transient-dependent", "alias-service")
	is.NoError(err)

	// Invoke the alias service multiple times
	result1, err := do.InvokeNamed[string](root, "alias-service")
	is.NoError(err)
	is.Contains(result1, "transient-lazy-eager-base-value")

	result2, err := do.InvokeNamed[string](root, "alias-service")
	is.NoError(err)
	is.Contains(result2, "transient-lazy-eager-base-value")

	// Transient services should return different values (due to timestamp)
	is.NotEqual(result1, result2)
}

// TestMixedServiceTypesCircularDependencies tests circular dependencies between different service types
func TestMixedServiceTypesCircularDependencies(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a lazy service that depends on an eager service
	do.ProvideNamed(root, "lazy-service", func(i do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](i, "eager-service")
		if err != nil {
			return "", err
		}
		return "lazy-value", nil
	})

	// Create an eager service that depends on the lazy service (circular dependency)
	do.ProvideNamed(root, "eager-service", func(i do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](i, "lazy-service")
		if err != nil {
			return "", err
		}
		return "eager-value", nil
	})

	// Attempting to invoke should detect circular dependency
	_, err := do.InvokeNamed[string](root, "lazy-service")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestMixedServiceTypesConcurrentAccess tests concurrent access to different service types
func TestMixedServiceTypesConcurrentAccess(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services of different types
	do.ProvideNamedValue(root, "eager-shared", "eager-shared-value")

	do.ProvideNamed(root, "lazy-shared", func(i do.Injector) (string, error) {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return "lazy-shared-value", nil
	})

	do.ProvideNamedTransient(root, "transient-shared", func(i do.Injector) (string, error) {
		return "transient-shared-value", nil
	})

	// Concurrently access different service types
	const numGoroutines = 10
	var wg sync.WaitGroup
	eagerResults := make([]string, numGoroutines)
	lazyResults := make([]string, numGoroutines)
	transientResults := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			eagerVal, err := do.InvokeNamed[string](root, "eager-shared")
			if err != nil {
				errors[index] = err
				return
			}
			eagerResults[index] = eagerVal

			lazyVal, err := do.InvokeNamed[string](root, "lazy-shared")
			if err != nil {
				errors[index] = err
				return
			}
			lazyResults[index] = lazyVal

			transientVal, err := do.InvokeNamed[string](root, "transient-shared")
			if err != nil {
				errors[index] = err
				return
			}
			transientResults[index] = transientVal
		}(i)
	}

	wg.Wait()

	// All operations should succeed with consistent results
	for i := 0; i < numGoroutines; i++ {
		is.NoError(errors[i])
		is.Equal("eager-shared-value", eagerResults[i])
		is.Equal("lazy-shared-value", lazyResults[i])
		is.Equal("transient-shared-value", transientResults[i])
	}
}

// TestMixedServiceTypesLifecycle tests lifecycle management with different service types
func TestMixedServiceTypesLifecycle(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services of different types with lifecycle tracking
	shutdownOrder := make([]string, 0)
	var mu sync.Mutex

	do.ProvideNamedValue(root, "eager-lifecycle", &mixedLifecycleService{name: "eager", shutdownOrder: &shutdownOrder, mu: &mu})

	do.ProvideNamed(root, "lazy-lifecycle", func(i do.Injector) (*mixedLifecycleService, error) {
		return &mixedLifecycleService{name: "lazy", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamedTransient(root, "transient-lifecycle", func(i do.Injector) (*mixedLifecycleService, error) {
		return &mixedLifecycleService{name: "transient", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	// Invoke services to ensure they are initialized
	_, err := do.InvokeNamed[*mixedLifecycleService](root, "eager-lifecycle")
	is.NoError(err)
	_, err = do.InvokeNamed[*mixedLifecycleService](root, "lazy-lifecycle")
	is.NoError(err)
	_, err = do.InvokeNamed[*mixedLifecycleService](root, "transient-lifecycle")
	is.NoError(err)

	// Shutdown the root scope
	report := root.Shutdown()
	is.True(report.Succeed)
	is.Empty(report.Errors)

	// Verify shutdown order (eager and lazy services should be shut down, but not transient)
	is.Len(shutdownOrder, 2)
	is.Contains(shutdownOrder, "eager")
	is.Contains(shutdownOrder, "lazy")
	// Transient services are not shutdown because they don't cache instances
}

// TestMixedServiceTypesNestedScopes tests mixed service types in nested scopes
func TestMixedServiceTypesNestedScopes(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()
	child1 := root.Scope("child1")
	child2 := child1.Scope("child2")

	// Services of different types in different scopes
	do.ProvideNamedValue(root, "root-eager", "root-eager-value")

	do.ProvideNamed(child1, "child1-lazy", func(i do.Injector) (string, error) {
		rootVal, err := do.InvokeNamed[string](i, "root-eager")
		if err != nil {
			return "", err
		}
		return "child1-lazy-" + rootVal, nil
	})

	do.ProvideNamedTransient(child2, "child2-transient", func(i do.Injector) (string, error) {
		rootVal, err := do.InvokeNamed[string](i, "root-eager")
		if err != nil {
			return "", err
		}
		child1Val, err := do.InvokeNamed[string](i, "child1-lazy")
		if err != nil {
			return "", err
		}
		return "child2-transient-" + rootVal + "-" + child1Val, nil
	})

	// Alias in child2 that points to child1 service
	err := do.AsNamed[string, string](child2, "child1-lazy", "child2-alias")
	is.NoError(err)

	// Invoke the transient service
	result, err := do.InvokeNamed[string](child2, "child2-transient")
	is.NoError(err)
	is.Contains(result, "child2-transient-root-eager-value-child1-lazy-root-eager-value")

	// Invoke the alias service
	aliasResult, err := do.InvokeNamed[string](child2, "child2-alias")
	is.NoError(err)
	is.Equal("child1-lazy-root-eager-value", aliasResult)
}

// TestMixedServiceTypesPerformance tests performance characteristics of different service types
func TestMixedServiceTypesPerformance(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services of different types
	do.ProvideNamedValue(root, "eager-perf", "eager-value")

	do.ProvideNamed(root, "lazy-perf", func(i do.Injector) (string, error) {
		time.Sleep(1 * time.Millisecond) // Simulate initialization cost
		return "lazy-value", nil
	})

	do.ProvideNamedTransient(root, "transient-perf", func(i do.Injector) (string, error) {
		time.Sleep(1 * time.Millisecond) // Simulate creation cost
		return "transient-value", nil
	})

	// Measure performance of multiple invocations
	const numInvocations = 100

	// Eager service should be fastest (no initialization cost)
	start := time.Now()
	for i := 0; i < numInvocations; i++ {
		_, err := do.InvokeNamed[string](root, "eager-perf")
		is.NoError(err)
	}
	eagerDuration := time.Since(start)

	// Lazy service should be slower on first invocation, then fast
	start = time.Now()
	for i := 0; i < numInvocations; i++ {
		_, err := do.InvokeNamed[string](root, "lazy-perf")
		is.NoError(err)
	}
	lazyDuration := time.Since(start)

	// Transient service should be slowest (creation cost on every invocation)
	start = time.Now()
	for i := 0; i < numInvocations; i++ {
		_, err := do.InvokeNamed[string](root, "transient-perf")
		is.NoError(err)
	}
	transientDuration := time.Since(start)

	// Verify performance characteristics
	is.Less(eagerDuration, lazyDuration)
	is.Less(lazyDuration, transientDuration)
}

// TestMixedServiceTypesMemoryUsage tests memory usage patterns of different service types
func TestMixedServiceTypesMemoryUsage(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services that allocate memory
	do.ProvideNamedValue(root, "eager-memory", make([]byte, 1024))

	do.ProvideNamed(root, "lazy-memory", func(i do.Injector) ([]byte, error) {
		return make([]byte, 1024), nil
	})

	do.ProvideNamedTransient(root, "transient-memory", func(i do.Injector) ([]byte, error) {
		return make([]byte, 1024), nil
	})

	// Invoke services multiple times
	const numInvocations = 100

	// Eager service - memory allocated once
	for i := 0; i < numInvocations; i++ {
		_, err := do.InvokeNamed[[]byte](root, "eager-memory")
		is.NoError(err)
	}

	// Lazy service - memory allocated once, then reused
	for i := 0; i < numInvocations; i++ {
		_, err := do.InvokeNamed[[]byte](root, "lazy-memory")
		is.NoError(err)
	}

	// Transient service - memory allocated on every invocation
	for i := 0; i < numInvocations; i++ {
		_, err := do.InvokeNamed[[]byte](root, "transient-memory")
		is.NoError(err)
	}

	// All services should still work correctly
	_, err := do.InvokeNamed[[]byte](root, "eager-memory")
	is.NoError(err)
	_, err = do.InvokeNamed[[]byte](root, "lazy-memory")
	is.NoError(err)
	_, err = do.InvokeNamed[[]byte](root, "transient-memory")
	is.NoError(err)
}

// TestMixedServiceTypesErrorHandling tests error handling with different service types
func TestMixedServiceTypesErrorHandling(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services of different types that can fail
	do.ProvideNamed(root, "lazy-error", func(i do.Injector) (string, error) {
		return "", fmt.Errorf("lazy service error")
	})

	do.ProvideNamedTransient(root, "transient-error", func(i do.Injector) (string, error) {
		return "", fmt.Errorf("transient service error")
	})

	// Service that depends on failing services
	do.ProvideNamed(root, "dependent-on-errors", func(i do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](i, "lazy-error")
		if err != nil {
			return "", err
		}
		_, err = do.InvokeNamed[string](i, "transient-error")
		if err != nil {
			return "", err
		}
		return "should-not-reach-here", nil
	})

	// Attempting to invoke should fail
	_, err := do.InvokeNamed[string](root, "dependent-on-errors")
	is.Error(err)
	is.Contains(err.Error(), "lazy service error")
}

// Helper types for testing

type mixedLifecycleService struct {
	name          string
	shutdownOrder *[]string
	mu            *sync.Mutex
}

func (m *mixedLifecycleService) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	*m.shutdownOrder = append(*m.shutdownOrder, m.name)
	m.mu.Unlock()
	return nil
}
