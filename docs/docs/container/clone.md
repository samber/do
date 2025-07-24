---
title: Clone
description: Clone your global DI container
sidebar_position: 3
---

# Clone

Clone has the same service registrations as its parent, but it doesn't share the invoked service state.

Cloning a container can be very useful for test purposes.

```go
injector := do.New()

Provide[*Car](i, NewCar)
Provide[*Engine](i, NewEngine)

// reset scope
injector = injector.Clone()
```

## Clone with options

```go
injector := do.New()

Provide[*Car](i, NewCar)
Provide[Engine](i, NewEngine)

// clone
injector = injector.Clone()

// replace Engine by *MockEngine
do.Override[Engine](injector, func (i do.Injector) (Engine, error) {
    return &MockEngine{}, nil
})
```
