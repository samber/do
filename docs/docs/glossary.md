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

`do.Injector` is either a `*do.RootScope`, a `*do.Scope` or a `*do.VirtualScope`.

## DI Container

A DI Container is another term for the Injector in Dependency Injection. It's responsible for providing instances of classes and their dependencies.

## Scope

A kind of module. It contains many declared singletons and service providers. It has access to services from ancestor scopes.

## Root scope

Top-level scope.

## Virtual scope

A chain of service invocations instantiates multiple virtual scopes to track dependency cycles.

Vitual Scope are not available under the public API. This is an internal stuff.

## Child scope

A scope that is nested within another scope. Variables defined in a child scope are only accessible within that scope and any nested scopes.

## DAG

Stands for Directed Acyclic Graph. It's a concept in mathematics and computer science. In the context of DI, it often refers to the graph of dependencies between different components or services.

## Provider

In DI, a provider is a component or a factory that creates instances of a service or a class.

## Injection

The process of providing a service to a scope. The injection can be done in different ways like provider injection or setter injection.

## Invocation

The process of executing a procedure or function through a call.

## Lazy Service

A service in DI that is not created until it is first requested.

## Eager Service

A service in DI that is created as soon as the application starts, not when it's first requested.

## Transient Service

A service in DI that is created anew each time it is requested.

## Service Alias

An alternative name given to a service in DI. It allows a service to be accessed using a different identifier.