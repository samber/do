---
title: From Uber Dig
description: Migrate your dependency injection from Uber Dig to samber/do
sidebar_position: 3
---

# Migrating from Uber Dig

This guide will help you migrate your dependency injection setup from Uber Dig to `samber/do`.

## Overview

Uber Dig and `samber/do` are both runtime dependency injection libraries for Go, but they have different APIs and features:

- **Uber Dig**: Uses a builder pattern with `dig.New()` and `Provide`/`Invoke` methods
- **samber/do**: Uses a fluent API with `do.New()` and similar `Provide`/`Invoke` methods

## Key Differences

| Feature               | Uber Dig                      | samber/do                      |
| --------------------- | ----------------------------- | ------------------------------ |
| **API Style**         | Builder pattern               | Fluent API                     |
| **Error Handling**    | Returns errors from `Build()` | Returns errors from `Invoke()` |
| **Service Lifecycle** | Basic                         | Full lifecycle management      |
| **Scoped Services**   | Limited                       | Full scoping support           |
| **Health Checks**     | Not built-in                  | Built-in support               |
| **Graceful Shutdown** | Manual                        | Built-in support               |

## Migration Steps

### 1. Remove Dig Dependencies

Remove Dig from your `go.mod`:

```bash
go mod edit -droprequire go.uber.org/dig
```

Remove Dig imports from your code.

### 2. Replace Container Creation

**Before (Dig):**
```go
import "go.uber.org/dig"

func main() {
    container := dig.New()
    // ... setup container
}
```

**After (samber/do):**
```go
import "github.com/samber/do/v2"

func main() {
    injector := do.New()
    // ... setup injector
}
```

### 3. Update Service Registration

**Before (Dig):**
```go
err := container.Provide(NewDatabase)
if err != nil {
    log.Fatal(err)
}

err = container.Provide(NewUserService)
if err != nil {
    log.Fatal(err)
}
```

**After (samber/do):**
```go
do.Provide(injector, NewDatabase)
do.Provide(injector, NewUserService)
```

### 4. Update Service Invocation

**Before (Dig):**
```go
err := container.Invoke(func(app *App) {
    app.Run()
})
if err != nil {
    log.Fatal(err)
}
```

**After (samber/do):**
```go
app, err := do.Invoke[*App](injector)
if err != nil {
    log.Fatal(err)
}
// or
app := do.MustInvoke[*App](injector)

app.Run()
```

### 5. Handle Constructor Functions

Dig and `samber/do` use different constructor function signatures. `samber/do` constructors receive `do.Injector` as the first parameter and return an additional error:

**Before (Dig):**
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
    // if service is not found, `do` will catch the panic and return an `error`
    db := do.MustInvoke[*Database](i)
    return &UserService{db: db}, nil
}

func NewDatabase(i do.Injector) (*Database, error) {
    config := do.MustInvoke[*Config](i)
    return &Database{config: config}, nil
}
```

### 6. Update Interface Bindings

**Before (Dig):**
```go
err := container.Provide(func() Repository {
    return &UserRepository{}
})
```

**After (samber/do):**
```go
do.Provide(injector, func(i do.Injector) (Repository, error) {
    return &UserRepository{}, nil
})
```

Or declare an explicit binding:

```go
// Register the concrete implementation
do.Provide(injector, func(i do.Injector) (*UserRepository, error) {
    repo := do.MustInvoke[*UserRepository](i)
    return repo, nil
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

## Advanced Migration Patterns

### Provider Groups

**Before (Dig):**
```go
type Result struct {
    dig.Out
    Service1 *Service1 `name:"service1"`
    Service2 *Service2 `name:"service2"`
}

func ProvideServices() Result {
    return Result{
        Service1: NewService1(),
        Service2: NewService2(),
    }
}

err := container.Provide(ProvideServices)
```

**After (samber/do):**
```go
type Result struct {
    Service1 *Service1 `do:""`
    Service2 *Service2 `do:""`
}

do.Provide(injector, NewService1)
do.Provide(injector, NewService2)
do.Provide(injector, do.InvokeStruct[*Result])
```

### Parameter Structs

**Before (Dig):**
```go
type Params struct {
    dig.In
    DB      *Database
    Logger  *Logger
}

func NewUserService(params Params) *UserService {
    return &UserService{
        db:     params.DB,
        logger: params.Logger,
    }
}
```

**After (samber/do):**
```go
func NewUserService(i do.Injector) (*UserService, error) {
    db := do.MustInvoke[*Database](i)
    logger := do.MustInvoke[*Logger](i)
    
    return &UserService{
        db:     db,
        logger: logger,
    }, nil
}
```

### Optional Dependencies

**Before (Dig):**
```go
type Params struct {
    dig.In
    DB      *Database
    Cache   *Cache `optional:"true"`
}

func NewUserService(params Params) *UserService {
    return &UserService{
        db:    params.DB,
        cache: params.Cache, // nil if not provided
    }
}
```

**After (samber/do):**
```go
func NewUserService(i do.Injector) (*UserService, error) {
    db := do.MustInvoke[*Database](i)
    cache, _ := do.Invoke[*Cache](i) // error is ignored
    return &UserService{
        db: db,
        cache: cache,
    }, nil
}

// Or use conditional registration
func RegisterServices(injector do.Injector, useCache bool) {
    do.Provide(injector, NewDatabase)
    if useCache {
        do.Provide(injector, NewCache)
    }
    do.Provide(injector, NewUserService)
}
```

### Value Groups

**Before (Dig):**
```go
type Result struct {
    dig.Out
    Handler Handler `group:"handlers"`
}

func ProvideHandler1() Result {
    return Result{Handler: NewHandler1()}
}

func ProvideHandler2() Result {
    return Result{Handler: NewHandler2()}
}

type Params struct {
    dig.In
    Handlers []Handler `group:"handlers"`
}
```

**After (samber/do):**
Should be available starting do v2.1

## Complete Example

**Before (Dig):**
```go
package main

import (
    "log"
    "go.uber.org/dig"
)

func main() {
    container := dig.New()
    
    err := container.Provide(NewDatabase)
    if err != nil {
        log.Fatal(err)
    }
    
    err = container.Provide(NewUserService)
    if err != nil {
        log.Fatal(err)
    }
    
    err = container.Provide(NewApp)
    if err != nil {
        log.Fatal(err)
    }
    
    err = container.Invoke(func(app *App) {
        app.Run()
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

**After (samber/do):**
```go
package main

import (
    "log"
    "github.com/samber/do/v2"
)

func main() {
    injector := do.New()
    
    // Register services
    do.Provide(injector, NewDatabase)
    do.Provide(injector, NewUserService)
    do.Provide(injector, NewApp)
    
    // Get the app and run it
    app, err := do.Invoke[*App](injector)
    if err != nil {
        log.Fatal(err)
    }
    
    // Optional: Graceful shutdown
    defer injector.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
    
    app.Run()
}
```

## Benefits of Migration

1. **Simpler API** - More intuitive fluent API
2. **Better error messages** - More detailed error information
3. **Service lifecycle management** - Built-in shutdown and health checks
4. **Scoped services** - Full support for request-scoped services
5. **No external dependencies** - Pure Go implementation
6. **Better testing support** - Easier to mock and test

## Common Pitfalls

1. **Error handling** - `samber/do` returns errors from `Invoke()`, not `Provide()`
2. **Circular dependencies** - `samber/do` will detect and report these
3. **Interface bindings** - Use `do.As` for interface implementations
4. **Error handling** - Use `do.MustInvoke` instead of `do.Invoke` as runtime errors and handled automatically

## Testing

Your existing tests should work with minimal changes:

```go
func TestUserService(t *testing.T) {
    injector := do.New()
    
    // Register test dependencies
    do.Provide(injector, NewMockDatabase)
    do.Provide(injector, NewUserService)
    
    service := do.MustInvoke[*UserService](injector)
    
    // Your test logic here
}
```

## Next Steps

After migration, consider exploring these `samber/do` features:

- [Interface binding](/docs/service-invocation/accept-interfaces-return-structs.md)
- [Scoped Services](/docs/container/scope.md)
- [Health Checks](/docs/service-lifecycle/healthchecker.md)
- [Graceful Shutdown](/docs/service-lifecycle/shutdowner.md)
- [Dependency resolution troubleshooting](/docs/troubleshooting/scope-tree.md)
