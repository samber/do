---
title: Clone
description: Clone your global DI container
sidebar_position: 3
---

# Clone

Cloning an injector can be very useful for test purposes.

```go
injector := do.New()

Provide[*Car](i, NewCar)
Provide[*Engine](i, NewEngine)

// reset scope
injector = injector.Clone()
```

## Clone with options

```go
injector = injector.CloneWithOpts(&do.InjectorOpts{
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
