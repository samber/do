package do

import (
	"context"
	"fmt"
	"time"
)

type lifecycleTestService struct {
	Name string
}

func (s *lifecycleTestService) HealthCheck() error {
	return nil
}

func (s *lifecycleTestService) Shutdown() error {
	return nil
}

func lifecycleTestServiceProvider(i Injector) (*lifecycleTestService, error) {
	return &lifecycleTestService{Name: "lifecycle-test-service"}, nil
}

func ExampleHealthCheck() {
	injector := New()

	Provide(injector, lifecycleTestServiceProvider)
	_, _ = Invoke[*lifecycleTestService](injector)

	err := HealthCheck[*lifecycleTestService](injector)
	fmt.Println(err)
	// Output: <nil>
}

func ExampleHealthCheckWithContext() {
	injector := New()

	Provide(injector, lifecycleTestServiceProvider)
	_, _ = Invoke[*lifecycleTestService](injector)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := HealthCheckWithContext[*lifecycleTestService](ctx, injector)
	fmt.Println(err)
	// Output: <nil>
}

func ExampleHealthCheckNamed() {
	injector := New()

	ProvideNamed(injector, "main-database", lifecycleTestServiceProvider)
	_, _ = InvokeNamed[*lifecycleTestService](injector, "main-database")

	err := HealthCheckNamed(injector, "main-database")
	fmt.Println(err)
	// Output: <nil>
}

func ExampleHealthCheckNamedWithContext() {
	injector := New()

	ProvideNamed(injector, "main-database", lifecycleTestServiceProvider)
	_, _ = InvokeNamed[*lifecycleTestService](injector, "main-database")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := HealthCheckNamedWithContext(ctx, injector, "main-database")
	fmt.Println(err)
	// Output: <nil>
}

func ExampleShutdown() {
	injector := New()

	Provide(injector, lifecycleTestServiceProvider)
	_, _ = Invoke[*lifecycleTestService](injector)

	err := Shutdown[*lifecycleTestService](injector)
	fmt.Println(err)
	// Output: <nil>
}

func ExampleShutdownWithContext() {
	injector := New()

	Provide(injector, lifecycleTestServiceProvider)
	_, _ = Invoke[*lifecycleTestService](injector)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ShutdownWithContext[*lifecycleTestService](ctx, injector)
	fmt.Println(err)
	// Output: <nil>
}

func ExampleShutdownNamed() {
	injector := New()

	ProvideNamed(injector, "main-database", lifecycleTestServiceProvider)
	_, _ = InvokeNamed[*lifecycleTestService](injector, "main-database")

	err := ShutdownNamed(injector, "main-database")
	fmt.Println(err)
	// Output: <nil>
}

func ExampleShutdownNamedWithContext() {
	injector := New()

	ProvideNamed(injector, "main-database", lifecycleTestServiceProvider)
	_, _ = InvokeNamed[*lifecycleTestService](injector, "main-database")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ShutdownNamedWithContext(ctx, injector, "main-database")
	fmt.Println(err)
	// Output: <nil>
}

func ExampleMustShutdown() {
	injector := New()

	Provide(injector, lifecycleTestServiceProvider)
	_, _ = Invoke[*lifecycleTestService](injector)

	// This will panic if shutdown fails
	MustShutdown[*lifecycleTestService](injector)
	fmt.Println("shutdown completed")
	// Output: shutdown completed
}

func ExampleMustShutdownWithContext() {
	injector := New()

	Provide(injector, lifecycleTestServiceProvider)
	_, _ = Invoke[*lifecycleTestService](injector)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will panic if shutdown fails
	MustShutdownWithContext[*lifecycleTestService](ctx, injector)
	fmt.Println("shutdown completed")
	// Output: shutdown completed
}

func ExampleMustShutdownNamed() {
	injector := New()

	ProvideNamed(injector, "main-database", lifecycleTestServiceProvider)
	_, _ = InvokeNamed[*lifecycleTestService](injector, "main-database")

	// This will panic if shutdown fails
	MustShutdownNamed(injector, "main-database")
	fmt.Println("shutdown completed")
	// Output: shutdown completed
}

func ExampleMustShutdownNamedWithContext() {
	injector := New()

	ProvideNamed(injector, "main-database", lifecycleTestServiceProvider)
	_, _ = InvokeNamed[*lifecycleTestService](injector, "main-database")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will panic if shutdown fails
	MustShutdownNamedWithContext(ctx, injector, "main-database")
	fmt.Println("shutdown completed")
	// Output: shutdown completed
}
