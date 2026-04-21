---
id: glossary
title: 📚 Glossary
description: Dependency injection glossary
sidebar_position: 6
---

# Glossary

## Dependency Injection (DI)

A design pattern used in software development to achieve Inversion of Control (IoC) between classes and their dependencies. It's a technique for achieving loose coupling between objects and their collaborators, or dependencies.

## Inversion of Control (IoC)

A design principle where the flow of control is inverted compared to traditional procedural programming. Instead of the application-specific code controlling the execution of reusable code, the reusable code controls the execution. This principle is often implemented using techniques such as Dependency Injection, leading to more modular and easily maintainable code.

## Injector

In Dependency Injection (DI), an injector is a component that creates instances of classes and manages their dependencies. It's also known as the DI container or IoC (Inversion of Control) container.

`do.Injector` is implemented by `*do.RootScope` and `*do.Scope`. `*do.virtualScope` are internal implementation details.

## DI Container

A DI Container is another term for the Injector in Dependency Injection. It's responsible for providing instances of classes and their dependencies.

## Scope

A kind of module. It contains many declared singletons and service providers. It has access to services from ancestor scopes.

## Root scope

Top-level scope.

## Virtual scope (internal)

A chain of service invocations instantiates internal virtual scopes to track dependency cycles and build the dependency graph. These are not exposed in the public API.

## Child scope

A scope that is nested within another scope. Variables defined in a child scope are only accessible within that scope and any nested scopes.

## Directed Acyclic Graph (DAG)

DAG stands for Directed Acyclic Graph. It's a concept in mathematics and computer science. In the context of DI, it refers to the graph of dependencies between different components or services. The "acyclic" property ensures that there are no circular dependencies, meaning no service can depend on itself either directly or indirectly through other services.

## Circular Dependency

A situation where two or more services depend on each other, either directly or indirectly, creating a dependency cycle. Circular dependencies are not allowed in DI systems as they prevent proper initialization and can lead to infinite loops. The framework detects and prevents circular dependencies to maintain a valid DAG structure.

## Singleton

A design pattern where only one instance of a service exists throughout the application's lifecycle. In DI, singleton services are created once and reused for all subsequent requests. This ensures consistent state and efficient resource usage.

## Factory

A design pattern that provides an interface for creating objects without specifying their exact classes. In DI, a factory is a provider function that creates and returns instances of services. Factories can be used for lazy loading, transient services, or when complex initialization logic is required.

## Provider

In DI, a provider is a component or a factory that creates instances of a service or a class.

## Injection

The process of providing a service to a scope. The injection can be done in different ways like provider injection or setter injection.

## Invocation

The process of executing a procedure or function through a call.

## Interface Binding

A technique in DI that allows a concrete implementation to be bound to an interface type. This enables loose coupling by allowing services to depend on interfaces rather than concrete implementations. Interface binding can be done explicitly using `do.As()` or implicitly using `do.InvokeAs()`.

This pattern follows the Go motto "Accept interfaces, return structs", which means:
- **Accept interfaces**: Functions and methods should accept interface types as parameters, allowing for flexibility and testability
- **Return structs**: Functions and methods should return concrete struct types, providing clear and specific implementations

In DI, this means you can register concrete struct implementations but invoke them using interface types, promoting loose coupling and making your code more modular and testable.

## Struct Tag

A metadata mechanism in Go that allows attaching information to struct fields using backtick-enclosed strings. In DI, struct tags (like `do:""`) are used to specify how fields should be injected, including service names and injection strategies. The framework uses reflection to read these tags and automatically inject dependencies.

## Service

A component or object that provides specific functionality or behavior within a DI container. Services are the building blocks of a DI application and can be any type of object that other components depend on. Services are registered with the DI container and can be retrieved (invoked) when needed by other services or application code.

## Lazy Service

A service in DI that is not created until it is first requested.

## Eager Service

A service in DI that is created as soon as the application starts, not when it's first requested.

## Transient Service

A service in DI that is created anew each time it is requested.

## Healthcheck

A mechanism for monitoring the health and status of services in a DI container. Services can implement health check interfaces (`do.Healthchecker` or `do.HealthcheckerWithContext`) to provide status information about their internal state, such as database connectivity or resource availability.

## Graceful Shutdown

A shutdown mechanism that allows services to clean up resources and complete ongoing operations before the application terminates. Services can implement shutdown interfaces (`do.Shutdowner`, `do.ShutdownerWithError`, `do.ShutdownerWithContext`, or `do.ShutdownerWithContextAndError`) to define cleanup behavior. The framework ensures shutdown occurs in reverse initialization order.

## Service Alias

An alternative name given to a service in DI. It allows a service to be accessed using a different identifier.

## Package Loader

A mechanism that groups multiple service registrations into a single unit that can be imported and registered with a DI container. Package loaders use `do.Package()` to assemble services with different loading strategies (lazy, eager, transient) and can include named services and interface bindings. This allows for modular service organization where related services can be bundled together and imported as a cohesive unit, promoting modularity and reusability across different parts of an application.

