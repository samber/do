---
title: Transient loading
description: Transient loading is a factory catalog
sidebar_position: 3
---

# Transient loading - a.k.a Factory

Transient loading is similar to lazy loading, but a new instance is provided at every invocation. It can be considered as a factory registry.

## Provider

A transient service must be defined with a `provider` (kind of constructor). This provider will be called at every invocation.

```go
type Provider[T any] func(do.Injector) (T, error)
```

Example:

```go
type MyService struct {
    Hello string
}

func NewMyService(i do.Injector) (*MyService, error) {
    return &MyService{Hello: "world"}, nil
}
```

## Inject service into DI container

A service can be injected in many ways. Here is the full list of transient loading service injections.

```go
func ProvideTransient[T any](i do.Injector, provider do.Provider[T])
func ProvideNamedTransient[T any](i do.Injector, name string, provider do.Provider[T])
func OverrideTransient[T any](i do.Injector, provider Provider[T])
func OverrideNamedTransient[T any](i do.Injector, name string, provider do.Provider[T])
```

:::info

This is highly recommended to use the anonymous service invocation to let the framework handle naming.

:::

```go
i := do.New()

do.ProvideTransient(i, NewMyService)
// or
do.ProvideNamedTransient(i, "my.really.cool.service", NewMyService)
```

## Error handling

On invocation, panics are cought by the framework and returned as an error.

## Hot service replacement

By default, providing a service twice will panic. Service can be replaced at runtime using `do.Override` helper.

For test purposes, it can be very useful to override a service given to the DI container.

:::warning

We highly discourage using this helper in production. Please use service aliasing instead.

:::

```go
type CalculatorTestSuite struct {
    suite.Suite
    i do.Injector
}

func (suite *CalculatorTestSuite) SetupSuite() {
    suite.i = do.New()
    do.ProvideTransient(suite.i, NewMyService)
}

func (suite *CalculatorTestSuite) TearDownSuite() {
    suite.i.Shutdown()
}

func (suite *CalculatorTestSuite) SetupTest() {
    do.OverrideTransient(suite.i, NewMyServiceMock)  // <- replace service by mock
}

func (suite *CalculatorTestSuite) Test1() {
  // ...
}

func (suite *CalculatorTestSuite) Test2() {
  // ...
}
```
