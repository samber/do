---
title: Injector options
description: Injector options
sidebar_position: 2
---

# Injector

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

Provide(nil, ...)
Invoke(nil, ...)

// equal to:

Provide(do.DefaultRootScope, ...)
Invoke(do.DefaultRootScope, ...)
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
