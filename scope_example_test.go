package do

import (
	"context"
	"fmt"
	"time"
)

func ExampleScope_ID() {
	injector := New()
	scope := injector.Scope("api")

	fmt.Println(scope.ID() != "")
	// Output: true
}

func ExampleScope_Name() {
	injector := New()
	scope := injector.Scope("api")

	fmt.Println(scope.Name())
	// Output: api
}

func ExampleScope_Scope() {
	injector := New()
	apiScope := injector.Scope("api")
	userScope := apiScope.Scope("user")

	fmt.Println(userScope.Name())
	// Output: user
}

func ExampleScope_RootScope() {
	injector := New()
	apiScope := injector.Scope("api")
	userScope := apiScope.Scope("user")
	root := userScope.RootScope()

	fmt.Println(root.Name())
	// Output: [root]
}

func ExampleScope_Ancestors() {
	injector := New()
	apiScope := injector.Scope("api")
	userScope := apiScope.Scope("user")
	ancestors := userScope.Ancestors()

	fmt.Println(len(ancestors))
	fmt.Println(ancestors[0].Name())
	// Output:
	// 2
	// api
}

func ExampleScope_Children() {
	injector := New()
	apiScope := injector.Scope("api")
	_ = apiScope.Scope("user")
	_ = apiScope.Scope("admin")
	children := apiScope.Children()

	fmt.Println(len(children))
	// Output: 2
}

func ExampleScope_ChildByID() {
	injector := New()
	apiScope := injector.Scope("api")
	userScope := apiScope.Scope("user")

	child, found := apiScope.ChildByID(userScope.ID())
	fmt.Println(found)
	fmt.Println(child.Name())
	// Output:
	// true
	// user
}

func ExampleScope_ChildByName() {
	injector := New()
	apiScope := injector.Scope("api")
	_ = apiScope.Scope("user")

	child, found := apiScope.ChildByName("user")
	fmt.Println(found)
	fmt.Println(child.Name())
	// Output:
	// true
	// user
}

func ExampleScope_ListProvidedServices() {
	type Configuration struct {
		Port int
	}

	injector := New()
	scope := injector.Scope("api")

	ProvideNamedValue(scope, "config", Configuration{Port: 8080})
	services := scope.ListProvidedServices()

	fmt.Println(len(services))
	fmt.Println(services[0].Service)
	// Output:
	// 1
	// config
}

func ExampleScope_ListInvokedServices() {
	type Configuration struct {
		Port int
	}

	injector := New()
	scope := injector.Scope("api")

	ProvideNamedValue(scope, "config", Configuration{Port: 8080})
	_, _ = InvokeNamed[Configuration](scope, "config")
	services := scope.ListInvokedServices()

	fmt.Println(len(services))
	fmt.Println(services[0].Service)
	// Output:
	// 1
	// config
}

type scopeDbService struct {
	Name string
}

func (s *scopeDbService) HealthCheck() error {
	return nil
}

func (s *scopeDbService) Shutdown() error {
	return fmt.Errorf("shutdown error")
}

func scopeDbServiceProvider(i Injector) (*scopeDbService, error) {
	return &scopeDbService{Name: "main-db"}, nil
}

func ExampleScope_HealthCheck() {
	injector := New()
	scope := injector.Scope("api")

	ProvideNamed(scope, "db", scopeDbServiceProvider)
	_, _ = InvokeNamed[*scopeDbService](scope, "db")
	health := scope.HealthCheck()

	fmt.Println(len(health))
	// Output: 1
}

func ExampleScope_HealthCheckWithContext() {
	injector := New()
	scope := injector.Scope("api")

	ProvideNamed(scope, "db", scopeDbServiceProvider)
	_, _ = InvokeNamed[*scopeDbService](scope, "db")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := scope.HealthCheckWithContext(ctx)
	fmt.Println(len(health))
	// Output: 1
}

func ExampleScope_Shutdown() {
	injector := New()
	scope := injector.Scope("api")

	ProvideNamed(scope, "db", scopeDbServiceProvider)
	_, _ = InvokeNamed[*scopeDbService](scope, "db")

	report := scope.Shutdown()
	fmt.Println(report.Succeed)
	fmt.Println(len(report.Services))
	fmt.Println(report.Error())
	// Output:
	// false
	// 1
	// DI: shutdown errors:
	//   - api > db: shutdown error
}

func ExampleScope_ShutdownWithContext() {
	injector := New()

	// Register a simple service without shutdown capability
	ProvideNamedValue(injector, "config", "value")
	_, _ = InvokeNamed[string](injector, "config")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	report := injector.ShutdownWithContext(ctx)
	fmt.Println(report.Succeed)
	fmt.Println(len(report.Services))
	fmt.Println(report.Error())
	// Output:
	// true
	// 1
}

func ExampleScope_ShutdownWithContext_timeout() {
	injector := New()
	scope := injector.Scope("api")

	ProvideNamed(scope, "db", scopeDbServiceProvider)
	_, _ = InvokeNamed[*scopeDbService](scope, "db")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	time.Sleep(100 * time.Millisecond) // will trigger timeout

	report := injector.ShutdownWithContext(ctx)
	fmt.Println(report.Error())
	// Output:
	// DI: shutdown errors:
	//   - api > db: context deadline exceeded
}
