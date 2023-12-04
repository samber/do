---
title: Service invocation
description: Invoke eager or lazy services from the DI container
sidebar_position: 1
---

# Service invocation

Service invocation in a Dependency Injection (DI) framework refers to the process of requesting the singleton of a service from the DI container. This is typically done when a piece of code needs to use that service. In the case of lazy loading, a recursive service invocation may occur. Transient services behave like factories.

In the context of the Go code you're working with, there are several helper functions provided for service invocation:

- `do.Invoke[T any](do.Injector) (T, error)`: This function invokes a service of type T from the DI container. If the service can be successfully created and returned, it does so. Otherwise, it returns an error.

- `do.InvokeNamed[T any](do.Injector, string) (T, error)`: This function is similar to `do.Invoke`, but it allows you to invoke a service by its name. This is useful when you have multiple instances of the same type and you want to distinguish between them.

- `do.MustInvoke[T any](do.Injector) T`: This function is a variant of `do.Invoke` that panics if the service cannot be created. This is useful when you're sure that the service should always be available, and if it's not, it's an error that should stop the program.

- `do.MustInvokeNamed[T any](do.Injector, string) T`: This function is a variant of `do.InvokeNamed` that also panics if the service cannot be created.

🚀 Lazy services are loaded in invocation order.

🐎 Lazy service invocation is protected against concurrent loading.

🧙‍♂️ When multiple [scopes](../scopes/scope.md) are assembled into a big application, the service lookup is recursive from the current nested scope to the root scope.

:::warning

Circular dependencies are not allowed. Services must be invoked in a Directed Acyclic Graph way.

:::


## Example

```go
type MyService struct {
    IP string
}

i := do.New()

do.ProvideNamedValue(i, "config.ip", "127.0.0.1")
do.Provide(i, func(i do.Injector) (*MyService, error) {
    return &MyService{
      IP: do.MustInvokeNamed(i, "config.ip"),
    }, nil
})

myService, err := do.Invoke[*MyService](i)
```

## Other services as dependencies

A service might rely on other services. In that case, you should invoke dependencies in the service provider instead of storing the injector for later.

```go
// ❌ bad
type MyService struct {
  injector do.Injector
}
func NewMyService(i do.Injector) (*MyService, error) {
  return &MyService{
    injector: i,
  }, nil
}

// ✅ good
type MyService struct {
  dependency *MyDependency
}
func NewMyService(i do.Injector) (*MyService, error) {
  return &MyService{
    dep: do.MustInvoke[*MyDependency](i),   // <- recursive invocation on service construction
  }, nil
}
```

## Error handling

Any panic during lazy loading is converted into a Go `error`.

An error is returned on missing service.
