---
title: samber/do vs uber/fx
description: Compare samber/do and uber/fx for Go dependency injection — generics vs reflection, lifecycle management, health checks, and when to pick each.
sidebar_position: 1
---

# samber/do vs uber/fx

**Verdict**: pick `uber/fx` if you're building a large, modular application that benefits from its batteries-included lifecycle framework and are comfortable with runtime reflection and its dependency footprint (`dig`, `zap`, `multierr`). Pick `samber/do` if you want compile-time-checked dependency resolution through generics, built-in health checks, and zero external dependencies, without giving up lifecycle management.

## At a glance {#at-a-glance}

|                           | `samber/do`                                                | `uber/fx`                                                                       |
| ------------------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------- |
| **Resolution mechanism**  | Go 1.18+ generics                                          | Reflection (via `uber/dig`)                                                     |
| **Type safety**           | Compile-time                                               | Runtime                                                                         |
| **External dependencies** | None                                                       | `dig`, `zap`, `multierr`                                                        |
| **Lifecycle hooks**       | `HookBeforeRegistration`, `HookAfterInvocation`, etc.      | `fx.Lifecycle` (`OnStart`/`OnStop`)                                             |
| **Health checks**         | Built-in `Healthchecker` interface, parallel with timeouts | Not built-in — implement your own                                               |
| **Graceful shutdown**     | Built-in, dependency-aware, reverse order                  | Via `fx.Lifecycle` `OnStop` hooks                                               |
| **Modules / scopes**      | Scope tree with visibility rules                           | `fx.Module`                                                                     |
| **Struct-tag injection**  | Optional, opt-in (`do.InvokeStruct`)                       | `fx.In` / `fx.Out` (common pattern)                                             |
| **Debugging tools**       | Scope tree, dependency graph, Web UI                       | `fx.WithLogger`, DOT graph export                                               |
| **Learning curve**        | Low — mirrors constructor injection you'd write by hand    | Moderate — own vocabulary (`fx.Provide`, `fx.Invoke`, `fx.Module`, annotations) |

## API style {#api-style}

`uber/fx` wraps `uber/dig` and adds an application container (`fx.New`) that owns startup and shutdown. Providers are functions registered with `fx.Provide`, wired together by reflecting on their argument and return types:

```go
app := fx.New(
    fx.Provide(NewDatabase),
    fx.Provide(NewUserService),
    fx.Invoke(func(s *UserService) {}),
)
app.Run()
```

`samber/do` uses an explicit injector value and generic type parameters instead of an implicit global app:

```go
injector := do.New()

do.Provide(injector, NewDatabase)
do.Provide(injector, NewUserService)

service := do.MustInvoke[*UserService](injector)
```

Both APIs use ordinary constructor functions. The difference is _when_ type mismatches surface: with `fx`, a missing or ambiguous dependency fails at `app.Run()` time; with `do`, most of the same mistakes are Go compiler errors because `Invoke[T]` and `Provide[T]` are parameterized by the concrete type.

## Type safety and performance {#type-safety-and-performance}

`fx` resolves the dependency graph via reflection on every `fx.New()` call, walking provider signatures to build the graph. `samber/do` builds the same kind of graph, but service identity and resolution go through generic type parameters resolved at compile time — there's no `reflect.Type` walk on the hot path for `Invoke`/`Provide`. For most applications the difference in raw invocation speed doesn't dominate startup time, but the compile-time checking catches wiring mistakes before code ships instead of at runtime `app.Run()`.

## Lifecycle and health checks {#lifecycle-and-health-checks}

This is where the libraries diverge most in scope. `fx.Lifecycle` gives you `OnStart(ctx)` / `OnStop(ctx)` hooks per component, run in dependency order — good for HTTP servers, background workers, and connection pools. It does not include a health-check abstraction; teams typically build their own on top of `OnStart`.

`samber/do` ships both: [lifecycle hooks](../container/options.md#custom-options) similar in spirit to `fx.Lifecycle`, plus a first-class [`Healthchecker` interface](../service-lifecycle/healthchecker.md) with configurable parallelism and timeouts, and a [`Shutdowner` interface](../service-lifecycle/shutdowner.md) for graceful, dependency-aware shutdown out of the box.

## Testing {#testing}

Both libraries support swapping real dependencies for test doubles. `fx` uses `fxtest.New` and can override providers with `fx.Replace` or `fx.Decorate`. `samber/do` uses [`Clone()`](../container/clone.md) to snapshot an injector's registrations and `do.Override` to swap a specific service — no separate test-only package required.

## Choose `uber/fx` if: {#choose-uberfx-if}

- You're already invested in the Uber Go ecosystem (`zap`, `dig`) and want a full application framework, not just a container.
- You need declarative module composition (`fx.Module`) across many teams contributing to the same binary.
- Reflection-based wiring and its associated dependencies are an acceptable tradeoff for your team.

## Choose `samber/do` if: {#choose-samberdo-if}

- You want compile-time type checking on service resolution without adopting a full application framework.
- You need built-in health checks and graceful shutdown without wiring your own on top of lifecycle hooks.
- Zero non-stdlib dependencies matters (binary size, supply-chain surface, vendoring simplicity).

Already using `uber/fx` (or its underlying `uber/dig`) and considering a switch? See the [migration guide from Uber Dig](../migrating/migrating-from-dig.md) — the same constructor-based patterns apply to `fx` providers.
