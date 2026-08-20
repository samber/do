---
title: samber/do vs uber/dig
description: Compare samber/do and uber/dig for Go dependency injection — generics vs reflection, error handling, and when to pick each.
sidebar_position: 3
---

# samber/do vs uber/dig

**Verdict**: pick `uber/dig` if you specifically want a minimal, unopinionated reflection-based container (for example, as the resolution engine under your own framework). Pick `samber/do` if you want the container itself to also handle lifecycle — health checks, graceful shutdown, scopes — with compile-time type safety instead of reflection.

## At a glance {#at-a-glance}

|                           | `samber/do`                                          | `uber/dig`                                |
| ------------------------- | ---------------------------------------------------- | ----------------------------------------- |
| **Mechanism**             | Go 1.18+ generics                                    | Reflection                                |
| **Type safety**           | Compile-time                                         | Runtime                                   |
| **API style**             | Fluent, generic functions (`do.Provide[T]`)          | Builder pattern (`container.Provide(fn)`) |
| **Error handling**        | Errors from `Invoke()` (or panics with `MustInvoke`) | Errors from `Provide()` and `Invoke()`    |
| **Service lifecycle**     | Health checks, graceful shutdown, hooks              | None built-in                             |
| **Scoped services**       | Full scope tree with visibility rules                | Limited (no native scoping)               |
| **Struct-tag injection**  | Optional, opt-in (`do.InvokeStruct`)                 | `dig.In` / `dig.Out` (common pattern)     |
| **Debugging tools**       | Scope tree, dependency graph, Web UI                 | `dig.Visualize` (DOT graph)               |
| **External dependencies** | None                                                 | None                                      |

## API style {#api-style}

`uber/dig` uses a builder-style container where every constructor is registered via `Provide` and consumed via `Invoke`, with both returning errors that must be checked individually:

```go
container := dig.New()

if err := container.Provide(NewDatabase); err != nil {
    log.Fatal(err)
}
if err := container.Provide(NewUserService); err != nil {
    log.Fatal(err)
}

err := container.Invoke(func(app *App) {
    app.Run()
})
```

`samber/do` uses generic functions parameterized by the concrete type, so most registration mistakes are caught by the compiler instead of surfacing as a `Provide()` error at runtime:

```go
injector := do.New()

do.Provide(injector, NewDatabase)
do.Provide(injector, NewUserService)

app, err := do.Invoke[*App](injector)
```

Constructor signatures differ too: `dig` constructors take their dependencies as plain parameters (`func(db *Database) *UserService`), resolved by matching types via reflection. `samber/do` constructors take the injector itself and call `do.MustInvoke` for each dependency (`func(i do.Injector) (*UserService, error)`), which keeps the dependency list explicit and the construction function itself easy to unit test without a container.

## Type safety and error handling {#type-safety-and-error-handling}

`dig` resolves constructor parameters by reflecting on function signatures at `Invoke` time; a typo'd type or missing provider fails at runtime, often with an error message several levels deep in the dependency chain. `samber/do`'s `Invoke[T]` is parameterized directly by the type you want, so the vast majority of "wrong type" mistakes are compiler errors well before the program runs — the remaining runtime errors are limited to genuinely dynamic conditions (a service that was never registered, or a health check that failed).

## Lifecycle, scopes, and debugging {#lifecycle-scopes-and-debugging}

This is the biggest functional gap between the two. `dig` is intentionally minimal: it resolves a graph and calls constructors, with no concept of scopes, health checks, or shutdown ordering, and no built-in per-service debugging beyond `dig.Visualize`. `samber/do` builds these in: a [scope tree](../container/scope.md) for module visibility, a [`Healthchecker` interface](../service-lifecycle/healthchecker.md) with configurable parallelism, a [`Shutdowner` interface](../service-lifecycle/shutdowner.md) for dependency-ordered graceful shutdown, and dedicated [debugging tools](../troubleshooting/scope-tree.md) (text dump, dependency explain, and a [Web UI](../troubleshooting/web-ui.md)) beyond what `dig.Visualize` offers.

## Testing {#testing}

Both containers support overriding providers for tests. `samber/do` centers this on [`Clone()`](../container/clone.md) plus `do.Override`, so a test injector shares registrations with production code but swaps specific services without touching the container's structure.

## Choose `uber/dig` if: {#choose-uberdig-if}

- You want the smallest possible reflection-based container, likely as a building block inside your own framework (as `uber/fx` does).
- You don't need scopes, health checks, or built-in graceful shutdown.

## Choose `samber/do` if: {#choose-samberdo-if}

- You want compile-time type safety on service resolution instead of runtime reflection.
- You need scopes, health checks, and graceful shutdown as part of the container itself, not layered on top separately.

Ready to switch? Follow the [step-by-step migration guide from Uber Dig](../migrating/migrating-from-dig.md).
