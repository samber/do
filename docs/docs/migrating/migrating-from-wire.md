---
title: From Google Wire
description: Migrate your dependency injection from Google Wire to samber/do
sidebar_position: 2
---

# Migrating from Google Wire

This guide will help you migrate your dependency injection setup from Google Wire to `samber/do`.

## Overview {#overview}

Google Wire and `samber/do` are both dependency injection libraries for Go, but they have different approaches:

- **Google Wire**: Uses code generation with `//go:build wire` directives
- **samber/do**: Uses runtime dependency injection with a fluent API

## Key Differences {#key-differences}

| Feature               | Google Wire             | samber/do                 |
| --------------------- | ----------------------- | ------------------------- |
| **Approach**          | Code generation         | Runtime injection         |
| **Build time**        | Requires `wire` command | No build tools needed     |
| **Flexibility**       | Static, compile-time    | Dynamic, runtime          |
| **Error handling**    | Compile-time errors     | Runtime errors            |
| **Service lifecycle** | Limited                 | Full lifecycle management |

## Migration Steps {#migration-steps}

### 1. Remove Wire Dependencies {#1-remove-wire-dependencies}

Remove Wire from your `go.mod`:

```bash
go mod edit -droprequire github.com/google/wire
```

Remove any Wire-related build tags and imports from your code.

### 2. Replace Wire Provider Functions {#2-replace-wire-provider-functions}

**Before (Wire):**

```go
// +build wireinject

package main

import (
    "github.com/google/wire"
    "myapp/services"
)

func InitializeApp() (*App, error) {
    wire.Build(
        services.NewDatabase,
        services.NewUserService,
        services.NewApp,
    )
    return &App{}, nil
}
```

**After (samber/do):**

```go
package main

import (
    "github.com/samber/do/v2"
    "myapp/services"
)

func InitializeApp() (*App, error) {
    injector := do.New()

    // Register services
    do.Provide(injector, services.NewDatabase)
    do.Provide(injector, services.NewUserService)
    do.Provide(injector, services.NewApp)

    // Invoke the app
    app, err := do.Invoke[*App](injector)
    if err != nil {
        return nil, err
    }

    return app, nil
}
```

### 3. Update Service Constructors {#3-update-service-constructors}

Wire provider functions need to be updated for `samber/do`. Service constructors receive `do.Injector` as the first parameter and return an additional error:

**Before (Wire):**

```go
func NewUserService(db *Database) *UserService {
    return &UserService{db: db}
}

func NewDatabase(config *Config) *Database {
    return &Database{config: config}
}
```

**After (samber/do):**

```go
func NewUserService(i do.Injector) (*UserService, error) {
    db := do.MustInvoke[*Database](i)
    return &UserService{db: db}, nil
}

func NewDatabase(i do.Injector) (*Database, error) {
    config := do.MustInvoke[*Config](i)
    return &Database{config: config}, nil
}
```

### 4. Handle Interface Bindings {#4-handle-interface-bindings}

**Before (Wire):**

```go
var Set = wire.NewSet(
    wire.Bind(new(Repository), new(*UserRepository)),
    NewUserRepository,
    NewUserService,
)
```

**After (samber/do):**

```go
injector := do.New()

do.Provide(injector, func(i do.Injector) (Repository, error) {
    return &UserRepository{}, nil
})
```

Or declare an explicit binding:

```go
// Register the concrete implementation
do.Provide(injector, func(i do.Injector) (*UserRepository, error) {
    return &UserRepository{}, nil
})

// Register the interface binding
do.As[*UserRepository, Repository](i)
```

Or bind on service loading:

```go
// Register the concrete implementation
do.Provide(injector, func(i do.Injector) (*UserRepository, error) {
    repo := do.MustInvoke[*UserRepository](i)
    return repo, nil
})

// Find the matching concrete type during invocation
userRepository, err := do.InvokeAs[Repository](i)
```

### 5. Update Main Function {#5-update-main-function}

**Before (Wire):**

```go
func main() {
    app, err := InitializeApp()
    if err != nil {
        log.Fatal(err)
    }
    app.Run()
}
```

**After (samber/do):**

```go
func main() {
    app, err := InitializeApp()
    if err != nil {
        log.Fatal(err)
    }

    // Optional: Graceful shutdown
    defer app.Shutdown()

    app.Run()
}
```

## Advanced Migration Patterns {#advanced-migration-patterns}

### Provider Sets {#provider-sets}

**Before (Wire):**

```go
var UserSet = wire.NewSet(
    NewUserRepository,
    NewUserService,
)

var AppSet = wire.NewSet(
    UserSet,
    NewDatabase,
    NewApp,
)
```

**After (samber/do):**

```go
var UserPackage = do.Package(
    do.Lazy(NewUserRepository),
    do.Lazy(NewUserService),
)

var AppPackage = do.Package(
    do.Lazy(UserPackage),
    do.Lazy(NewDatabase),
    do.Lazy(NewApp),
)
```

### Conditional Providers {#conditional-providers}

**Before (Wire):**

```go
func InitializeApp(env string) (*App, error) {
    if env == "test" {
        wire.Build(TestSet)
    } else {
        wire.Build(ProdSet)
    }
    return &App{}, nil
}
```

**After (samber/do):**

```go
func InitializeApp(env string) (*App, error) {
    injector := do.New()

    if env == "test" {
        RegisterTestServices(injector)
    } else {
        RegisterProdServices(injector)
    }

    app := do.MustInvoke[*App](injector)
    return app, nil
}
```

## Benefits of Migration {#benefits-of-migration}

1. **No build tools required** - No need to run `wire` command
2. **Runtime flexibility** - Can change dependencies at runtime
3. **Better error handling** - More detailed error messages
4. **Service lifecycle management** - Built-in shutdown and health checks
5. **Scoped injection** - Support for request-scoped services
6. **Lazy loading** - Services are created only when needed

## Common Pitfalls {#common-pitfalls}

1. **Missing dependencies** - Ensure all required services are registered
2. **Circular dependencies** - `samber/do` will detect and report these
3. **Interface bindings** - Use `do.As` for interface implementations
4. **Panics vs errors** - Prefer `do.MustInvoke` over `do.Invoke` when a missing service should stop the program immediately instead of being handled as a recoverable error

## Testing {#testing}

Your existing tests should work with minimal changes:

```go
func TestUserService(t *testing.T) {
    injector := testInjector.Clone()

    // NewMockUserService is a provider returning *UserService wired with test doubles
    do.Override(injector, NewMockUserService)

    service := do.MustInvoke[*UserService](injector)

    // Your test logic here
}
```

## Next Steps {#next-steps}

After migration, consider exploring these `samber/do` features:

- [Interface binding](/docs/service-invocation/accept-interfaces-return-structs.md)
- [Scoped Services](/docs/container/scope.md)
- [Health Checks](/docs/service-lifecycle/healthchecker.md)
- [Graceful Shutdown](/docs/service-lifecycle/shutdowner.md)
- [Dependency resolution troubleshooting](/docs/troubleshooting/scope-tree.md)
