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

## Testing and mocking

A package can ship a second variant, exposing test doubles behind the same interfaces as the production services. Swapping `Package` for its mock counterpart in a test injector replaces every service it registers, without touching the code under test.

This requires each service to be exposed through an interface (see [Accept interfaces, return structs](../service-invocation/accept-interfaces-return-structs.md)), with both the real and the mock implementation satisfying it.

<Tabs>
  <TabItem value="repository" label="pkg/repositories/user_repository.go" default>
    ```go
    package repositories

    type UserRepository interface {
        FindByID(id string) (*User, error)
    }

    func newUserRepository(i do.Injector) (UserRepository, error) {
        pool := do.MustInvoke[*pgxpool.Pool](i)
        return &userRepository{pool: pool}, nil
    }

    type userRepository struct {
        pool *pgxpool.Pool
    }

    func (r *userRepository) FindByID(id string) (*User, error) {
        // query the database
    }

    var _ UserRepository = (*userRepository)(nil)
    ```
  </TabItem>
  <TabItem value="mock" label="pkg/repositories/user_repository_mock.go">
    ```go
    package repositories

    func newUserRepositoryMock(i do.Injector) (UserRepository, error) {
        return &userRepositoryMock{users: map[string]*User{}}, nil
    }

    type userRepositoryMock struct {
        users map[string]*User
    }

    func (r *userRepositoryMock) FindByID(id string) (*User, error) {
        // return a fixture, or an error if absent
    }

    var _ UserRepository = (*userRepositoryMock)(nil)
    ```
  </TabItem>
  <TabItem value="package" label="pkg/repositories/package.go">
    ```go
    package repositories

    var Package = do.Package(
        do.Lazy(newPostgreSQLPool),
        do.Lazy(newUserRepository),
        do.Lazy(newBillingRepository),
    )

    var PackageMock = do.Package(
        do.Lazy(newUserRepositoryMock),
        do.Lazy(newBillingRepositoryMock),
    )
    ```
  </TabItem>
  <TabItem value="main" label="cmd/main.go">
    ```go
    package main

    import (
        "github.com/foo/bar/pkg/auth"
        "github.com/foo/bar/pkg/repositories"
    )

    func main() {
        injector := do.New(
            auth.Package,
            repositories.Package,
        )
    }
    ```
  </TabItem>
  <TabItem value="test" label="pkg/auth/auth_test.go">
    ```go
    package auth

    import (
        "testing"

        "github.com/foo/bar/pkg/repositories"
    )

    func TestAuthRegister(t *testing.T) {
        injector := do.New(
            Package,
            repositories.PackageMock,
        )

        // ...
    }
    ```
  </TabItem>
</Tabs>

**Play: https://go.dev/play/p/s-ZWLUGiMaT**

:::tip

Only mock the packages relevant to the test. Combining `repositories.Package` for services you don't need to fake with `repositories.PackageMock` for the one you do is a common pattern, as long as both variants don't register the same service twice in the same injector.

:::
