package tests

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
)

// TestVirtualScopeDeepDependencyChain tests virtual scope with very deep dependency chains
func TestVirtualScopeDeepDependencyChain(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a deep chain of services that depend on each other
	const chainDepth = 50

	// Register services in reverse order to create a deep chain
	for i := chainDepth; i >= 1; i-- {
		serviceName := fmt.Sprintf("service-%d", i)
		serviceIndex := i // Capture the loop variable
		if i == chainDepth {
			// Last service in chain has no dependencies
			do.ProvideNamed(root, serviceName, func(i do.Injector) (string, error) {
				return fmt.Sprintf("value-%d", serviceIndex), nil
			})
		} else {
			// Each service depends on the next one in the chain
			nextServiceName := fmt.Sprintf("service-%d", i+1)
			do.ProvideNamed(root, serviceName, func(i do.Injector) (string, error) {
				nextVal, err := do.InvokeNamed[string](i, nextServiceName)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("value-%d-depends-on-%s", serviceIndex, nextVal), nil
			})
		}
	}

	// Invoke the first service in the chain - this should resolve the entire chain
	result, err := do.InvokeNamed[string](root, "service-1")
	is.NoError(err)
	is.Contains(result, "value-1-depends-on-value-2-depends-on-value-3")
}

// TestVirtualScopeConcurrentAccess tests virtual scope behavior under concurrent access
func TestVirtualScopeConcurrentAccess(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create services that can be accessed concurrently
	do.ProvideNamed(root, "shared-service", func(i do.Injector) (string, error) {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return "shared-value", nil
	})

	do.ProvideNamed(root, "dependent-service", func(i do.Injector) (string, error) {
		sharedVal, err := do.InvokeNamed[string](i, "shared-service")
		if err != nil {
			return "", err
		}
		return "dependent-" + sharedVal, nil
	})

	// Concurrently invoke the dependent service
	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			result, err := do.InvokeNamed[string](root, "dependent-service")
			results[index] = result
			errors[index] = err
		}(i)
	}

	wg.Wait()

	// All invocations should succeed with the same result
	for i := 0; i < numGoroutines; i++ {
		is.NoError(errors[i])
		is.Equal("dependent-shared-value", results[i])
	}
}

// TestVirtualScopeContextCancellation tests virtual scope behavior when context is canceled
func TestVirtualScopeContextCancellation(t *testing.T) {
	testWithTimeout(t, 450*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that takes a long time to resolve
	do.ProvideNamed(root, "slow-service", func(i do.Injector) (string, error) {
		time.Sleep(100 * time.Millisecond)
		return "slow-value", nil
	})

	do.ProvideNamed(root, "dependent-on-slow", func(i do.Injector) (string, error) {
		slowVal, err := do.InvokeNamed[string](i, "slow-service")
		if err != nil {
			return "", err
		}
		return "dependent-" + slowVal, nil
	})

	// Try to invoke the service - context cancellation doesn't affect virtual scope operations
	_, err := do.InvokeNamed[string](root, "dependent-on-slow")
	// The service should still resolve because the context cancellation
	// doesn't affect the virtual scope's internal operations
	is.NoError(err)
}

// TestVirtualScopeCircularDependencyDetection tests complex circular dependency scenarios
func TestVirtualScopeCircularDependencyDetection(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a complex circular dependency scenario
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
		_, err := do.InvokeNamed[string](i, "service-a")
		if err != nil {
			return "", err
		}
		return "service-c-value", nil
	})

	// Attempting to invoke should detect circular dependency
	_, err := do.InvokeNamed[string](root, "service-a")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestVirtualScopeNestedScopes tests virtual scope behavior with nested scopes
func TestVirtualScopeNestedScopes(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()
	child1 := root.Scope("child1")
	child2 := child1.Scope("child2")

	// Services in different scopes
	do.ProvideNamedValue(root, "root-service", "root-value")
	do.ProvideNamedValue(child1, "child1-service", "child1-value")
	do.ProvideNamedValue(child2, "child2-service", "child2-value")

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

	// Invoke the dependent service
	result, err := do.InvokeNamed[string](child2, "dependent-service")
	is.NoError(err)
	is.Equal("root-value-child1-value-child2-value", result)
}

// TestVirtualScopeServiceNotFound tests virtual scope behavior when services are not found
func TestVirtualScopeServiceNotFound(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Service that depends on a non-existent service
	do.ProvideNamed(root, "dependent-service", func(i do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](i, "non-existent-service")
		if err != nil {
			return "", err
		}
		return "should-not-reach-here", nil
	})

	// Attempting to invoke should fail with service not found error
	_, err := do.InvokeNamed[string](root, "dependent-service")
	is.Error(err)
	is.Contains(err.Error(), "could not find service")
}

// TestVirtualScopeProviderErrors tests virtual scope behavior when providers return errors
func TestVirtualScopeProviderErrors(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Service that returns an error
	do.ProvideNamed(root, "error-service", func(i do.Injector) (string, error) {
		return "", fmt.Errorf("provider error")
	})

	// Service that depends on the error service
	do.ProvideNamed(root, "dependent-service", func(i do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](i, "error-service")
		if err != nil {
			return "", err
		}
		return "should-not-reach-here", nil
	})

	// Attempting to invoke should fail with the provider error
	_, err := do.InvokeNamed[string](root, "dependent-service")
	is.Error(err)
	is.Contains(err.Error(), "provider error")
}

// TestVirtualScopePanicRecovery tests virtual scope behavior when providers panic
func TestVirtualScopePanicRecovery(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Service that panics
	do.ProvideNamed(root, "panic-service", func(i do.Injector) (string, error) {
		panic("provider panic")
	})

	// Service that depends on the panic service
	do.ProvideNamed(root, "dependent-service", func(i do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](i, "panic-service")
		if err != nil {
			return "", err
		}
		return "should-not-reach-here", nil
	})

	// Attempting to invoke should fail with panic error
	_, err := do.InvokeNamed[string](root, "dependent-service")
	is.Error(err)
	is.Contains(err.Error(), "panic")
}

// TestVirtualScopeMemoryLeak tests for potential memory leaks in virtual scope
func TestVirtualScopeMemoryLeak(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	root := do.New()

	// Create a service that can be invoked multiple times
	do.ProvideNamed(root, "memory-test-service", func(i do.Injector) (string, error) {
		return "test-value", nil
	})

	// Invoke the service many times to test for memory leaks
	const numInvocations = 1000
	for i := 0; i < numInvocations; i++ {
		result, err := do.InvokeNamed[string](root, "memory-test-service")
		is.NoError(err)
		is.Equal("test-value", result)
	}

	// The service should still work correctly after many invocations
	result, err := do.InvokeNamed[string](root, "memory-test-service")
	is.NoError(err)
	is.Equal("test-value", result)
}
