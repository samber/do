package do

import (
	"fmt"
)

func ExampleNew() {
	injector := New()

	ProvideNamedValue(injector, "PG_URI", "postgres://user:pass@host:5432/db")
	uri, err := InvokeNamed[string](injector, "PG_URI")

	fmt.Println(uri)
	fmt.Println(err)
	// Output:
	// postgres://user:pass@host:5432/db
	// <nil>
}

func ExampleDefaultRootScope() {
	ProvideNamedValue(nil, "PG_URI", "postgres://user:pass@host:5432/db")
	uri, err := InvokeNamed[string](nil, "PG_URI")

	fmt.Println(uri)
	fmt.Println(err)
	// Output:
	// postgres://user:pass@host:5432/db
	// <nil>
}

func ExampleNewWithOpts() {
	injector := NewWithOpts(&InjectorOpts{
		HookAfterShutdown: []func(scope *Scope, serviceName string, err error){
			func(scope *Scope, serviceName string, err error) {
				fmt.Printf("service shutdown: %s\n", serviceName)
			},
		},
	})

	ProvideNamed(injector, "PG_URI", func(i Injector) (string, error) {
		return "postgres://user:pass@host:5432/db", nil
	})
	MustInvokeNamed[string](injector, "PG_URI")
	err := injector.Shutdown()
	fmt.Println(err)

	// Output:
	// service shutdown: PG_URI
	// <nil>
}

func ExampleRootScope_AddBeforeRegistrationHook() {
	injector := New()

	injector.AddBeforeRegistrationHook(func(scope *Scope, serviceName string) {
		fmt.Printf("registering service: %s\n", serviceName)
	})

	ProvideNamedValue(injector, "test", "value")
	// Output: registering service: test
}

func ExampleRootScope_AddAfterRegistrationHook() {
	injector := New()

	injector.AddAfterRegistrationHook(func(scope *Scope, serviceName string) {
		fmt.Printf("registered service: %s\n", serviceName)
	})

	ProvideNamedValue(injector, "test", "value")
	// Output: registered service: test
}

func ExampleRootScope_AddBeforeInvocationHook() {
	injector := New()

	injector.AddBeforeInvocationHook(func(scope *Scope, serviceName string) {
		fmt.Printf("invoking service: %s\n", serviceName)
	})

	ProvideNamedValue(injector, "test", "value")
	_, _ = InvokeNamed[string](injector, "test")
	// Output: invoking service: test
}

func ExampleRootScope_AddAfterInvocationHook() {
	injector := New()

	injector.AddAfterInvocationHook(func(scope *Scope, serviceName string, err error) {
		fmt.Printf("invoked service: %s, error: %v\n", serviceName, err)
	})

	ProvideNamedValue(injector, "test", "value")
	_, _ = InvokeNamed[string](injector, "test")
	// Output: invoked service: test, error: <nil>
}

func ExampleRootScope_AddBeforeShutdownHook() {
	injector := New()

	injector.AddBeforeShutdownHook(func(scope *Scope, serviceName string) {
		fmt.Printf("shutting down service: %s\n", serviceName)
	})

	ProvideNamedValue(injector, "test", "value")
	_ = injector.Shutdown()
	// Output: shutting down service: test
}

func ExampleRootScope_AddAfterShutdownHook() {
	injector := New()

	injector.AddAfterShutdownHook(func(scope *Scope, serviceName string, err error) {
		fmt.Printf("shut down service: %s, error: %v\n", serviceName, err)
	})

	ProvideNamedValue(injector, "test", "value")
	_ = injector.Shutdown()
	// Output: shut down service: test, error: <nil>
}

func ExampleRootScope_Clone() {
	injector := New()
	ProvideNamedValue(injector, "test", "value")

	clone := injector.Clone()
	services := clone.ListProvidedServices()

	fmt.Println(len(services))
	fmt.Println(services[0].Service)
	// Output:
	// 1
	// test
}

func ExampleRootScope_CloneWithOpts() {
	injector := New()
	ProvideNamedValue(injector, "test", "value")

	clone := injector.CloneWithOpts(&InjectorOpts{
		HookAfterShutdown: []func(scope *Scope, serviceName string, err error){
			func(scope *Scope, serviceName string, err error) {
				fmt.Printf("shutdown: %s\n", serviceName)
			},
		},
	})

	_ = clone.Shutdown()
	// Output: shutdown: test
}

func ExampleRootScope_ShutdownOnSignals() {
	// This example is commented out because it would block waiting for signals
	// injector := New()
	// signal, errors := injector.ShutdownOnSignals(syscall.SIGINT, syscall.SIGTERM)
	// fmt.Println(signal)
	// fmt.Println(errors.Len())
}

func ExampleRootScope_ShutdownOnSignalsWithContext() {
	// This example is commented out because it would block waiting for signals
	// injector := New()
	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	// defer cancel()
	// signal, errors := injector.ShutdownOnSignalsWithContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	// fmt.Println(signal)
	// fmt.Println(errors.Len())
}
