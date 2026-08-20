---
title: samber/do vs google/wire
description: Compare samber/do and google/wire for Go dependency injection — runtime generics vs compile-time code generation, and when to pick each.
sidebar_position: 2
---

# samber/do vs google/wire

**Verdict**: pick `google/wire` if you want zero runtime cost and are fine running a code-generation step (`wire` CLI) on every dependency change, with a fully static graph known at compile time. Pick `samber/do` if you want compile-time type safety _without_ a build step, plus runtime flexibility (conditional registration, scopes, overriding services at runtime) that generated code can't easily express.

## At a glance {#at-a-glance}

|                          | `samber/do`                                       | `google/wire`                                          |
| ------------------------ | ------------------------------------------------- | ------------------------------------------------------ |
| **Mechanism**            | Go 1.18+ generics, resolved at runtime            | Code generation (`wire` CLI), resolved at compile time |
| **Build step required**  | No                                                | Yes — regenerate `wire_gen.go` on every graph change   |
| **Type safety**          | Compile-time (via generics)                       | Compile-time (generated code is plain Go)              |
| **Runtime flexibility**  | Full — conditional registration, override, scopes | None — the graph is fixed at generation time           |
| **Runtime cost**         | Small map lookups per `Invoke`                    | None — generated code is direct function calls         |
| **Lifecycle management** | Health checks, graceful shutdown, hooks           | None — you wire cleanup manually                       |
| **Debugging tools**      | Scope tree, dependency graph, Web UI              | Read the generated `wire_gen.go`                       |
| **Tooling dependency**   | None                                              | `wire` binary must be installed and run in CI          |

## API style {#api-style}

`wire` doesn't run at runtime at all — you write "injector" functions annotated with `wire.Build`, and the `wire` command generates the real implementation as plain Go:

```go
//go:build wireinject

func InitializeApp() (*App, error) {
    wire.Build(NewDatabase, NewUserService, NewApp)
    return &App{}, nil // never actually executed
}
```

Running `wire` produces `wire_gen.go`, a hand-written-looking file with no reflection and no `wire` import at runtime. `samber/do` skips the generation step entirely — the same constructor functions are registered and resolved directly:

```go
injector := do.New()

do.Provide(injector, NewDatabase)
do.Provide(injector, NewUserService)

app, err := do.Invoke[*App](injector)
```

## Compile-time safety vs runtime flexibility {#compile-time-safety-vs-runtime-flexibility}

Both libraries catch dependency-graph mistakes before your code reaches production, but at different points. `wire` fails at `go generate` time if a dependency is missing or ambiguous — you can't even compile without fixing it. `samber/do` fails at `go build` time for type mismatches (thanks to generics) and at `do.Invoke` time for missing registrations — slightly later, but still before the request path in most apps, since services are typically invoked once at startup.

The tradeoff is flexibility: `wire`'s graph is fixed once generated, so conditional wiring (feature flags, per-environment services, request-scoped overrides) requires generating multiple injector functions or falling back to manual code. `samber/do`'s graph is built at runtime, so you can register services conditionally, [override](../service-registration/lazy-loading.md#hot-service-replacement) them for tests, or attach [scopes](../container/scope.md) dynamically — patterns that don't fit `wire`'s static model.

## Lifecycle and health checks {#lifecycle-and-health-checks}

`wire` has no opinion on lifecycle — it only wires constructors together. Cleanup functions (the second return value some `wire` providers support) are called once, in `wire_gen.go`, with no health check or dependency-aware shutdown ordering. `samber/do` provides both: a [`Healthchecker` interface](../service-lifecycle/healthchecker.md) for on-demand health probing, and a [`Shutdowner` interface](../service-lifecycle/shutdowner.md) for reverse-order graceful shutdown across the whole scope tree.

## Testing {#testing}

`wire` tests typically generate a second injector function (e.g. `InitializeTestApp`) wired to mocks — meaning a second `wire.Build` call and a second generated file to keep in sync. `samber/do` avoids the duplication: [`Clone()`](../container/clone.md) an existing injector and `do.Override` the services you need to fake, in the test file itself.

## Choose `google/wire` if: {#choose-googlewire-if}

- You want zero runtime dependency-injection cost — the generated code is indistinguishable from hand-written Go.
- Your dependency graph is static and known entirely at compile time.
- You're comfortable adding `wire` to your build/CI pipeline and committing generated code.

## Choose `samber/do` if: {#choose-samberdo-if}

- You don't want a code-generation step in your build.
- You need runtime flexibility: conditional registration, scopes, or swapping services without regenerating code.
- You want built-in health checks and graceful shutdown, not just constructor wiring.

Already using `google/wire`? See the [migration guide from Google Wire](../migrating/migrating-from-wire.md) for a step-by-step conversion of provider sets and bindings.
