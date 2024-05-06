---
title: 🚀 Getting started
description: Let's discover samber/do in less than 5 minutes.
sidebar_position: 1
---

# Getting started

Let's discover **samber/do in less than 5 minutes**.

## What you'll need

Compatible with [Go](https://go.dev/doc/install/) 1.18 or more.

This library has no dependencies except the Go std lib.

Import package:

```sh
go get -u github.com/samber/do/v2
```

## Create a DI container

The simplest way to start is to use the default options:

```go
import "github.com/samber/do/v2"

injector := do.New()
```

## Service registration and invocation

Services can be declared as a singleton or a factory. In this example, we will create 2 services `Car` and `Engine`, with a simple dependency relation.

```go
func main() {
    // create DI container
    i := do.New()

    // inject both services into DI container
    Provide[*Car](i, NewCar)
    Provide[*Engine](i, NewEngine)

    // invoking car will instantiate Car services and its Engine dependency
    car, err := Invoke[*Car](i)
    if err != nil {
        log.Fatal(err.Error())
    }

    car.Start()  // that's all folk 🤗

    // handle ctrl-c and shutdown services
    i.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
}
```

Engine:

```go
type Engine struct {
    Started bool
}

func (e *Engine) Shutdown() error {
    // called on injector shutdown
    e.Started = false
    return nil
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
