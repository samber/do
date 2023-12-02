---
title: "Scopes (a.k.a modules)"
description: Group services into independent modules
sidebar_position: 1
---

# Scopes (a.k.a modules)

A `Scope` can be viewed as a module of an application, with restricted visibility. We advocate for each project possessing a root scope for common code, and shifting business logic to dedicated Scopes.

The Scopes are structured with a parent (root scope or other scopes), numerous services, and potentially a few children Scopes as well.

Root scope owns options and is the only scope that supports cloning.

Services from a scope can invoke services available locally or in ancestors' scopes. Therefore, a single Service can be instantiated many times in different branches of the scope tree, without conflict.

A chain of service invocation instantiates many virtual scopes, to track dependency cycles.

Scopes are almost invisible to developers: services keep using Injector API without awareness of the underlying implementation, whether it's a root scope, scope or virtual scope.

## New scope

A root scope is created when calling `do.New()`. Multiple layers of child scopes can be added. Each nested scope shares the same root scope.

```go
// root scope
injector := do.New()

// nested scopes
driverModule := injector.Scope("driver")
passengersModule := injector.Scope("passengers")
```

Then, inject services into respective scopes. Some services can depend on parent services.

```go
// inject services to root scope
do.Provide(injector, NewEngine)

// inject *SteeringWheel service to "driver" scope
do.Provide(driverModule, func(i do.Injector) (*SteeringWheel, error) {
    return &SteeringWheel{
        // invoke *Engine service from parent scope
        Engine: do.MustInvoke[*Engine](i),
    }, nil
})

// inject many *Passenger services to "passengers" scope
do.ProvideNamed(passengersModule, "passenger-1", func(i do.Injector) (*Passenger, error) {
    return &Passenger{ ... }, nil
})
do.ProvideNamed(passengersModule, "passenger-2", func(i do.Injector) (*Passenger, error) {
    return &Passenger{ ... }, nil
})
do.ProvideNamed(passengersModule, "passenger-3", func(i do.Injector) (*Passenger, error) {
    return &Passenger{ ... }, nil
})
```

At this step, services have not been instantiated. Let's invoke them:

```go
// instantiate *SteeringWheel and *Engine
svc := do.MustInvoke[*SteeringWheel](driverModule)

// instantiate passengers
svc := do.MustInvokeNamed[*Passenger](passengersModule, "passenger-1")
svc := do.MustInvokeNamed[*Passenger](passengersModule, "passenger-2")
svc := do.MustInvokeNamed[*Passenger](passengersModule, "passenger-3")
```

## Debug

A debugging toolchain has been added to illustrate the service [dependency](../debugging/service-dependencies.md) chain and [scope tree](../debugging/scope-tree.md).

## Examples

See examples in the [project repository](https://github.com/samber/do).
