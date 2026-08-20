---
title: Service registration
description: List provided and invoked services on a samber/do injector, and inspect the inferred service name for a given type.
sidebar_position: 3
---

# Service registration

`injector.ListProvidedServices()` and `injector.ListInvokedServices()` return every service registered on an injector and, separately, every service that has actually been instantiated so far. Comparing the two lists is the fastest way to spot a service that's registered but never used, or one that's missing entirely.

## Spec {#spec}

```go
injector.ListProvidedServices() []do.ServiceDescription
injector.ListInvokedServices() []do.ServiceDescription

type ServiceDescription struct {
	ScopeID   string
	ScopeName string
	Service   string
}
```

## Service name {#service-name}

Each service is identified in the DI container by a slug.

When using implicit naming, the `do` framework infers the service name from the type, by using the [go-type-to-string](https://github.com/samber/go-type-to-string) library.

For debugging purposes, you might want to print the service name.

**Play: https://go.dev/play/p/g549GqBbj-n**

```go
i := do.New()

do.Provide(i, func(i do.Injector) (*MyService, error) {
    return &MyService{}, nil
})

println(do.NameOf[*MyService]())
```

Output:

```txt
// *github.com/samber/example.MyService
```

## Provided services {#provided-services}

For debugging purposes, the list of services provided to the container can be printed:

**Play: https://go.dev/play/p/e_oxd7b-q9h**

```go
i := do.New()

do.Provide(i, func(i do.Injector) (*MyService, error) {
    return &MyService{}, nil
})
do.ProvideNamed(i, "a-number", 42)

services := i.ListProvidedServices()
println(services)
```

Output:

```json
[
  {
    "ScopeID": "xxxxx",
    "ScopeName": "[root]",
    "Service": "*github.com/samber/example.MyService"
  },
  {"ScopeID": "xxxxx", "ScopeName": "[root]", "Service": "a-number"}
]
```

## Invoked services {#invoked-services}

For debugging purposes, the list of invoked services can be printed:

**Play: https://go.dev/play/p/pJcJGOF5zeK**

```go
i := do.New()

do.Provide(i, func(i do.Injector) (*MyService, error) {
    return &MyService{}, nil
})
do.ProvideNamed(i, "a-number", 42)

services := i.ListInvokedServices()
println(services)
```

Output:

```json
[{"ScopeID": "xxxxx", "ScopeName": "[root]", "Service": "a-number"}]
```

In the example above, the lazy-loaded service `*MyService` has not been invoked.

## Finding registered-but-unused services {#finding-registered-but-unused-services}

Diffing `ListProvidedServices()` against `ListInvokedServices()` is a quick way to find dead registrations during cleanup: anything present in the first list but absent from the second was registered but never actually invoked in that run. This doesn't necessarily mean the service is unused in general — a service invoked only on a rarely-hit code path won't appear until that path runs — but it's a good starting point for auditing a container after a refactor.

See also the [scope tree](./scope-tree.md) to visualize how these services relate across scopes, and [service dependencies](./service-dependencies.md) to inspect a single service's dependency chain.
