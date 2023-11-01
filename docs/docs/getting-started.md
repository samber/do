---
sidebar_position: 1
---

# Getting started

Let's discover **samber/do in less than 5 minutes**.

## What you'll need

[Go](https://go.dev/doc/install/) 18 or more.

Import package:

```sh
go get -u github.com/samber/do/v2
```

## Create a DI container

The simplest way to start is to use the default parameters:

```go
import "github.com/samber/do/v2"

injector := do.New()
```

With options:

```go
import "github.com/samber/do/v2"

injector := do.NewWithOps(&do.InjectorOpts{
    HookAfterRegistration func(scope *do.Scope, serviceName string) {
        // ...
    },
    HookAfterShutdown     func(scope *do.Scope, serviceName string) {
        // ...
    },

    Logf func(format string, args ...any) {
        // ...
    },
})
```

## Service provider and invocation

Services can be declared as a singleton or a factory. In this example, we will create 2 services `Car` and `Engine`, with a simple dependency relation.

Engine:

```go
type Engine struct {
    Started bool
}

// Provider
func NewEngine(i do.Injector) (*Engine, error) {
    return &Engine{
        Started: false,
    }, nil
}
```

Car:

```go
type Car struct {
    Engine *Engine
}

func (c *Car) Start() {
    c.Engine.Started = true
    println("vroooom")
}

// Provider
func NewCar(i do.Injector) (*Car, error) {
    return &Car{
        // import dependency
        Engine: do.MustInvoke[*Engine](i),
    }, nil
}
```

Now, let's start the application:

```go
func main() {
    // create DI container
    i := do.New()

    // inject both services into DI container
    Provide[*Car](i, NewCar)
    Provide[*Engine](i, NewEngine)

    // invoking car will instanciate Car services and its Engine dependency
    car, err := Invoke[*Car](i)
    if err != nil {
        log.Fatal(err.Error())
    }

    car.Start()  // that's all folk 🤗
}
```
