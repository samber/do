---
title: Scope tree
description: Visualize your module tree
sidebar_position: 1
---

# Scope tree

## Spec

```go
do.DescribeInjector(do.Injector) (string, bool)
```

## Example

Having the following services:

```go
// create scopes
i := do.New()
scope := i.Scope("scope-child")
scope0 := i.Scope("scope-0")
scope1a := scope0.Scope("scope-1a")
scope1b := scope0.Scope("scope-1b")
scope2a := scope1a.Scope("scope-2a")
scope2b := scope1a.Scope("scope-2b")

// inject many services
ProvideNamed(i, "SERVICE-A1", provider1)
ProvideNamed(i, "SERVICE-A2", provider1)
ProvideNamed(i, "SERVICE-B", provider2)
ProvideNamed(scope1a, "SERVICE-C1", provider3)
ProvideNamed(scope1a, "SERVICE-C2", provider3)
ProvideNamed(scope1a, "SERVICE-D", provider4)
ProvideNamed(scope1a, "SERVICE-E", provider5)
ProvideNamed(scope1b, "SERVICE-F", provider6)
ProvideNamedTransient[*MyService1](scope2a, "SERVICE-TRANSIENT-SIMPLE", func(i Injector) (*MyService1, error) { return &MyService1{}, nil })
ProvideNamed[*MyService2](scope2a, "SERVICE-LAZY-HEALTH", func(i Injector) (*MyService2, error) { return &MyService2{}, nil })
ProvideNamed[*MyService3](scope2b, "SERVICE-LAZY-SHUTDOWN", func(i Injector) (*MyService3, error) { return &MyService3{}, nil })
ProvideNamedValue[int](scope1a, "SERVICE-EAGER-VALUE", 1)
_ = AsNamed[*MyService2, Healthchecker](scope2a, "SERVICE-LAZY-HEALTH", "SERVICE-ALIAS-HEALTH")

// load some services and their own dependencies
_ = MustInvokeNamed[int](scope1a, "SERVICE-D")
_ = MustInvokeNamed[*MyService2](scope2a, "SERVICE-LAZY-HEALTH")
_ = MustInvokeNamed[*MyService3](scope2b, "SERVICE-LAZY-SHUTDOWN")
```

### Print scope tree

```go
output, found := do.DescribeInjector[int](scope)
if found {
    println(output)
} else {
    println("service not found")
}
```

Output:

```txt
Scope ID: 35d18a30-0cb9-4fad-aa3c-1438937612ad
Scope name: [root]
DAG:
\_ [root] (ID: 35d18a30-0cb9-4fad-aa3c-1438937612ad)
    * 😴 SERVICE-A1
    * 😴 SERVICE-A2
    * 😴 SERVICE-B
    |
    |
     \_ scope-0 (ID: dce6f365-fc52-4407-ac02-a6a857e0cbc9)
         |
         |
         |\_ scope-1a (ID: b9cac0c2-da74-4ee1-a5e0-eb30dca912c3)
         |    * 😴 SERVICE-C1
         |    * 😴 SERVICE-C2
         |    * 😴 SERVICE-D
         |    * 😴 SERVICE-E
         |    * 🔁 SERVICE-EAGER-VALUE
         |    |
         |    |
         |    |\_ scope-2a (ID: 53406825-b7cc-46cc-9baf-ab0319c8a3bd)
         |    |    * 🔗 SERVICE-ALIAS-HEALTH 🏥
         |    |    * 😴 SERVICE-LAZY-HEALTH 🏥
         |    |    * 🏭 SERVICE-TRANSIENT-SIMPLE
         |    |     
         |    |
         |     \_ scope-2b (ID: adc75f0a-a2dc-488f-9457-510e151d9e34)
         |         * 😴 SERVICE-LAZY-SHUTDOWN 🙅
         |          
         |
          \_ scope-1b (ID: a9e3adfc-ad10-4cbd-9b0d-85cfad95d99c)
              * 😴 SERVICE-F
```

Emojis describe service type and capabilities:
- 😴 Lazy service
- 🔁 Eager service
- 🏭 Transient service
- 🔗 Service alias

...and capabilities:
- 🏥 Implements Healthchecker
- 🙅 Implements Shutdowner
