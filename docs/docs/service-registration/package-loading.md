---
title: Package loading
description: Package loading groups multiple service registrations.
sidebar_position: 4
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Package loading

Package loading groups multiple service registrations.

## Registration

The services can be assembled into a package, and then, exported all at once.


<Tabs>
  <TabItem value="stores" label="pkg/stores/package.go" default>
    ```go
    package stores

    var Package = do.Package(
        do.Lazy(NewPostgreSQLConnectionService),
        do.Lazy(NewUserRepository),
        do.Lazy(NewArticleRepository),
    )
    ```
  </TabItem>
  <TabItem value="observability" label="pkg/observability/package.go">
    ```go
    package observability

    var Package = do.Package(
        do.Eager(slog.New(slog.NewTextHandler(os.Stdout, nil))),
        do.EagerNamed("prometheus.collector", DefaultMetricCollector),
    )
    ```
  </TabItem>
  <TabItem value="main" label="cmd/main.go">
    ```go
    package main

    import (
        "github.com/foo/bar/pkg/stores"
        "github.com/foo/bar/pkg/observability"
        "github.com/foo/bar/pkg/handlers"
    )

    func main() {
        injector := do.New()
        stores.Package(injector)
        observability.Package(injector)

        // could be replaced by:
        // injector := do.New(
        //     stores.Package,
        //     observability.Package,
        // )

        // optional scope:
        scope := injector.Scope("handlers", handlers.Package)
    }
    ```
  </TabItem>
</Tabs>

**Play: https://go.dev/play/p/kmf8aOVyj96**

The traditional vocab can be translated for package registration:

- `Provide[T](Injector, Provider[T])` -> `Lazy(Provider[T])`
- `ProvideNamed[T](Injector, string, Provider[T])` -> `LazyNamed(string, Provider[T])`
- `ProvideValue(Injector, T)` -> `Eager(T)`
- `ProvideNamedValue[T](Injector, string, T)` -> `EagerNamed(string, T)`
- `ProvideTransient[T](Injector, Provider[T])` -> `Transient(Provider[T])`
- `ProvideNamedTransient[T](Injector, string, Provider[T])` -> `TransientNamed(string, Provider[T])`
- `As[Initial, Alias](Injector)` -> `Bind[Initial, Alias]()`
- `AsNamed[Initial, Alias](Injector, string, string)` -> `BindNamed[Initial, Alias](string, string)`
