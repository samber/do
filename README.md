
# do - Dependency Injection

[![tag](https://img.shields.io/github/tag/samber/do.svg)](https://github.com/samber/do/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/do?status.svg)](https://pkg.go.dev/github.com/samber/do)
![Build Status](https://github.com/samber/do/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/do)](https://goreportcard.com/report/github.com/samber/do)
[![Coverage](https://img.shields.io/codecov/c/github/samber/do)](https://codecov.io/gh/samber/do)
[![License](https://img.shields.io/github/license/samber/do)](./LICENSE)


**⚙️ A dependency injection toolkit based on Go 1.18+ Generics.**

This library implements the Dependency Injection design pattern. It may replace the `uber/dig` fantastic package in simple Go projects. `samber/do` uses Go 1.18+ generics instead of reflection and therefore is typesafe.

**See also:**

- [samber/lo](https://github.com/samber/lo): A Lodash-style Go library based on Go 1.18+ Generics
- [samber/mo](https://github.com/samber/mo): Monads based on Go 1.18+ Generics (Option, Result, Either...)

**Why this name?**

I love **short name** for such utility library. This name is the sum of `DI` and `Go` and no Go package currently uses this name.

## 💡 Features

- Service registration
- Service invocation
- Service health check
- Service shutdown
- Service lifecycle hooks
- Named or anonymous services
- Eagerly or lazily loaded services
- Dependency graph resolution
- Default injector
- Injector cloning
- Service override

🚀 Services are loaded in invocation order.

🕵️ Service health can be checked individually or globally. Services implementing `do.Healthcheckable` interface will be called via `do.HealthCheck[type]()` or `injector.HealthCheck()`.

🛑 Services can be shutdowned properly, in back-initialization order. Services implementing `do.Shutdownable` interface will be called via `do.Shutdown[type]()` or `injector.Shutdown()`.

## 🚀 Install

```sh
go get github.com/samber/do@v1
```

This library is v1 and follows SemVer strictly.

No breaking changes will be made to exported APIs before v2.0.0.

This library has no dependencies outside the Go standard library.

## 💡 Quick start

You can import `do` using:

```go
import (
    "github.com/samber/do"
)
```

Then instanciate services:

```go
func main() {
    injector := do.New()

    // provides CarService
    do.Provide(injector, NewCarService)

    // provides EngineService
    do.Provide(injector, NewEngineService)

    car := do.MustInvoke[*CarService](injector)
    car.Start()
    // prints "car starting"

    do.HealthCheck[EngineService](injector)
    // returns "engine broken"

    // injector.ShutdownOnSIGTERM()    // will block until receiving sigterm signal
    injector.Shutdown()
    // prints "car stopped"
}
```

Services:

```go
type EngineService interface{}

func NewEngineService(i *do.Injector) (EngineService, error) {
    return &engineServiceImplem{}, nil
}

type engineServiceImplem struct {}

// [Optional] Implements do.Healthcheckable.
func (c *engineServiceImplem) HealthCheck() error {
	return fmt.Errorf("engine broken")
}
```

```go
func NewCarService(i *do.Injector) (*CarService, error) {
    engine := do.MustInvoke[EngineService](i)
    car := CarService{Engine: engine}
    return &car, nil
}

type CarService struct {
	Engine EngineService
}

func (c *CarService) Start() {
	println("car starting")
}

// [Optional] Implements do.Shutdownable.
func (c *CarService) Shutdown() error {
	println("car stopped")
	return nil
}
```

## 🤠 Spec

[GoDoc: https://godoc.org/github.com/samber/do](https://godoc.org/github.com/samber/do)

Injector:

- [do.New](https://pkg.go.dev/github.com/samber/do#New)
- [do.NewWithOpts](https://pkg.go.dev/github.com/samber/do#NewWithOpts)
  - [injector.Clone](https://pkg.go.dev/github.com/samber/do#injector.Clone)
  - [injector.CloneWithOpts](https://pkg.go.dev/github.com/samber/do#injector.CloneWithOpts)
  - [injector.HealthCheck](https://pkg.go.dev/github.com/samber/do#injector.HealthCheck)
  - [injector.Shutdown](https://pkg.go.dev/github.com/samber/do#injector.Shutdown)
  - [injector.ShutdownOnSIGTERM](https://pkg.go.dev/github.com/samber/do#injector.ShutdownOnSIGTERM)
  - [injector.ListProvidedServices](https://pkg.go.dev/github.com/samber/do#injector.ListProvidedServices)
  - [injector.ListInvokedServices](https://pkg.go.dev/github.com/samber/do#injector.ListInvokedServices)
- [do.HealthCheck](https://pkg.go.dev/github.com/samber/do#HealthCheck)
- [do.HealthCheckNamed](https://pkg.go.dev/github.com/samber/do#HealthCheckNamed)
- [do.Shutdown](https://pkg.go.dev/github.com/samber/do#Shutdown)
- [do.ShutdownNamed](https://pkg.go.dev/github.com/samber/do#ShutdownNamed)
- [do.MustShutdown](https://pkg.go.dev/github.com/samber/do#MustShutdown)
- [do.MustShutdownNamed](https://pkg.go.dev/github.com/samber/do#MustShutdownNamed)

Service registration:

- [do.Provide](https://pkg.go.dev/github.com/samber/do#Provide)
- [do.ProvideNamed](https://pkg.go.dev/github.com/samber/do#ProvideNamed)
- [do.ProvideNamedValue](https://pkg.go.dev/github.com/samber/do#ProvideNamedValue)
- [do.ProvideValue](https://pkg.go.dev/github.com/samber/do#ProvideValue)

Service invocation:

- [do.Invoke](https://pkg.go.dev/github.com/samber/do#Invoke)
- [do.MustInvoke](https://pkg.go.dev/github.com/samber/do#MustInvoke)
- [do.InvokeNamed](https://pkg.go.dev/github.com/samber/do#InvokeNamed)
- [do.MustInvokeNamed](https://pkg.go.dev/github.com/samber/do#MustInvokeNamed)

Service override:

- [do.Override](https://pkg.go.dev/github.com/samber/do#Override)
- [do.OverrideNamed](https://pkg.go.dev/github.com/samber/do#OverrideNamed)
- [do.OverrideNamedValue](https://pkg.go.dev/github.com/samber/do#OverrideNamedValue)
- [do.OverrideValue](https://pkg.go.dev/github.com/samber/do#OverrideValue)

### Injector (DI container)

Build a container for your components. `Injector` is responsible for building services in the right order, and managing service lifecycle.

```go
injector := do.New()
```

Or use `nil` as the default injector:

```go
do.Provide(nil, func (i *Injector) (int, error) {
    return 42, nil
})

service := do.MustInvoke[int](nil)
```

You can check health of services implementing `func HealthCheck() error`.

```go
type DBService struct {
    db *sql.DB
}

func (s *DBService) HealthCheck() error {
    return s.db.Ping()
}

injector := do.New()
do.Provide(injector, ...)
do.Invoke(injector, ...)

statuses := injector.HealthCheck()
// map[string]error{
//   "*DBService": nil,
// }
```

De-initialize all compoments properly. Services implementing `func Shutdown() error` will be called synchronously in back-initialization order.

```go
type DBService struct {
    db *sql.DB
}

func (s *DBService) Shutdown() error {
    return s.db.Close()
}

injector := do.New()
do.Provide(injector, ...)
do.Invoke(injector, ...)

// shutdown all services in reverse order
injector.Shutdown()
```

List services:

```go
type DBService struct {
    db *sql.DB
}

injector := do.New()

do.Provide(injector, ...)
println(do.ListProvidedServices())
// output: []string{"*DBService"}

do.Invoke(injector, ...)
println(do.ListInvokedServices())
// output: []string{"*DBService"}
```

### Service registration

Services can be registered in multiple way:

- with implicit name (struct or interface name)
- with explicit name
- eagerly
- lazily

Anonymous service, loaded lazily:

```go
type DBService struct {
    db *sql.DB
}

do.Provide[DBService](injector, func(i *Injector) (*DBService, error) {
    db, err := sql.Open(...)
    if err != nil {
        return nil, err
    }

    return &DBService{db: db}, nil
})
```

Named service, loaded lazily:

```go
type DBService struct {
    db *sql.DB
}

do.ProvideNamed(injector, "dbconn", func(i *Injector) (*DBService, error) {
    db, err := sql.Open(...)
    if err != nil {
        return nil, err
    }

    return &DBService{db: db}, nil
})
```

Anonymous service, loaded eagerly:

```go
type Config struct {
    uri string
}

do.ProvideValue[Config](injector, Config{uri: "postgres://user:pass@host:5432/db"})
```

Named service, loaded eagerly:

```go
type Config struct {
    uri string
}

do.ProvideNamedValue(injector, "configuration", Config{uri: "postgres://user:pass@host:5432/db"})
```

### Service invocation

Loads anonymous service:

```go
type DBService struct {
    db *sql.DB
}

dbService, err := do.Invoke[DBService](injector)
```

Loads anonymous service or panics if service was not registered:

```go
type DBService struct {
    db *sql.DB
}

dbService := do.MustInvoke[DBService](injector)
```

Loads named service:

```go
config, err := do.InvokeNamed[Config](injector, "configuration")
```

Loads named service or panics if service was not registered:

```go
config := do.MustInvokeNamed[Config](injector, "configuration")
```

### Individual service healthcheck

Check health of anonymous service:

```go
type DBService struct {
    db *sql.DB
}

dbService, err := do.Invoke[DBService](injector)
err = do.HealthCheck[DBService](injector)
```

Check health of named service:

```go
config, err := do.InvokeNamed[Config](injector, "configuration")
err = do.HealthCheckNamed(injector, "configuration")
```

### Individual service shutdown

Unloads anonymous service:

```go
type DBService struct {
    db *sql.DB
}

dbService, err := do.Invoke[DBService](injector)
err = do.Shutdown[DBService](injector)
```

Unloads anonymous service or panics if service was not registered:

```go
type DBService struct {
    db *sql.DB
}

dbService := do.MustInvoke[DBService](injector)
do.MustShutdown[DBService](injector)
```

Unloads named service:

```go
config, err := do.InvokeNamed[Config](injector, "configuration")
err = do.ShutdownNamed(injector, "configuration")
```

Unloads named service or panics if service was not registered:

```go
config := do.MustInvokeNamed[Config](injector, "configuration")
do.MustShutdownNamed(injector, "configuration")
```

### Service override

By default, providing a service twice will panic. Service can be replaced at runtime using `do.Override` helper.

```go
do.Provide[Vehicle](injector, func (i *do.Injector) (Vehicle, error) {
    return &CarImplem{}, nil
})

do.Override[Vehicle](injector, func (i *do.Injector) (Vehicle, error) {
    return &BusImplem{}, nil
})
```

### Hooks

2 lifecycle hooks are available in Injectors:
- After registration
- After shutdown

```go
injector := do.NewWithOpts(&do.InjectorOpts{
    HookAfterRegistration: func(injector *do.Injector, serviceName string) {
        fmt.Printf("Service registered: %s\n", serviceName)
    },
    HookAfterShutdown: func(injector *do.Injector, serviceName string) {
        fmt.Printf("Service stopped: %s\n", serviceName)
    },

    Logf: func(format string, args ...any) {
        log.Printf(format, args...)
    },
})
```

### Cloning injector

Cloned injector have same service registrations as it's parent, but it doesn't share invoked service state.

Clones are useful for unit testing by replacing some services to mocks.

```go
var injector *do.Injector;

func init() {
    do.Provide[Service](injector, func (i *do.Injector) (Service, error) {
        return &RealService{}, nil
    })
    do.Provide[*App](injector, func (i *do.Injector) (*App, error) {
        return &App{i.MustInvoke[Service](i)}, nil
    })
}

func TestService(t *testing.T) {
    i := injector.Clone()
    defer i.Shutdown()

    // replace Service to MockService
    do.Override[Service](i, func (i *do.Injector) (Service, error) {
        return &MockService{}, nil
    }))

    app := do.Invoke[*App](i)
    // do unit testing with mocked service
}
```

## 🛩 Benchmark

// @TODO

This library does not use `reflect` package. We don't expect overhead.

## 🤝 Contributing

- Ping me on twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/samber/do)
- Fix [open issues](https://github.com/samber/do/issues) or request new features

Don't hesitate ;)

### With Docker

```bash
docker-compose run --rm dev
```

### Without Docker

```bash
# Install some dev dependencies
make tools

# Run tests
make test
# or
make watch-test
```

## 👤 Contributors

![Contributors](https://contrib.rocks/image?repo=samber/do)

## 💫 Show your support

Give a ⭐️ if this project helped you!

[![support us](https://c5.patreon.com/external/logo/become_a_patron_button.png)](https://www.patreon.com/samber)

## 📝 License

Copyright © 2022 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
