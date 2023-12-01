---
title: Service health check
description: Service health check
sidebar_position: 1
---

# Service health check

If your service relies on a dependency, you might want to periodically check its state.

When the `injector.HealthCheck()` function is called, the framework triggers `HealthCheck` method of each service implementing a `Healthchecker` interface, in reverse invocation order.

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
