---
title: Clone
description: Clone a samber/do root scope to reset invoked service state between tests while keeping the same service registrations.
sidebar_position: 3
---

# Clone

`Clone()` copies a root scope's service **registrations** into a new root scope, without carrying over any already-**invoked** singleton state. It only works on `*do.RootScope`; regular and virtual scopes cannot be cloned directly.

This is the recommended way to isolate tests: register your services (and mocks) once, then clone the injector for each test case instead of rebuilding it from scratch.

```go
injector := do.New()

Provide[*Car](i, NewCar)
Provide[*Engine](i, NewEngine)

// reset scope
injector = injector.Clone()
```

Because the clone starts with no invoked services, the next call to `do.Invoke` re-runs the provider functions and builds fresh instances — nothing is shared with the original injector.

## Clone with options {#clone-with-options}

`CloneWithOpts` clones a root scope while also overriding its [container options](./options.md), such as the struct tag key used for `do.InvokeStruct`. Combined with `do.Override`, it's the standard pattern for swapping a real dependency with a mock in a single test:

```go
injector := do.New()

Provide[*Car](i, NewCar)
Provide[Engine](i, NewEngine)

// clone
injector = injector.CloneWithOpts(&do.InjectorOpts{
    StructTagKey: "di",
})

// replace Engine by *MockEngine
do.Override[Engine](injector, func (i do.Injector) (Engine, error) {
    return &MockEngine{}, nil
})
```

## When to use Clone {#when-to-use-clone}

- **Table-driven tests**: clone once per test case to avoid state leaking between them.
- **Parallel tests**: each goroutine gets its own instances from a shared registration set.
- **Feature flags / A-B setups**: clone a base injector and override a handful of services per variant.

## Common mistake: cloning the wrong scope {#common-mistake-cloning-the-wrong-scope}

`Clone()` and `CloneWithOpts()` are only available on `*do.RootScope`. If you're holding a child scope (from `injector.Scope("name")`) and try to reset it between tests, clone the root scope instead and re-derive the child scope from the clone — cloning a child scope directly isn't supported, since a scope's identity is tied to its position in the parent's tree.

For mocking whole packages instead of individual services, see [package loading](../service-registration/package-loading.md#testing-and-mocking). For a broader picture of how scopes relate to each other, see [Scopes](./scope.md).
