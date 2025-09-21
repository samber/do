---
title: Lazy loading
description: Lazy loading limits the number of components that load in an app.
sidebar_position: 1
---

# Lazy loading

Lazy loading limits the number of components that load in an app.

🚀 Lazy services are loaded in invocation order.

🐎 Lazy service invocation is protected against concurrent loading.

## Provider

A lazy service must be defined with a `provider` (kind of constructor). This provider will be called once and must return the singleton or an error.

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

A service can be injected in many ways. Here is the full list of lazy loading service injections.

```go
func Provide[T any](i do.Injector, provider do.Provider[T])
func ProvideNamed[T any](i do.Injector, name string, provider do.Provider[T])
func Override[T any](i do.Injector, provider Provider[T])
func OverrideNamed[T any](i do.Injector, name string, provider do.Provider[T])
```

:::info

It is highly recommended to use anonymous service invocation to let the framework handle naming.

:::

```go
i := do.New()

do.Provide(i, NewMyService)
// or
do.ProvideNamed(i, "my.really.cool.service", NewMyService)
```

## Error handling

On invocation, panics are caught by the framework and returned as an error.

## Hot service replacement

By default, providing a service twice will panic. Service can be replaced at runtime using `do.Override` helper.

For test purposes, it can be very useful to override a service given to the DI container.

:::warning

We strongly discourage using this helper in production. Please use service aliasing instead.

:::

```go
type CalculatorTestSuite struct {
    suite.Suite
    i do.Injector
}

func (suite *CalculatorTestSuite) SetupSuite() {
    suite.i = do.New()
    do.Provide(suite.i, NewMyService)
}

func (suite *CalculatorTestSuite) TearDownSuite() {
    suite.i.Shutdown()
}

func (suite *CalculatorTestSuite) SetupTest() {
    do.Override(suite.i, NewMyServiceMock)  // <- replace service by mock
}

func (suite *CalculatorTestSuite) Test1() {
  // ...
}

func (suite *CalculatorTestSuite) Test2() {
  // ...
}
```
