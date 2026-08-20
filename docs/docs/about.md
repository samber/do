---
title: ✌️ Dependency Injection in Go
description: samber/do is a type-safe dependency injection toolkit for Go, built on 1.18+ generics instead of reflection or code generation. Compare it to uber/dig, uber/fx, and google/wire.
sidebar_position: 0
---

import Head from '@docusaurus/Head';

<Head>
  <script type="application/ld+json">
    {JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'TechArticle',
      headline: 'Dependency Injection in Go',
      description: 'samber/do is a type-safe dependency injection toolkit for Go, built on 1.18+ generics instead of reflection or code generation.',
      about: {'@id': 'https://do.samber.dev/#software'},
      author: {'@id': 'https://do.samber.dev/#person-samber'},
      mainEntityOfPage: 'https://do.samber.dev/docs/about',
    })}
  </script>
</Head>

# ✌️ Dependency Injection in Go

⚙️ Dependency injection (DI) in Go is the practice of supplying a component with the services it depends on, instead of letting it construct them itself — making code easier to test, swap, and reason about. `samber/do` is a DI toolkit for Go that implements this pattern using **1.18+ generics** instead of reflection, giving you a type-safe API with **no code generation** and **zero external dependencies**.

<img
  src="/img/cover.webp"
  srcSet="/img/cover-640.webp 640w, /img/cover.webp 1200w"
  sizes="(max-width: 640px) 100vw, 1200px"
  width={1200}
  height={404}
  decoding="async"
  alt="samber/do — type-safe dependency injection for Go using generics"
/>

**See also:**

- [samber/lo](https://github.com/samber/lo): A Lodash-style Go library based on Go 1.18+ Generics
- [samber/mo](https://github.com/samber/mo): Monads based on Go 1.18+ Generics (Option, Result, Either...)

**Why this name?**

I love the **short name** for such a utility library. This name is the sum of `DI` and `Go` and no Go package uses this name.

## samber/do vs uber/dig vs uber/fx vs google/wire {#samberdo-vs-uberdig-vs-uberfx-vs-googlewire}

|                           | `samber/do`                             | `uber/dig`    | `uber/fx`                 | `google/wire`                 | Manual DI    |
| ------------------------- | --------------------------------------- | ------------- | ------------------------- | ----------------------------- | ------------ |
| **Mechanism**             | Generics                                | Reflection    | Reflection (built on dig) | Code generation               | Hand-written |
| **Type safety**           | Compile-time                            | Runtime       | Runtime                   | Compile-time                  | Compile-time |
| **Code generation**       | None                                    | None          | None                      | Required (`wire` CLI)         | None         |
| **Lifecycle management**  | Health checks, graceful shutdown, hooks | None built-in | Start/stop hooks          | None                          | Hand-written |
| **Scopes / modules**      | Full scope tree                         | Limited       | Modules                   | Provider sets                 | Hand-written |
| **Debugging tools**       | Scope tree, dependency graph, Web UI    | None built-in | `fx.WithLogger`           | None (inspect generated code) | N/A          |
| **External dependencies** | None                                    | None          | Several (dig, zap, ...)   | `wire` build tool             | None         |

**When to use `samber/do`**: you want compile-time-checked service resolution without a build step, plus built-in health checks and graceful shutdown for production services. See the detailed comparisons: [do vs uber/fx](./comparison/samber-do-vs-uber-fx.md), [do vs google/wire](./comparison/samber-do-vs-google-wire.md), and [do vs uber/dig](./comparison/samber-do-vs-uber-dig.md). Already using one of these? Check the [migration guides](./migrating/migrating-from-dig.md).

## 💡 Features {#-features}

`samber/do` is built with high attention to the developer and contributor experience.

- **📒 [Service registration](./service-registration/lazy-loading.md)**
  - Register by type
  - Register by name
  - Register multiple services from a package at once
- **🪃 [Service invocation](./service-invocation/service-invocation.md)**
  - [Eager loading](./service-registration/eager-loading.md)
  - [Lazy loading](./service-registration/lazy-loading.md)
  - [Transient loading](./service-registration/transient-loading.md)
  - Tag-based invocation
  - Circular dependency detection
- **🧙‍♂️ [Service aliasing](./service-invocation/accept-interfaces-return-structs.md)**
  - Implicit (provide struct, invoke interface)
  - Explicit (provide struct, bind interface, invoke interface)
- **🔁 Service lifecycle**
  - [Health check](./service-lifecycle/healthchecker.md)
  - [Graceful unload (shutdown)](./service-lifecycle/shutdowner.md)
  - Dependency-aware parallel shutdown
  - Lifecycle hooks
- **📦 [Scope (a.k.a. module) tree](./container/scope.md)**
  - Visibility control
  - Dependency grouping
- **📤 [Container](./container/options.md)**
  - Dependency graph resolution and visualization
  - Default container
  - [Container cloning](./container/clone.md)
  - Service override
- **🧪 Debugging & introspection**
  - Explain APIs: [scope tree](./troubleshooting/scope-tree.md) and [service dependencies](./troubleshooting/service-dependencies.md)
  - [Web UI](./troubleshooting/web-ui.md) & HTTP middleware (std, Gin, Fiber, Echo, Chi)
- **🌈 Lightweight, no dependencies**
- **🔅 No code generation**
- **😷 Type‑safe API**

## Where to go next {#where-to-go-next}

- New to `samber/do`? Start with [Getting started](./getting-started.md).
- Coming from another DI library? Read the [migration guides](./migrating/migrating-from-dig.md) or the [comparison pages](./comparison/samber-do-vs-uber-fx.md).
- Have a specific question? Check the [FAQ](./faq.md) or the [glossary](./glossary.md).
- Upgrading from v1? See the [v1 to v2 upgrade guide](./upgrading/from-v1-x-to-v2.md).
