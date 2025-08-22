---
title: Accept interfaces, return structs
description: Accept interfaces, return structs
sidebar_position: 2
---

# Interface binding

You have probably already heard the Go proverb "Accept interfaces, return structs". The basic idea is to let the consumer define the interfaces it uses while the producer returns concrete types.

This library offers a great API for assembling a large number of modules with loose coupling and a clear separation of concerns.

Aliases must be Go interfaces. It can be declared explicitly on injection using `do.As()`, or invoked implicitly using `do.InvokeAs()`.

1. **Implicit alias invocation**:
   - provide struct, invoke interface
   - `do.InvokeAs()`
2. **Explicit alias injection**:
   - provide struct, bind interface, invoke interface
   - `do.As()`

## Implicit invocation (preferred)

2 methods are available for implicit invocation:
- `do.InvokeAs`
- `do.MustInvokeAs`

Named invocation is not available for now. Feel free to open an issue to discuss your needs.

Implicit aliasing is recommended for production.

```go
type Metric interface {
    Inc()
}

type RequestPerSecond struct {
    counter int
}

func (r *RequestPerSecond) Inc() {
    r.counter++
}

i := do.New()

// inject the struct
do.Provide(i, func(i do.Injector) (*RequestPerSecond, error) {
    return &RequestPerSecond{}, nil
})

// invoke using the Metric interface
metric := do.MustInvokeAs[Metric](injector)
metric.Inc()    // <- r.counter will be incremented
```

:::info

The first matching service in the scope tree is returned.

:::

:::warning

Invoking an implicit alias with a very simple interface signature might lead to loading the wrong service.

Eg: `fmt.Stringer` might match dozens of services in a container.

:::

### InvokeStruct and implicit aliasing

When using `do.InvokeStruct`, if a field is tagged with an empty tag (eg: `` `do:""` ``) and the inferred name does not match a registered service, the injector will fall back to the first registered service that is assignable to the field type, equivalent to `do.InvokeAs[T]` resolution. Prefer explicit names when multiple assignable services exist.

## Explicit injection

Explicit injection will be used in rare cases when you need to adapt to legacy code or change the catalog of services at runtime.

If you rely a lot on `do.Scope`, explicit injection might be used to define custom aliases in the scope tree branches.

4 methods are available for explicit aliasing:
- `do.As`
- `do.MustAs`
- `do.AsNamed`
- `do.MustAsNamed`

```go
type Metric interface {
    Inc()
}

type RequestPerSecond struct {
    counter int
}

func (r *RequestPerSecond) Inc() {
    r.counter++
}

i := do.New()

// inject the struct
do.Provide(i, func(i do.Injector) (*RequestPerSecond, error) {
    return &RequestPerSecond{}, nil
})

// explicit aliasing
err := do.As[*RequestPerSecond, Metric](injector)
if err != nil {
    // ...
}

// invoke using the explicit alias
metric := do.MustInvoke[Metric](injector)
metric.Inc()    // <- r.counter will be incremented
```

:::danger

Explicit aliasing can lead to brittle designs and should be used very carefully. You've been warned. 😁

:::
