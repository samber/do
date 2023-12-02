---
title: Injector options
description: Injector options
sidebar_position: 2
---

# Injector options

## Default container

The simplest way to start is to use the default parameters:

```go
import "github.com/samber/do/v2"

injector := do.New()
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
})
```
