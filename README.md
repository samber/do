
# do - Dependency Injection

[![tag](https://img.shields.io/github/tag/samber/di.svg)](https://github.com/cryptoniumX/di/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-%23007d9c)
[![GoDoc](https://godoc.org/github.com/cryptoniumX/di?status.svg)](https://pkg.go.dev/github.com/cryptoniumX/di)
![Build Status](https://github.com/cryptoniumX/di/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/cryptoniumX/di)](https://goreportcard.com/report/github.com/cryptoniumX/di)
[![Coverage](https://img.shields.io/codecov/c/github/samber/do)](https://codecov.io/gh/samber/do)
[![License](https://img.shields.io/github/license/samber/do)](./LICENSE)


**⚙️ A dependency injection toolkit based on Go 1.18+ Generics.**

This library implements the Dependency Injection design pattern. It may replace the `uber/dig` fantastic package in simple Go projects. `samber/do` uses Go 1.18+ generics and therefore is typesafe.

**See also:**

- [samber/lo](https://github.com/samber/lo): A Lodash-style Go library based on Go 1.18+ Generics
- [samber/mo](https://github.com/samber/mo): Monads based on Go 1.18+ Generics (Option, Result, Either...)

**Why this name?**

I love **short name** for such a utility library. This name is the sum of `DI` and `Go` and no Go package currently uses this name.

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
- Lightweight, no dependencies
- No code generation

🚀 Services are loaded in invocation order.

🕵️ Service health can be checked individually or globally. Services implementing `di.Healthcheckable` interface will be called via `di.HealthCheck[type]()` or `injector.HealthCheck()`.

🛑 Services can be shutdowned properly, in back-initialization order. Services implementing `di.Shutdownable` interface will be called via `di.Shutdown[type]()` or `injector.Shutdown()`.

## 🚀 Install

```sh
go get github.com/cryptoniumX/di@v1
```

This library is v1 and follows SemVer strictly.

No breaking changes will be made to exported APIs before v2.0.0.

This library has no dependencies except the Go std lib.

## 💡 Quick start

You can import `do` using:

```go
import (
    "github.com/cryptoniumX/di"
)
```

Then instanciate services:

```go
func main() {
    injector := di.New()

    // provides CarService
    di.Provide(injector, NewCarService)

    // provides EngineService
    di.Provide(injector, NewEngineService)

    car := di.MustInvoke[*CarService](injector)
    car.Start()
    // prints "car starting"

    di.HealthCheck[EngineService](injector)
    // returns "engine broken"

    // injector.ShutdownOnSIGTERM()    // will block until receiving sigterm signal
    injector.Shutdown()
    // prints "car stopped"
}
```

Services:

```go
type EngineService interface{}

func NewEngineService(i *di.Injector) (EngineService, error) {
    return &engineServiceImplem{}, nil
}

type engineServiceImplem struct {}

// [Optional] Implements di.Healthcheckable.
func (c *engineServiceImplem) HealthCheck() error {
	return fmt.Errorf("engine broken")
}
```

```go
func NewCarService(i *di.Injector) (*CarService, error) {
    engine := di.MustInvoke[EngineService](i)
    car := CarService{Engine: engine}
    return &car, nil
}

type CarService struct {
	Engine EngineService
}

func (c *CarService) Start() {
	println("car starting")
}

// [Optional] Implements di.Shutdownable.
func (c *CarService) Shutdown() error {
	println("car stopped")
	return nil
}
```

## 🤠 Spec

[GoDoc: https://godoc.org/github.com/cryptoniumX/di](https://godoc.org/github.com/cryptoniumX/di)

Injector:

- [di.New](https://pkg.go.dev/github.com/cryptoniumX/di#New)
- [di.NewWithOpts](https://pkg.go.dev/github.com/cryptoniumX/di#NewWithOpts)
  - [injector.Clone](https://pkg.go.dev/github.com/cryptoniumX/di#injector.Clone)
  - [injector.CloneWithOpts](https://pkg.go.dev/github.com/cryptoniumX/di#injector.CloneWithOpts)
  - [injector.HealthCheck](https://pkg.go.dev/github.com/cryptoniumX/di#injector.HealthCheck)
  - [injector.Shutdown](https://pkg.go.dev/github.com/cryptoniumX/di#injector.Shutdown)
  - [injector.ShutdownOnSIGTERM](https://pkg.go.dev/github.com/cryptoniumX/di#injector.ShutdownOnSIGTERM)
  - [injector.ShutdownOnSignals](https://pkg.go.dev/github.com/cryptoniumX/di#injector.ShutdownOnSignals)
  - [injector.ListProvidedServices](https://pkg.go.dev/github.com/cryptoniumX/di#injector.ListProvidedServices)
  - [injector.ListInvokedServices](https://pkg.go.dev/github.com/cryptoniumX/di#injector.ListInvokedServices)
- [di.HealthCheck](https://pkg.go.dev/github.com/cryptoniumX/di#HealthCheck)
- [di.HealthCheckNamed](https://pkg.go.dev/github.com/cryptoniumX/di#HealthCheckNamed)
- [di.Shutdown](https://pkg.go.dev/github.com/cryptoniumX/di#Shutdown)
- [di.ShutdownNamed](https://pkg.go.dev/github.com/cryptoniumX/di#ShutdownNamed)
- [di.MustShutdown](https://pkg.go.dev/github.com/cryptoniumX/di#MustShutdown)
- [di.MustShutdownNamed](https://pkg.go.dev/github.com/cryptoniumX/di#MustShutdownNamed)

Service registration:

- [di.Provide](https://pkg.go.dev/github.com/cryptoniumX/di#Provide)
- [di.ProvideNamed](https://pkg.go.dev/github.com/cryptoniumX/di#ProvideNamed)
- [di.ProvideNamedValue](https://pkg.go.dev/github.com/cryptoniumX/di#ProvideNamedValue)
- [di.ProvideValue](https://pkg.go.dev/github.com/cryptoniumX/di#ProvideValue)

Service invocation:

- [di.Invoke](https://pkg.go.dev/github.com/cryptoniumX/di#Invoke)
- [di.MustInvoke](https://pkg.go.dev/github.com/cryptoniumX/di#MustInvoke)
- [di.InvokeNamed](https://pkg.go.dev/github.com/cryptoniumX/di#InvokeNamed)
- [di.MustInvokeNamed](https://pkg.go.dev/github.com/cryptoniumX/di#MustInvokeNamed)

Service override:

- [di.Override](https://pkg.go.dev/github.com/cryptoniumX/di#Override)
- [di.OverrideNamed](https://pkg.go.dev/github.com/cryptoniumX/di#OverrideNamed)
- [di.OverrideNamedValue](https://pkg.go.dev/github.com/cryptoniumX/di#OverrideNamedValue)
- [di.OverrideValue](https://pkg.go.dev/github.com/cryptoniumX/di#OverrideValue)

### Injector (DI container)

Build a container for your components. `Injector` is responsible for building services in the right order, and managing service lifecycle.

```go
injector := di.New()
```

Or use `nil` as the default injector:

```go
di.Provide(nil, func (i *Injector) (int, error) {
    return 42, nil
})

service := di.MustInvoke[int](nil)
```

You can check health of services implementing `func HealthCheck() error`.

```go
type DBService struct {
    db *sql.DB
}

func (s *DBService) HealthCheck() error {
    return s.db.Ping()
}

injector := di.New()
di.Provide(injector, ...)
di.Invoke(injector, ...)

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

injector := di.New()
di.Provide(injector, ...)
di.Invoke(injector, ...)

// shutdown all services in reverse order
injector.Shutdown()
```

List services:

```go
type DBService struct {
    db *sql.DB
}

injector := di.New()

di.Provide(injector, ...)
println(di.ListProvidedServices())
// output: []string{"*DBService"}

di.Invoke(injector, ...)
println(di.ListInvokedServices())
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

di.Provide[DBService](injector, func(i *Injector) (*DBService, error) {
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

di.ProvideNamed(injector, "dbconn", func(i *Injector) (*DBService, error) {
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

di.ProvideValue[Config](injector, Config{uri: "postgres://user:pass@host:5432/db"})
```

Named service, loaded eagerly:

```go
type Config struct {
    uri string
}

di.ProvideNamedValue(injector, "configuration", Config{uri: "postgres://user:pass@host:5432/db"})
```

### Service invocation

Loads anonymous service:

```go
type DBService struct {
    db *sql.DB
}

dbService, err := di.Invoke[DBService](injector)
```

Loads anonymous service or panics if service was not registered:

```go
type DBService struct {
    db *sql.DB
}

dbService := di.MustInvoke[DBService](injector)
```

Loads named service:

```go
config, err := di.InvokeNamed[Config](injector, "configuration")
```

Loads named service or panics if service was not registered:

```go
config := di.MustInvokeNamed[Config](injector, "configuration")
```

### Individual service healthcheck

Check health of anonymous service:

```go
type DBService struct {
    db *sql.DB
}

dbService, err := di.Invoke[DBService](injector)
err = di.HealthCheck[DBService](injector)
```

Check health of named service:

```go
config, err := di.InvokeNamed[Config](injector, "configuration")
err = di.HealthCheckNamed(injector, "configuration")
```

### Individual service shutdown

Unloads anonymous service:

```go
type DBService struct {
    db *sql.DB
}

dbService, err := di.Invoke[DBService](injector)
err = di.Shutdown[DBService](injector)
```

Unloads anonymous service or panics if service was not registered:

```go
type DBService struct {
    db *sql.DB
}

dbService := di.MustInvoke[DBService](injector)
di.MustShutdown[DBService](injector)
```

Unloads named service:

```go
config, err := di.InvokeNamed[Config](injector, "configuration")
err = di.ShutdownNamed(injector, "configuration")
```

Unloads named service or panics if service was not registered:

```go
config := di.MustInvokeNamed[Config](injector, "configuration")
di.MustShutdownNamed(injector, "configuration")
```

### Service override

By default, providing a service twice will panic. Service can be replaced at runtime using `di.Override` helper.

```go
di.Provide[Vehicle](injector, func (i *di.Injector) (Vehicle, error) {
    return &CarImplem{}, nil
})

di.Override[Vehicle](injector, func (i *di.Injector) (Vehicle, error) {
    return &BusImplem{}, nil
})
```

### Hooks

2 lifecycle hooks are available in Injectors:
- After registration
- After shutdown

```go
injector := di.NewWithOpts(&di.InjectorOpts{
    HookAfterRegistration: func(injector *di.Injector, serviceName string) {
        fmt.Printf("Service registered: %s\n", serviceName)
    },
    HookAfterShutdown: func(injector *di.Injector, serviceName string) {
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
var injector *di.Injector;

func init() {
    di.Provide[Service](injector, func (i *di.Injector) (Service, error) {
        return &RealService{}, nil
    })
    di.Provide[*App](injector, func (i *di.Injector) (*App, error) {
        return &App{i.MustInvoke[Service](i)}, nil
    })
}

func TestService(t *testing.T) {
    i := injector.Clone()
    defer i.Shutdown()

    // replace Service to MockService
    di.Override[Service](i, func (i *di.Injector) (Service, error) {
        return &MockService{}, nil
    }))

    app := di.Invoke[*App](i)
    // do unit testing with mocked service
}
```

## 🛩 Benchmark

// @TODO

## 🤝 Contributing

- Ping me on twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/cryptoniumX/di)
- Fix [open issues](https://github.com/cryptoniumX/di/issues) or request new features

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

[![GitHub Sponsors](https://img.shields.io/github/sponsors/samber?style=for-the-badge)](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2022 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
