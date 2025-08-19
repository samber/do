---
title: Service dependencies
description: Understand your dependency graph
sidebar_position: 2
---

# Service dependencies

## Dependency tree

### Spec

```go
do.ExplainService[T any](do.Injector) (do.ExplainServiceOutput, bool)
do.ExplainNamedService(do.Injector, string) (do.ExplainServiceOutput, bool)
```

### Print tree

```go
debug, found := do.ExplainNamedService[*MyService](scope, "SERVICE-E")
if found {
    println(debug.String())
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

### Print dependencies or dependents

```go
description, found := do.ExplainNamedService[*MyService](scope, "SERVICE-E")
if found {
    println(description.Dependencies)
} else {
    println("service not found")
}
```

Output:

```txt
[
    { ScopeID: "6cbfc332-0276-4c28-b0c3-d6256210e4d6", ScopeName: "scope-child", Service: "SERVICE-F", Recursive: [...] },
    { ScopeID: "6cbfc332-0276-4c28-b0c3-d6256210e4d6", ScopeName: "scope-child", Service: "SERVICE-G", Recursive: [...] },
]
```
