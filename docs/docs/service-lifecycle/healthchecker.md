---
title: Service health check
description: Service health check
sidebar_position: 1
---

# Service health check

If your service relies on a dependency, you might want to periodically check its state.

When the `do.HealthCheck[type]()` or the `injector.HealthCheck()` function is called, the framework triggers the `HealthCheck` method of each service implementing the `do.Healthchecker` interface, in reverse invocation order.

🕵️ Service health can be checked individually or globally. Requesting a health check on a nested scope will run checks on ancestors.

Lazy services that were not invoked are not checked.

## Trigger health check

A health check can be triggered for a root injector:

```go
// returns the status (error or nil) for each service
injector.HealthCheck() map[string]error
injector.HealthCheckWithContext(context.Context) map[string]error
```

...on a single service:

```go
// returns error on failure
do.HealthCheck[T any](do.Injector) error
do.HealthCheckWithContext[T any](context.Context, do.Injector) error
do.HealthCheckNamed[T any](do.Injector, string) error
do.HealthCheckNamedWithContext[T any](context.Context, do.Injector, string) error
```

## Healthchecker interfaces

Your service can implement one of the following signatures:

```go
type Healthchecker interface {
	HealthCheck() error
}

type HealthcheckerWithContext interface {
	HealthCheck(context.Context) error
}
```

Example:

```go
// Ensure at compile-time MyService implements do.HealthcheckerWithContext
var _ do.HealthcheckerWithContext = (*MyService)(nil)

type MyService struct {}

func (*MyService) HealthCheck(context.Context) error {
    // ...
    return nil
}

i := do.New()

Provide(i, ...)
Invoke(i, ...)

ctx := context.WithTimeout(10 * time.Second)
i.HealthCheckWithContext(ctx)
```

## Health check options

The root scope can be created with health check parameters, for controlling parallelism or timeouts.

```go
do.InjectorOpts{
    // ...

    // By default, health checks are triggered concurrently.
    // HealthCheckParallelism==1 will trigger sequential checks.
    HealthCheckParallelism    uint

    // When many services are checked at the same time.
    HealthCheckGlobalTimeout: time.Duration
    HealthCheckTimeout:       time.Duration
}
```

Example:

```go
type MyPostgreSQLConnection struct {
    DB *sql.DB
}

func (pg *MyPostgreSQLConnection) HealthCheck() error {
    return pg.DB.Ping() // <- might be very slow
}

i := do.NewWithOpts(&do.InjectorOpts{
    HealthCheckParallelism:   100,
    HealthCheckGlobalTimeout: 1 * time.Second,
    HealthCheckTimeout:       100 * time.Millisecond,
})

Provide(i, NewMyPostgreSQLConnection)
_ = MustInvoke(i, *MyPostgreSQLConnection)

status := i.HealthCheckWithContext(ctx)
// {
//     "*github.com/samber/example.MyPostgreSQLConnection": "DI: health check timeout: context deadline exceeded",
// }
```
