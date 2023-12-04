---
title: Service registration
description: Learn how to troubleshoot service registration
sidebar_position: 3
---

# Service registration

## Service name

Each service is identified in the DI container by a slug.

When using implicit naming, the `do` framework infers the service name from the type, by using the [go-type-to-string](https://github.com/samber/go-type-to-string) library.

For debugging purposes, you might want to print the service name.

```go
i := do.New()

do.Provide(i, func(i do.Injector) (*MyService, error) {
    return &MyService{}, nil
})

println(do.Name[*MyService](i))
// *github.com/samber/example.MyService
```

## Provided services

For debugging purposes, the list of services provided to the container can be printed:

```go
i := do.New()

do.Provide(i, func(i do.Injector) (*MyService, error) {
    return &MyService{}, nil
})
do.ProvideNamed(i, "a-number", 42)

services := i.ListProvidedServices()
println(services)
// []{
//    {ScopeID: "xxxxx", ScopeName: "[root]", Service: "*github.com/samber/example.MyService"},
//    {ScopeID: "xxxxx", ScopeName: "[root]", Service: "a-number"},
// }
```

## Invoked services

For debugging purposes, the list of invoked services can be printed:

```go
i := do.New()

do.Provide(i, func(i do.Injector) (*MyService, error) {
    return &MyService{}, nil
})
do.ProvideNamed(i, "a-number", 42)

services := i.ListInvokedServices()
println(services)
// []{
//    {ScopeID: "xxxxx", ScopeName: "[root]", Service: "a-number"},
// }
```

In the example above, the lazy-loaded service `*MyService` has not been invoked.
