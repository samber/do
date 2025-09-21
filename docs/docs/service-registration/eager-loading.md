---
title: Eager loading
description: Eager loading initializes a resource as soon as the code is executed.
sidebar_position: 2
---

# Eager loading

Eager loading initializes a resource as soon as the code is executed.

## Inject service into DI container

An eager service is simply initialized by you and then injected into the DI container.

A service can be injected in many ways. Here is the full list of eager loading service injections.

```go
func ProvideValue[T any](i do.Injector, value T)
func ProvideNamedValue[T any](i do.Injector, name string, value T)
func OverrideValue[T any](i do.Injector, value T)
func OverrideNamedValue[T any](i do.Injector, name string, value T)
```

:::info

It is highly recommended to use anonymous service invocation to let the framework handle naming.

:::

```go
type Config struct {
  Port int
}
config := &Config{
  Port: os.Getenv("PORT"),
}

i := do.New()

do.ProvideValue(i, config)
// or
do.ProvideNamedValue(i, "my.really.cool.config", config)
```

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
    do.ProvideNamedValue(suite.i, "a-number", 42)
}

func (suite *CalculatorTestSuite) TearDownSuite() {
    suite.i.Shutdown()
}

func (suite *CalculatorTestSuite) SetupTest() {
    do.OverrideNamedValue(suite.i, "a-number", 1337)  // <- replace service by mock
}

func (suite *CalculatorTestSuite) Test1() {
  // ...
}

func (suite *CalculatorTestSuite) Test2() {
  // ...
}
```
