---
title: Package loading
description: Package loading groups multiple service registrations.
sidebar_position: 4
---

# Package loading

Package loading groups multiple service registrations.

## Registration

The services can be assembled into a package, and then, exported all at once.

```go
// pkg/stores/package.go

var Package = do.Package(
    do.Lazy(NewPostgresqlConnectionService),
    do.Lazy(NewUserRepository),
    do.Lazy(NewArticleRepository),
)
```

```go
// pkg/observability/package.go

var Package = do.Package(
    do.Eager(slog.New(slog.NewTextHandler(os.Stdout, nil))),
    do.EagerNamed("prometheus.collector", DefaultMetricCollector),
)
```

```go
// cmd/main.go

import (
    "example/pkg/stores"
    "example/pkg/observability"
    "example/pkg/handlers"
)

func main() {
    injector := do.New(stores.Package, observability.Package)
    // 
    // equivalent to:
    // 
    // injector := do.New()
    // stores.Package(injector)
    // observability.Package(injector)

    // ...

    // optional scope:
    scope := injector.Scope("handlers", handlers.Package)

    // ...
}
```

The traditional vocab can be translated for service registration:

- `Provide[T](Injector)` -> `Lazy(T)`
- `ProvideNamed[T](Injector, string)` -> `LazyNamed(string, T)`
- `ProvideValue(Injector, T)` -> `Eager(T)`
- `ProvideNamedValue[T](Injector, string, T)` -> `EagerNamed(string, T)`
- `ProvideTransient[T](Injector)` -> `Transient(T)`
- `ProvideNamedTransient[T](Injector, string)` -> `TransientNamed(string, T)`
- `As[Initial, Alias](Injector)` -> `Bind[Initial, Alias]()`
- `AsNamed[Initial, Alias](Injector, string, string)` -> `BindNamed[Initial, Alias](string, string)`
