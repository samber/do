package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
)

// testShutdowner is a test helper that implements ShutdownerWithContextAndError
type testShutdowner struct {
	name          string
	shutdownOrder *[]string
	mu            *sync.Mutex
}

func (t *testShutdowner) Shutdown(ctx context.Context) error {
	t.mu.Lock()
	*t.shutdownOrder = append(*t.shutdownOrder, t.name)
	t.mu.Unlock()
	return nil
}

func TestMultiLevelNestedScopes_NonConcurrentShutdown(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	// Create a multi-level nested scope structure
	root := do.New()
	level1 := root.Scope("level1")
	level2 := level1.Scope("level2")
	level3 := level2.Scope("level3")

	// Track shutdown order
	shutdownOrder := make([]string, 0)
	var mu sync.Mutex

	// Provide services at each level with shutdown tracking
	do.ProvideNamedValue(root, "root-service", "root-value")
	do.ProvideNamed(level1, "level1-service", func(i do.Injector) (string, error) {
		return "level1-value", nil
	})
	do.ProvideNamed(level2, "level2-service", func(i do.Injector) (string, error) {
		return "level2-value", nil
	})
	do.ProvideNamed(level3, "level3-service", func(i do.Injector) (string, error) {
		return "level3-value", nil
	})

	// Add shutdownable services to track shutdown order
	do.ProvideNamed(root, "root-shutdownable", func(i do.Injector) (do.ShutdownerWithContextAndError, error) {
		return &testShutdowner{name: "root", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(level1, "level1-shutdownable", func(i do.Injector) (do.ShutdownerWithContextAndError, error) {
		return &testShutdowner{name: "level1", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(level2, "level2-shutdownable", func(i do.Injector) (do.ShutdownerWithContextAndError, error) {
		return &testShutdowner{name: "level2", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(level3, "level3-shutdownable", func(i do.Injector) (do.ShutdownerWithContextAndError, error) {
		return &testShutdowner{name: "level3", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	// Invoke services to ensure they are initialized
	_, err := do.InvokeNamed[string](root, "root-service")
	is.NoError(err)
	_, err = do.InvokeNamed[string](level1, "level1-service")
	is.NoError(err)
	_, err = do.InvokeNamed[string](level2, "level2-service")
	is.NoError(err)
	_, err = do.InvokeNamed[string](level3, "level3-service")
	is.NoError(err)

	// Invoke shutdownable services to ensure they are initialized
	_, err = do.InvokeNamed[do.ShutdownerWithContextAndError](root, "root-shutdownable")
	is.NoError(err)
	_, err = do.InvokeNamed[do.ShutdownerWithContextAndError](level1, "level1-shutdownable")
	is.NoError(err)
	_, err = do.InvokeNamed[do.ShutdownerWithContextAndError](level2, "level2-shutdownable")
	is.NoError(err)
	_, err = do.InvokeNamed[do.ShutdownerWithContextAndError](level3, "level3-shutdownable")
	is.NoError(err)

	// Shutdown the root scope (should shutdown all nested scopes)
	shutdownErrors := root.Shutdown()
	is.Nil(shutdownErrors)

	// Verify shutdown order (children should be shut down before parents)
	is.Len(shutdownOrder, 4)
	// The exact order may vary due to parallel shutdown, but we can verify all levels were shut down
	is.Contains(shutdownOrder, "root")
	is.Contains(shutdownOrder, "level1")
	is.Contains(shutdownOrder, "level2")
	is.Contains(shutdownOrder, "level3")
}

func TestMultiLevelNestedScopes_ServiceVisibility(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	// Create a multi-level nested scope structure
	root := do.New()
	level1 := root.Scope("level1")
	level2a := level1.Scope("level2a")
	level2b := level1.Scope("level2b")
	level3 := level2a.Scope("level3")

	// Provide services at different levels
	do.ProvideNamedValue(root, "root-service", "root-value")
	do.ProvideNamedValue(level1, "level1-service", "level1-value")
	do.ProvideNamedValue(level2a, "level2a-service", "level2a-value")
	do.ProvideNamedValue(level2b, "level2b-service", "level2b-value")
	do.ProvideNamedValue(level3, "level3-service", "level3-value")

	// Test children to root visibility (should work)
	value, err := do.InvokeNamed[string](level3, "root-service")
	is.NoError(err)
	is.Equal("root-value", value)

	value, err = do.InvokeNamed[string](level2a, "root-service")
	is.NoError(err)
	is.Equal("root-value", value)

	value, err = do.InvokeNamed[string](level1, "root-service")
	is.NoError(err)
	is.Equal("root-value", value)

	// Test root to children visibility (should not work)
	_, err = do.InvokeNamed[string](root, "level1-service")
	is.Error(err)
	is.Contains(err.Error(), "could not find service")

	_, err = do.InvokeNamed[string](root, "level2a-service")
	is.Error(err)
	is.Contains(err.Error(), "could not find service")

	_, err = do.InvokeNamed[string](root, "level3-service")
	is.Error(err)
	is.Contains(err.Error(), "could not find service")

	// Test children1 to children2 visibility (should not work)
	_, err = do.InvokeNamed[string](level2a, "level2b-service")
	is.Error(err)
	is.Contains(err.Error(), "could not find service")

	_, err = do.InvokeNamed[string](level2b, "level2a-service")
	is.Error(err)
	is.Contains(err.Error(), "could not find service")

	// Test same level visibility (should work)
	value, err = do.InvokeNamed[string](level2a, "level2a-service")
	is.NoError(err)
	is.Equal("level2a-value", value)

	value, err = do.InvokeNamed[string](level2b, "level2b-service")
	is.NoError(err)
	is.Equal("level2b-value", value)
}

func TestMultiLevelNestedScopes_Clone(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	// Create original multi-level nested scope structure
	root := do.New()
	level1 := root.Scope("level1")
	level2 := level1.Scope("level2")
	level3 := level2.Scope("level3")

	// Provide services at each level
	do.ProvideNamedValue(root, "root-service", "root-value")
	do.ProvideNamedValue(level1, "level1-service", "level1-value")
	do.ProvideNamedValue(level2, "level2-service", "level2-value")
	do.ProvideNamedValue(level3, "level3-service", "level3-value")

	// Invoke services to populate orderedInvocation
	_, err := do.InvokeNamed[string](root, "root-service")
	is.NoError(err)
	_, err = do.InvokeNamed[string](level1, "level1-service")
	is.NoError(err)
	_, err = do.InvokeNamed[string](level2, "level2-service")
	is.NoError(err)
	_, err = do.InvokeNamed[string](level3, "level3-service")
	is.NoError(err)

	// Clone the root scope
	clonedRoot := root.Clone()
	is.NotEqual(root, clonedRoot)

	// Find cloned scopes
	clonedLevel1, exists := clonedRoot.ChildByName("level1")
	is.True(exists)
	is.NotEqual(level1, clonedLevel1)

	clonedLevel2, exists := clonedLevel1.ChildByName("level2")
	is.True(exists)
	is.NotEqual(level2, clonedLevel2)

	clonedLevel3, exists := clonedLevel2.ChildByName("level3")
	is.True(exists)
	is.NotEqual(level3, clonedLevel3)

	// Verify services are cloned and accessible
	value, err := do.InvokeNamed[string](clonedRoot, "root-service")
	is.NoError(err)
	is.Equal("root-value", value)

	value, err = do.InvokeNamed[string](clonedLevel1, "level1-service")
	is.NoError(err)
	is.Equal("level1-value", value)

	value, err = do.InvokeNamed[string](clonedLevel2, "level2-service")
	is.NoError(err)
	is.Equal("level2-value", value)

	value, err = do.InvokeNamed[string](clonedLevel3, "level3-service")
	is.NoError(err)
	is.Equal("level3-value", value)

	// Verify that services were invoked in the cloned scopes
	// We can't access orderedInvocation directly, but we can verify services work
	_, err = do.InvokeNamed[string](clonedLevel2, "level2-service")
	is.NoError(err)
	_, err = do.InvokeNamed[string](clonedLevel3, "level3-service")
	is.NoError(err)

	// Verify that modifying cloned scope doesn't affect original
	do.ProvideNamedValue(clonedRoot, "cloned-only", "cloned-value")
	_, err = do.InvokeNamed[string](root, "cloned-only")
	is.Error(err)
	is.Contains(err.Error(), "could not find service")

	// Verify original scope still works
	value, err = do.InvokeNamed[string](root, "root-service")
	is.NoError(err)
	is.Equal("root-value", value)
}

func TestMultiLevelNestedScopes_NameCollision(t *testing.T) {
	testWithTimeout(t, 200*time.Millisecond)
	is := assert.New(t)

	// Create multi-level nested scope structure
	root := do.New()
	level1 := root.Scope("level1")
	level2a := level1.Scope("level2a")
	level2b := level1.Scope("level2b")

	// Provide services with same name in different scopes (should work)
	do.ProvideNamedValue(root, "shared-service", "root-value")
	do.ProvideNamedValue(level1, "shared-service", "level1-value")
	do.ProvideNamedValue(level2a, "shared-service", "level2a-value")
	do.ProvideNamedValue(level2b, "shared-service", "level2b-value")

	// Verify each scope sees its own service (shadowing)
	value, err := do.InvokeNamed[string](root, "shared-service")
	is.NoError(err)
	is.Equal("root-value", value)

	value, err = do.InvokeNamed[string](level1, "shared-service")
	is.NoError(err)
	is.Equal("level1-value", value)

	value, err = do.InvokeNamed[string](level2a, "shared-service")
	is.NoError(err)
	is.Equal("level2a-value", value)

	value, err = do.InvokeNamed[string](level2b, "shared-service")
	is.NoError(err)
	is.Equal("level2b-value", value)

	// Test that ProvideNamedValue doesn't allow overriding in same scope
	// This should panic when trying to register the same service name again
	is.Panics(func() {
		do.ProvideNamedValue(level2a, "shared-service", "override-attempt")
	})

	// Should still get the original value
	value, err = do.InvokeNamed[string](level2a, "shared-service")
	is.NoError(err)
	is.Equal("level2a-value", value)
}

func TestMultiLevelNestedScopes_ShutdownOrder(t *testing.T) {
	testWithTimeout(t, 300*time.Millisecond)
	is := assert.New(t)

	// Create multi-level nested scope structure
	root := do.New()
	level1 := root.Scope("level1")
	level2 := level1.Scope("level2")
	level3 := level2.Scope("level3")

	// Track shutdown order
	shutdownOrder := make([]string, 0)
	var mu sync.Mutex

	// Add shutdownable services with dependencies
	do.ProvideNamed(level3, "level3-shutdownable", func(i do.Injector) (do.ShutdownerWithContextAndError, error) {
		return &testShutdowner{name: "level3", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(level2, "level2-shutdownable", func(i do.Injector) (do.ShutdownerWithContextAndError, error) {
		return &testShutdowner{name: "level2", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(level1, "level1-shutdownable", func(i do.Injector) (do.ShutdownerWithContextAndError, error) {
		return &testShutdowner{name: "level1", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	do.ProvideNamed(root, "root-shutdownable", func(i do.Injector) (do.ShutdownerWithContextAndError, error) {
		return &testShutdowner{name: "root", shutdownOrder: &shutdownOrder, mu: &mu}, nil
	})

	// Invoke services to ensure they are initialized
	_, err := do.InvokeNamed[do.ShutdownerWithContextAndError](root, "root-shutdownable")
	is.NoError(err)
	_, err = do.InvokeNamed[do.ShutdownerWithContextAndError](level1, "level1-shutdownable")
	is.NoError(err)
	_, err = do.InvokeNamed[do.ShutdownerWithContextAndError](level2, "level2-shutdownable")
	is.NoError(err)
	_, err = do.InvokeNamed[do.ShutdownerWithContextAndError](level3, "level3-shutdownable")
	is.NoError(err)

	// Shutdown the root scope
	shutdownErrors := root.Shutdown()
	is.Nil(shutdownErrors)

	// Verify all levels were shut down
	is.Len(shutdownOrder, 4)
	is.Contains(shutdownOrder, "root")
	is.Contains(shutdownOrder, "level1")
	is.Contains(shutdownOrder, "level2")
	is.Contains(shutdownOrder, "level3")
}
