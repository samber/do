---
title: Service dependencies
description: Understand your dependency graph
sidebar_position: 2
---

# Service dependencies

## Spec

```go
do.DescribeService[T any](do.Injector) (string, bool)
do.DescribeNamedService(do.Injector, string) (string, bool)

do.ExplainService[T any](do.Injector) (dependencies []do.EdgeService, dependents []do.EdgeService, found bool)
do.ExplainNamedService(do.Injector, string) (dependencies []do.EdgeService, dependents []do.EdgeService, found bool)
```

## Example

Having the following services:

```go
i := do.New()
// create scopes
scope := i.Scope("scope-child")

// inject many services
do.ProvideNamed(i, "SERVICE-A1", provider1)
do.ProvideNamed(i, "SERVICE-A2", provider2)
do.ProvideNamed(i, "SERVICE-B", provider3)
do.ProvideNamed(scope, "SERVICE-C1", provider4)
do.ProvideNamed(scope, "SERVICE-C2", provider5)
do.ProvideNamed(scope, "SERVICE-D", provider6)
do.ProvideNamed(scope, "SERVICE-E", provider7)
do.ProvideNamed(scope, "SERVICE-F", provider8)
do.ProvideNamed(scope, "SERVICE-G", provider9)

// load SERVICE-G and its dependencies
_ = do.MustInvokeNamed[*MyService](scope, "SERVICE-G")
```

### Service description

```go
output, found := do.DescribeNamedService[*MyService](scope, "SERVICE-E")
if found {
    println(output)
} else {
    println("service not found")
}
```

Output:

```txt
Scope ID: 2a821628-7175-4b6d-a2ea-de8b82a203de
Scope name: scope-child

Service name: SERVICE-E
Service type: lazy
Invoked: github.com/acme/example/main.go:main:42

Dependencies:
* SERVICE-D from scope scope-child
  * SERVICE-C1 from scope scope-child
    * SERVICE-B from scope [root]
      * SERVICE-A1 from scope [root]
      * SERVICE-A2 from scope [root]
  * SERVICE-C2 from scope scope-child
    * SERVICE-B from scope [root]
      * SERVICE-A1 from scope [root]
      * SERVICE-A2 from scope [root]

Dependents:
* SERVICE-F from scope scope-child
  * SERVICE-G from scope scope-child
```

### Dependencies

```go
dependencies, _, found := do.ExplainNamedService[*MyService](scope, "SERVICE-E")
if found {
    println(dependencies)
} else {
    println("service not found")
}
```

Output:

```txt
[
    { ScopeID: "6cbfc332-0276-4c28-b0c3-d6256210e4d6", ScopeName: "scope-child", Service: "SERVICE-D" },
    { ScopeID: "6cbfc332-0276-4c28-b0c3-d6256210e4d6", ScopeName: "scope-child", Service: "SERVICE-C1" },
    { ScopeID: "6cbfc332-0276-4c28-b0c3-d6256210e4d6", ScopeName: "scope-child", Service: "SERVICE-C2" },
    { ScopeID: "2a821628-7175-4b6d-a2ea-de8b82a203de", ScopeName: "[root]",      Service: "SERVICE-B" },
    { ScopeID: "2a821628-7175-4b6d-a2ea-de8b82a203de", ScopeName: "[root]",      Service: "SERVICE-A1" },
    { ScopeID: "2a821628-7175-4b6d-a2ea-de8b82a203de", ScopeName: "[root]",      Service: "SERVICE-A2" },
]
```

### Dependents

```go
_, dependents, found := do.ExplainNamedService[*MyService](scope, "SERVICE-E")
if found {
    println(dependents)
} else {
    println("service not found")
}
```

Output:

```txt
[
    { ScopeID: "6cbfc332-0276-4c28-b0c3-d6256210e4d6", ScopeName: "scope-child", Service: "SERVICE-F" },
    { ScopeID: "6cbfc332-0276-4c28-b0c3-d6256210e4d6", ScopeName: "scope-child", Service: "SERVICE-G" },
]
```
