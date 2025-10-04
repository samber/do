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

- `do.InvokeAs[T any](do.Injector) (T, error)`: This function invokes a service by finding the first service that matches the provided interface type T. It's useful for interface-based dependency injection without explicit aliasing.

- `do.InvokeAsAll[T any](do.Injector) ([]T, error)`: This function invokes all services that match the provided interface type T, returning them as a slice. Services are returned in deterministic order based on their registration names. This is useful when you need to work with multiple implementations of the same interface.

- `do.MustInvoke[T any](do.Injector) T`: This function is a variant of `do.Invoke` that panics if the service cannot be created. This is useful when you're sure that the service should always be available, and if it's not, it's an error that should stop the program.

- `do.MustInvokeNamed[T any](do.Injector, string) T`: This function is a variant of `do.InvokeNamed` that also panics if the service cannot be created.

- `do.MustInvokeAs[T any](do.Injector) T`: This function is a variant of `do.InvokeAs` that panics if the service cannot be found or created.

- `do.MustInvokeAsAll[T any](do.Injector) []T`: This function is a variant of `do.InvokeAsAll` that panics if any service cannot be found or created.

🚀 Lazy services are loaded in invocation order.

🐎 Lazy service invocation is protected against concurrent loading.

🧙‍♂️ When multiple [scopes](../container/scope.md) are assembled into a big application, the service lookup is recursive from the current nested scope to the root scope.

:::warning

Circular dependencies are not allowed. Services must be invoked in a Directed Acyclic Graph way.

:::

## Example

**Play: https://go.dev/play/p/9JuTQhLGIlh**

```go
type MyService struct {
    IP string
}

func main() {
    i := do.New()

    do.ProvideNamedValue(i, "config.ip", "127.0.0.1")
    do.Provide(i, func(i do.Injector) (*MyService, error) {
        return &MyService{
            IP: do.MustInvokeNamed[string](i, "config.ip"),
        }, nil
    })

    myService, err := do.Invoke[*MyService](i)
}
```

## Auto-magically load a service

You can also use the `do.InvokeStruct` function to auto-magically provide a service with its dependencies. The fields can be either exported or not.

The `do:""` tag indicates the DI must infer the service name from its type (equivalent to `do.Invoke[*logrus.Logger](i)`).

**Play: https://go.dev/play/p/Rqa4RCjThoI**

```go
type MyService struct {
  // injected automatically
  serverPort             int                     `do:"config.listen_port"`
  logger                 *logrus.Logger          `do:""`
  postgresqlClient       *PostgreSQLClient       `do:""`
  dataProcessingService  *DataProcessingService  `do:""`

  // other things, not related to DI
  mu sync.Mutex
}
```

Then add `*MyService` to the list of available services.

```go
do.Provide[*MyService](injector, func (i do.Injector) (*MyService, error) {
  return do.InvokeStruct[MyService](i)
})
// or
do.Provide[*MyService](i, do.InvokeStruct[MyService])
```

### Implicit aliasing behavior with InvokeStruct

When a field uses an empty tag value (eg: `do:""`) and no service is registered under the field type, the injector falls back to finding the first service whose type is assignable to the field type (same resolution strategy as `do.InvokeAs[T]`).

Implications:

- Prefer explicit names when multiple assignable services exist to avoid ambiguity.
- This fallback only applies when the tag key is present and empty; a missing tag does nothing.
- The struct tag key can be customized via `do.InjectorOpts.StructTagKey`.

:::info

Nested structs are not supported.

:::

:::warning

This feature relies on reflection and is therefore not recommended for performance-critical code or serverless environments. Please do your due diligence with proper benchmarks.

:::

## Error handling

Any panic during lazy loading is converted into a Go `error`.

An error is returned on missing service.

## Invoke from Provider

A service might rely on other services. In that case, you should invoke dependencies in the service provider instead of storing the injector for later.

In this way, if a service is not found, the resolution will report an error on application start, instead of runtime.

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

## Bulk service invocation

When you need to work with multiple services that implement the same interface, you can use `do.InvokeAsAll` to retrieve all matching services as a slice:

```go
type Database interface {
    Name() string
    Connect() error
}

type PostgresDB struct {
    name string
}

func (p *PostgresDB) Name() string { return p.name }
func (p *PostgresDB) Connect() error { return nil }

type MySQLDB struct {
    name string
}

func (m *MySQLDB) Name() string { return m.name }
func (m *MySQLDB) Connect() error { return nil }

func main() {
    i := do.New()

    // Register multiple database implementations
    do.Provide(i, func(i do.Injector) (*PostgresDB, error) {
        return &PostgresDB{name: "postgres"}, nil
    })
    do.Provide(i, func(i do.Injector) (*MySQLDB, error) {
        return &MySQLDB{name: "mysql"}, nil
    })

    // Invoke all databases
    databases, err := do.InvokeAsAll[Database](i)
    if err != nil {
        log.Fatal(err)
    }

    // databases contains both PostgresDB and MySQLDB instances
    // in deterministic order (sorted by service name)
    for _, db := range databases {
        fmt.Printf("Connecting to %s database\n", db.Name())
        db.Connect()
    }
}
```

### Key characteristics of InvokeAsAll

- **Returns a slice**: `[]T` instead of `T`
- **Deterministic ordering**: Services sorted alphabetically by registration name
- **Partial failure handling**: Returns successfully invoked services even if some fail
- **Empty results**: Returns empty slice (not error) when no services match
- **Scope inheritance**: Finds services across the entire scope hierarchy

Example:
```go
// Register multiple storage backends
do.Provide(i, NewS3Storage)
do.Provide(i, NewLocalStorage)

// Invoke all storage services
storages, err := do.InvokeAsAll[Storage](i)
// Returns []Storage{S3Storage, LocalStorage} in deterministic order
```

### Use cases

`InvokeAsAll` is particularly useful for:

- **Multiple database connections**: PostgreSQL, MySQL, MongoDB instances
- **Multiple message queues**: Redis, RabbitMQ, Kafka processors
- **Multiple storage backends**: S3, local filesystem, database storage
- **Plugin systems**: Multiple implementations of the same interface
- **Load balancing**: Multiple instances of the same service type

### Error handling

Unlike `InvokeAs` which returns an error when no service is found, `InvokeAsAll` treats "no services found" as a valid empty result. It only returns an error if:

- A service fails to instantiate
- A circular dependency is detected
- A type assertion fails during invocation

This makes `InvokeAsAll` more suitable for scenarios where having zero services of a given type is acceptable.
