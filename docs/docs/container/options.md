---
title: Container options
description: Container options
sidebar_position: 2
---

# Container

## Default options

The simplest way to start is to use the default parameters:

```go
import "github.com/samber/do/v2"

injector := do.New()
```

## Global container

For a quick start, you may use the default global container. This is highly discouraged in production.

```go
import "github.com/samber/do/v2"

do.Provide(nil, ...)
do.Invoke(nil, ...)

// equal to:

do.Provide(do.DefaultRootScope, ...)
do.Invoke(do.DefaultRootScope, ...)
```

## Register services on container initialization

The services can be assembled into a package, and then, imported all at once into a new container.

```go
// pkg/stores/package.go

var Package = do.Package(
    do.Lazy(NewPostgresqlConnectionService),
    do.Lazy(NewUserRepository),
    do.Lazy(NewArticleRepository),
    do.EagerNamed("repository.logger", slog.New(slog.NewTextHandler(os.Stdout, nil))),
)
```

```go
// cmd/main.go

import (
    "example/pkg/stores"
    "example/pkg/handlers"
)

injector := do.New(stores.Package)
```

## Custom options

```go
import "github.com/samber/do/v2"

injector := do.NewWithOps(&do.InjectorOpts{
    HookAfterRegistration func(scope *do.Scope, serviceName string) {
        // ...
    },
    HookAfterShutdown     func(scope *do.Scope, serviceName string) {
        // ...
    },

    Logf func(format string, args ...any) {
        // ...
    },

    HealthCheckParallelism:   100,
    HealthCheckGlobalTimeout: 1 * time.Second,
    HealthCheckTimeout:       100 * time.Millisecond,
})
```
