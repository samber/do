---
title: ✌️ About
description: Discover "do", the dependency injection toolkit for Go
sidebar_position: 0
---

# ✌️ About

**⚙️ A dependency injection toolkit based on Go 1.18+ Generics.**

This library implements the Dependency Injection design pattern. It may replace the `uber/dig` fantastic package. `samber/do` uses Go 1.18+ generics instead of reflection and therefore offers a typesafe API.

**See also:**

- [samber/lo](https://github.com/samber/lo): A Lodash-style Go library based on Go 1.18+ Generics
- [samber/mo](https://github.com/samber/mo): Monads based on Go 1.18+ Generics (Option, Result, Either...)

**Why this name?**

I love the **short name** for such a utility library. This name is the sum of `DI` and `Go` and no Go package uses this name.

## 💡 Features

`samber/do` is built with high attention to the developer and contributor experience.

- **📒 Service registration**
  - By type inference
  - By name
- **🪃 Service invocation**
  - Eager loading
  - Lazy loading
  - Transient loading
- **🧙‍♂️ Service aliasing**
  - Implicit (provide struct, invoke interface)
  - Explicit (provide struct, bind interface, invoke interface)
- **🔁 Service lifecycle**
  - Health check
  - Graceful unload (shutdown)
  - Lifecycle hooks
- **📦 Scope (a.k.a module) tree**
  - Visibility control
  - Dependency grouping
- **📤 Injector**
  - Dependency graph resolution and visualization
  - Default injector
  - Injector cloning
  - Service override
- **🌈 Lightweight, no dependencies**
- **🔅 No code generation**
- **😷 Typesafe API**
