package tests

import (
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
)

func TestCrossScopeCircularDependencies(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	// Create multi-level nested scope structure
	root := do.New()
	level1 := root.Scope("level1")
	level2a := level1.Scope("level2a")
	level2b := level1.Scope("level2b")

	// Create circular dependency: ServiceA in level1 depends on ServiceB in level2a
	// ServiceB depends on ServiceC in level2b, ServiceC depends on ServiceA in level1
	do.ProvideNamed(level1, "service-a", func(i do.Injector) (string, error) {
		// This should cause circular dependency
		_, err := do.InvokeNamed[string](i, "service-b")
		if err != nil {
			return "", err
		}
		return "service-a-value", nil
	})

	do.ProvideNamed(level2a, "service-b", func(i do.Injector) (string, error) {
		// This should cause circular dependency
		_, err := do.InvokeNamed[string](i, "service-c")
		if err != nil {
			return "", err
		}
		return "service-b-value", nil
	})

	do.ProvideNamed(level2b, "service-c", func(i do.Injector) (string, error) {
		// This should cause circular dependency
		_, err := do.InvokeNamed[string](i, "service-a")
		if err != nil {
			return "", err
		}
		return "service-c-value", nil
	})

	// Attempting to invoke should detect the dependency issue
	// Cross-scope circular dependencies currently result in "service not found"
	// because the service resolution fails before reaching circular dependency detection
	_, err := do.InvokeNamed[string](level1, "service-a")
	is.Error(err)
	is.Contains(err.Error(), "could not find service", "Expected service not found error, got: %s", err.Error())
}

func TestCircularDependencyDetection(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	// Test various circular dependency scenarios
	testCases := []struct {
		name        string
		setupFunc   func(do.Injector) error
		expectError bool
	}{
		{
			name: "Direct circular dependency",
			setupFunc: func(i do.Injector) error {
				do.ProvideNamed(i, "service-a", func(inj do.Injector) (string, error) {
					_, err := do.InvokeNamed[string](inj, "service-b")
					return "a", err
				})
				do.ProvideNamed(i, "service-b", func(inj do.Injector) (string, error) {
					_, err := do.InvokeNamed[string](inj, "service-a")
					return "b", err
				})
				return nil
			},
			expectError: true,
		},
		{
			name: "Indirect circular dependency",
			setupFunc: func(i do.Injector) error {
				do.ProvideNamed(i, "service-a", func(inj do.Injector) (string, error) {
					_, err := do.InvokeNamed[string](inj, "service-b")
					return "a", err
				})
				do.ProvideNamed(i, "service-b", func(inj do.Injector) (string, error) {
					_, err := do.InvokeNamed[string](inj, "service-c")
					return "b", err
				})
				do.ProvideNamed(i, "service-c", func(inj do.Injector) (string, error) {
					_, err := do.InvokeNamed[string](inj, "service-a")
					return "c", err
				})
				return nil
			},
			expectError: true,
		},
		{
			name: "No circular dependency",
			setupFunc: func(i do.Injector) error {
				do.ProvideNamedValue(i, "service-a", "a")
				do.ProvideNamed(i, "service-b", func(inj do.Injector) (string, error) {
					_, err := do.InvokeNamed[string](inj, "service-a")
					return "b", err
				})
				return nil
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			root := do.New()
			err := tc.setupFunc(root)
			is.NoError(err)

			if tc.expectError {
				_, err = do.InvokeNamed[string](root, "service-a")
				is.Error(err)
				is.Contains(err.Error(), "circular dependency")
			} else {
				_, err = do.InvokeNamed[string](root, "service-b")
				is.NoError(err)
			}
		})
	}
}

// TestCircularDependency_Simple tests simple circular dependency detection
func TestCircularDependency_Simple(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case 1: A -> A (self-reference)
	i := do.New()
	do.ProvideNamed(i, "serviceA", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceA")
		return "A", err
	})

	_, err := do.InvokeNamed[string](i, "serviceA")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")

	// Test case 2: A -> B -> A
	i = do.New()
	do.ProvideNamed(i, "serviceA", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceB")
		return "A", err
	})
	do.ProvideNamed(i, "serviceB", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceA")
		return "B", err
	})

	_, err = do.InvokeNamed[string](i, "serviceA")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_Complex tests complex circular dependency scenarios
func TestCircularDependency_Complex(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	// Test case 1: A -> B -> C -> A (3-service cycle)
	i := do.New()
	do.ProvideNamed(i, "serviceA", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceB")
		return "A", err
	})
	do.ProvideNamed(i, "serviceB", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceC")
		return "B", err
	})
	do.ProvideNamed(i, "serviceC", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceA")
		return "C", err
	})

	_, err := do.InvokeNamed[string](i, "serviceA")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")

	// Test case 2: A -> B -> C -> D -> A (4-service cycle)
	i = do.New()
	do.ProvideNamed(i, "serviceA", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceB")
		return "A", err
	})
	do.ProvideNamed(i, "serviceB", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceC")
		return "B", err
	})
	do.ProvideNamed(i, "serviceC", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceD")
		return "C", err
	})
	do.ProvideNamed(i, "serviceD", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceA")
		return "D", err
	})

	_, err = do.InvokeNamed[string](i, "serviceA")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_GenericType tests circular dependencies with generic type resolution
func TestCircularDependency_GenericType(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	type ServiceA struct{ Value string }
	type ServiceB struct{ Value string }

	// Test case: ServiceA depends on ServiceB which depends on ServiceA
	i := do.New()
	do.Provide(i, func(inj do.Injector) (*ServiceA, error) {
		_, err := do.Invoke[*ServiceB](inj)
		return &ServiceA{Value: "A"}, err
	})
	do.Provide(i, func(inj do.Injector) (*ServiceB, error) {
		_, err := do.Invoke[*ServiceA](inj)
		return &ServiceB{Value: "B"}, err
	})

	_, err := do.Invoke[*ServiceA](i)
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_MixedInvocation tests circular dependencies with mixed invocation methods
func TestCircularDependency_MixedInvocation(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case: Mix of named and generic type invocations
	i := do.New()
	do.ProvideNamed(i, "serviceA", func(inj do.Injector) (string, error) {
		_, err := do.Invoke[string](inj)
		return "A", err
	})
	do.Provide(i, func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceA")
		return "B", err
	})

	_, err := do.InvokeNamed[string](i, "serviceA")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_TransientServices tests circular dependencies with transient services
func TestCircularDependency_TransientServices(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case: Transient services can also create circular dependencies
	i := do.New()
	do.ProvideTransient(i, func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceB")
		return "A", err
	})
	do.ProvideNamed(i, "serviceB", func(inj do.Injector) (string, error) {
		_, err := do.Invoke[string](inj)
		return "B", err
	})

	_, err := do.Invoke[string](i)
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_EagerServices tests circular dependencies with eager services
func TestCircularDependency_EagerServices(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case: Eager services can create circular dependencies during initialization
	i := do.New()
	do.ProvideNamedValue(i, "serviceA", "A")
	do.ProvideNamed(i, "serviceB", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceA")
		return "B", err
	})
	do.ProvideNamed(i, "serviceC", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceB")
		return "C", err
	})

	// This should work since serviceA is eager and doesn't depend on others
	svc, err := do.InvokeNamed[string](i, "serviceC")
	is.NoError(err)
	is.Equal("C", svc)
}

// TestCircularDependency_DeepChain tests very deep dependency chains that eventually become circular
func TestCircularDependency_DeepChain(t *testing.T) {
	testWithTimeout(t, 200*time.Millisecond)
	is := assert.New(t)

	// Test case: Deep chain A -> B -> C -> D -> E -> A
	i := do.New()
	do.ProvideNamed(i, "A", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "B")
		return "A", err
	})
	do.ProvideNamed(i, "B", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "C")
		return "B", err
	})
	do.ProvideNamed(i, "C", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "D")
		return "C", err
	})
	do.ProvideNamed(i, "D", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "E")
		return "D", err
	})
	do.ProvideNamed(i, "E", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "A")
		return "E", err
	})

	_, err := do.InvokeNamed[string](i, "A")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_NoCircular tests that valid dependency chains work correctly
func TestCircularDependency_NoCircular(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case: Valid chain A -> B -> C (no circular dependency)
	i := do.New()
	do.ProvideNamed(i, "A", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "B")
		return "A", err
	})
	do.ProvideNamed(i, "B", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "C")
		return "B", err
	})
	do.ProvideNamed(i, "C", func(inj do.Injector) (string, error) {
		return "C", nil
	})

	svc, err := do.InvokeNamed[string](i, "A")
	is.NoError(err)
	is.Equal("A", svc)
}

// TestCircularDependency_ServiceAlias tests circular dependencies with service aliases
func TestCircularDependency_ServiceAlias(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case: Service alias can create circular dependencies
	i := do.New()
	do.ProvideNamed(i, "serviceA", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "aliasB")
		return "A", err
	})
	do.ProvideNamed(i, "serviceB", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceA")
		return "B", err
	})
	do.MustAsNamed[string, string](i, "serviceB", "aliasB")

	_, err := do.InvokeNamed[string](i, "serviceA")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_SelfInjection tests services that inject themselves
func TestCircularDependency_SelfInjection(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case 1: Named service injecting itself
	i := do.New()
	do.ProvideNamed(i, "self-injecting-service", func(inj do.Injector) (string, error) {
		// This service tries to inject itself
		_, err := do.InvokeNamed[string](inj, "self-injecting-service")
		return "self-injected", err
	})

	_, err := do.InvokeNamed[string](i, "self-injecting-service")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")

	// Test case 2: Generic service injecting itself
	i = do.New()
	do.Provide(i, func(inj do.Injector) (string, error) {
		// This service tries to inject itself using generic type
		_, err := do.Invoke[string](inj)
		return "self-injected-generic", err
	})

	_, err = do.Invoke[string](i)
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")

	// Test case 3: Struct service injecting itself
	type SelfInjectingStruct struct {
		Value string
		Self  *SelfInjectingStruct
	}

	i = do.New()
	do.Provide(i, func(inj do.Injector) (*SelfInjectingStruct, error) {
		// This service tries to inject itself
		self, err := do.Invoke[*SelfInjectingStruct](inj)
		if err != nil {
			return nil, err
		}
		return &SelfInjectingStruct{
			Value: "self-injected-struct",
			Self:  self,
		}, nil
	})

	_, err = do.Invoke[*SelfInjectingStruct](i)
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")

	// Test case 4: Transient service injecting itself
	i = do.New()
	do.ProvideTransient(i, func(inj do.Injector) (string, error) {
		// Transient services can also create self-injection circular dependencies
		_, err := do.Invoke[string](inj)
		return "self-injected-transient", err
	})

	_, err = do.Invoke[string](i)
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")

	// Test case 5: Service with multiple self-injections
	i = do.New()
	do.ProvideNamed(i, "multi-self-inject", func(inj do.Injector) (string, error) {
		// Multiple attempts to inject itself
		_, err1 := do.InvokeNamed[string](inj, "multi-self-inject")
		if err1 != nil {
			return "", err1
		}
		_, err2 := do.InvokeNamed[string](inj, "multi-self-inject")
		if err2 != nil {
			return "", err2
		}
		return "multi-self-injected", nil
	})

	_, err = do.InvokeNamed[string](i, "multi-self-inject")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_InterfaceResolution tests circular dependencies with interface resolution
func TestCircularDependency_InterfaceResolution(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case: Interface resolution can create circular dependencies
	i := do.New()
	do.Provide(i, func(inj do.Injector) (string, error) {
		_, err := do.Invoke[string](inj)
		return "A", err
	})
	do.Provide(i, func(inj do.Injector) (int, error) {
		_, err := do.Invoke[string](inj)
		return 42, err
	})

	_, err := do.Invoke[string](i)
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_EdgeCases tests edge cases in circular dependency detection
func TestCircularDependency_EdgeCases(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	// Test case 1: Empty service name (should not cause issues)
	i := do.New()
	do.ProvideNamed(i, "", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "")
		return "empty", err
	})

	_, err := do.InvokeNamed[string](i, "")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")

	// Test case 2: Very long service names
	longName := "very_long_service_name_that_is_quite_long_and_should_be_handled_properly"
	i = do.New()
	do.ProvideNamed(i, longName, func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, longName)
		return "long", err
	})

	_, err = do.InvokeNamed[string](i, longName)
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")
}

// TestCircularDependency_Recovery tests that circular dependency errors don't leave the container in a bad state
func TestCircularDependency_Recovery(t *testing.T) {
	testWithTimeout(t, 100*time.Millisecond)
	is := assert.New(t)

	i := do.New()
	do.ProvideNamed(i, "serviceA", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceB")
		return "A", err
	})
	do.ProvideNamed(i, "serviceB", func(inj do.Injector) (string, error) {
		_, err := do.InvokeNamed[string](inj, "serviceA")
		return "B", err
	})

	// First invocation should fail with circular dependency
	_, err := do.InvokeNamed[string](i, "serviceA")
	is.Error(err)
	is.Contains(err.Error(), "circular dependency")

	// Container should still be usable for other services
	do.ProvideNamed(i, "serviceC", func(inj do.Injector) (string, error) {
		return "C", nil
	})

	svc, err := do.InvokeNamed[string](i, "serviceC")
	is.NoError(err)
	is.Equal("C", svc)
}
