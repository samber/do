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
   - `do.InvokeAs()` (single service)
   - `do.InvokeAsAll()` (all matching services)
2. **Explicit alias injection**:
   - provide struct, bind interface, invoke interface
   - `do.As()`

## Implicit invocation (preferred)

4 methods are available for implicit invocation:
- `do.InvokeAs` (single service)
- `do.MustInvokeAs` (single service, panics on error)
- `do.InvokeAsAll` (all matching services)
- `do.MustInvokeAsAll` (all matching services, panics on error)

Named invocation is not available for now. Feel free to open an issue to discuss your needs.

Implicit aliasing is recommended for production.

**Play: https://go.dev/play/p/29gb2TJG4m5**

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

## Bulk implicit invocation

For scenarios where you need to work with multiple services that implement the same interface, use `do.InvokeAsAll`:

```go
type Processor interface {
    Process(data string) error
}

type FileProcessor struct {}
func (f *FileProcessor) Process(data string) error { return nil }

type NetworkProcessor struct {}
func (n *NetworkProcessor) Process(data string) error { return nil }

i := do.New()

// Register multiple processors
do.Provide(i, func(i do.Injector) (*FileProcessor, error) {
    return &FileProcessor{}, nil
})
do.Provide(i, func(i do.Injector) (*NetworkProcessor, error) {
    return &NetworkProcessor{}, nil
})

// Invoke all processors
processors, err := do.InvokeAsAll[Processor](i)
if err != nil {
    log.Fatal(err)
}

// Process with all available processors
for _, processor := range processors {
    processor.Process("data")
}
```

### Characteristics of InvokeAsAll

- **Returns a slice**: `[]T` instead of single `T`
- **Deterministic ordering**: Services sorted by registration name
- **Scope inheritance**: Finds services across all scopes
- **Partial failures**: Returns successful services even if some fail
- **Empty results**: Valid empty slice when no services match

### When to use InvokeAsAll vs InvokeAs

| Scenario | Use InvokeAs | Use InvokeAsAll |
|----------|--------------|-----------------|
| Single service needed | ✅ | ❌ |
| Multiple services needed | ❌ | ✅ |
| Fail-fast on missing service | ✅ | ❌ |
| Graceful handling of zero services | ❌ | ✅ |
| Load balancing across services | ❌ | ✅ |

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

**Play: https://go.dev/play/p/_wGjnRJfwV8**

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
