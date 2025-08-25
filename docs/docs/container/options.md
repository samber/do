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

// equivalent to:

do.Provide(do.DefaultRootScope, ...)
do.Invoke(do.DefaultRootScope, ...)
```

## Register services on container initialization

The services can be assembled into a package, and then, imported all at once into a new container.

```go
// pkg/stores/package.go

var Package = do.Package(
    do.Lazy(NewPostgreSQLConnectionService),
    do.Lazy(NewUserRepository),
    do.Lazy(NewArticleRepository),
    do.EagerNamed("repository.logger", slog.New(slog.NewTextHandler(os.Stdout, nil))),
)
```

```go
// cmd/main.go

import "example/pkg/stores"

injector := do.New(stores.Package)
```

## Custom options

```go
import "github.com/samber/do/v2"

injector := do.NewWithOpts(&do.InjectorOpts{
    HookBeforeRegistration: []func(scope *do.Scope, serviceName string){},
    HookAfterRegistration:  []func(scope *do.Scope, serviceName string){},
    HookBeforeInvocation:   []func(scope *do.Scope, serviceName string){},
    HookAfterInvocation:    []func(scope *do.Scope, serviceName string, err error){},
    HookBeforeShutdown:     []func(scope *do.Scope, serviceName string){},
    HookAfterShutdown:      []func(scope *do.Scope, serviceName string, err error){},

    Logf: func(format string, args ...any) {
        // ...
    },

    HealthCheckParallelism:   100,
    HealthCheckGlobalTimeout: 1 * time.Second,
    HealthCheckTimeout:       100 * time.Millisecond,
})
```

### Add hooks at runtime

Hooks can also be registered after the injector is created using helper methods on the root scope. These append to the corresponding hook lists in `do.InjectorOpts` and apply to subsequent registrations/invocations/shutdowns.

```go
import "github.com/samber/do/v2"

injector := do.New()

// Registration hooks
injector.AddBeforeRegistrationHook(func(scope *do.Scope, serviceName string) {
    // ...
})
injector.AddAfterRegistrationHook(func(scope *do.Scope, serviceName string) {
    // ...
})

// Invocation hooks
injector.AddBeforeInvocationHook(func(scope *do.Scope, serviceName string) {
    // ...
})
injector.AddAfterInvocationHook(func(scope *do.Scope, serviceName string, err error) {
    // ...
})

// Shutdown hooks
injector.AddBeforeShutdownHook(func(scope *do.Scope, serviceName string) {
    // ...
})
injector.AddAfterShutdownHook(func(scope *do.Scope, serviceName string, err error) {
    // ...
})
```
