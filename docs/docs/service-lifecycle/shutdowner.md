---
title: Services shutdown
description: Graceful service shutdown
sidebar_position: 2
---

# Graceful service shutdown

If your service brings a state such as a connection pool, a buffer, etc... you might want to close or flush it gracefully on the application exit.

When the `injector.Shutdown()` function is called, the framework triggers `Shutdown` method of each service implementing a `Shutdowner` interface, in reverse invocation order.

## Trigger shutdown

A shutdown can be triggered on a root injector:

```go
// on demand
injector.Shutdown() error
injector.ShutdownWithContext(context.Context) error

// on signal
injector.ShutdownOnSignals(...os.Signal) (os.Signal, error)
injector.ShutdownOnSignalsWithContext(context.Context, ...os.Signal) (os.Signal, error)
```

...on a single service:

```go
// returns error on failure
do.Shutdown[T any](do.Injector) error
do.ShutdownWithContext[T any](context.Context, do.Injector) error
do.ShutdownNamed[T any](do.Injector, string) error
do.ShutdownNamedWithContext[T any](context.Context, do.Injector, string) error

// panics on failure
do.MustShutdown[T any](do.Injector)
do.MustShutdownWithContext[T any](context.Context, do.Injector)
do.MustShutdownNamed[T any](do.Injector, string)
do.MustShutdownNamedWithContext[T any](context.Context, do.Injector, string)
```

## Shutdowner interfaces

Your service can implement one of the following signatures:

```go
type Shutdowner interface {
	Shutdown()
}

type ShutdownerWithError interface {
	Shutdown() error
}

type ShutdownerWithContext interface {
	Shutdown(context.Context)
}

type ShutdownerWithContextAndError interface {
	Shutdown(context.Context) error
}
```

Example:

```go
type MyService struct {}

func (*MyService) Shutdown(context.Context) error {
    // ...
    return nil
}

i := do.New()

Provide(i, ...)
Invoke(i, ...)

ctx := context.WithTimeout(10 * time.Second)
i.ShutdownWithContext(ctx)
```
