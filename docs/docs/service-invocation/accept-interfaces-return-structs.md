---
title: Accept interfaces, return structs
description: Accept interfaces, return structs
sidebar_position: 2
---

# What the hell?

You have probably already heard the Go proverb "Accept interfaces, return structs". The basic idea is to let the consumer define the interfaces it uses while the producer returns concrete types.

This library offers a great API for assembling a large number of modules with loose coupling and a nice separation of concern.

Aliases must be interface and can be declared explicitly on injection using `do.As()`, or can be invoked implicitly using `do.InvokeAs()`.

## Implicit invocation (prefered)

2 methods are available for implicit invocation:
- `do.InvokeAs`
- `do.MustInvokeAs`

Named invocation is not available for now. Please open an issue to discuss about your needs.

Implicit alias is the way to go for production.

```go
type Metric interface {
    Inc()
}

type ReqPerSecond struct {
    counter int
}

func (r *ReqPerSecond) Inc() {
    m.counter++
}

i := do.New()

// inject the struct
Provide(i, func(i do.Injector) (*ReqPerSecond, error) {
    return &ReqPerSecond{}, nil
})

// invoke using the Metric interface
metric := do.MustInvokeAs[Metric](injector)
metric.Inc()    // <- r.counter will be incremented
```

:::info

The first matching service in the scope tree is returned.

:::

## Explicit injection

Explicit injection will be used in rare cases when you need to adapt to legacy code or change the catalog of services at runtime.

If you rely a lot on `do.Scope`, explicit injection might be used to define custom aliases in the scope tree branchs.

4 methods are available for explicit aliasing:
- `do.As`
- `do.MustAs`
- `do.NamedAs`
- `do.MustNamedAs`

```go
type Metric interface {
    Inc()
}

type ReqPerSecond struct {
    counter int
}

func (r *ReqPerSecond) Inc() {
    m.counter++
}

i := do.New()

// inject the struct
Provide(i, func(i do.Injector) (*ReqPerSecond, error) {
    return &ReqPerSecond{}, nil
})

// explicit aliasing
err := do.As[*ReqPerSecond, Metric](injector)
if err != nil {
    // ...
}

// invoke using the explicit alias
metric := do.MustInvoke[Metric](injector)
metric.Inc()    // <- r.counter will be incremented
```

:::danger

Explicit aliasing is pure shit and should be used very carefully. You've been warned. 😁

:::
