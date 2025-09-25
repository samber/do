# Package System Example

This example demonstrates how to organize services into modular packages using the dependency injection container. It shows how to create reusable service packages that can be composed together to build complex applications.

## Features Demonstrated

- **Package System**: Using `do.Package` to group related services together
- **Modular Registration**: Organizing services into logical packages
- **Service Composition**: Combining multiple packages to build applications
- **Interface Aliasing**: Using `do.As` to create type aliases across packages
- **Dependency Resolution**: Services in packages can depend on services from other packages

## Playground Demo

**Play: https://go.dev/play/p/dUKDptcQ1a3**

## How it Works

1. **Configuration Package**: Provides application configuration
2. **Database Package**: Provides database and cache services
3. **Logging Package**: Provides logging services
4. **Services Package**: Provides business logic services
5. **Application Package**: Orchestrates all services into the main application

## Package Structure

- **`DatabasePackage`**: Database connection and cache services
- **`LoggingPackage`**: Logging services with configuration
- **`ServicesPackage`**: Business logic services (User, Order)
- **`ApplicationPackage`**: Main application orchestration

## Benefits of Package System

- **Modular Service Registration**: Services are organized into logical packages
- **Reusable Service Packages**: Packages can be reused across different applications
- **Clear Separation of Concerns**: Each package has a specific responsibility
- **Easy Testing and Mocking**: Packages can be easily replaced with test doubles
- **Composable Architecture**: Different combinations of packages can create different applications

## Use Cases

- **Microservices**: Each service can be its own package
- **Plugin Systems**: Different packages can be loaded based on configuration
- **Testing**: Test packages can replace production packages
- **Modular Applications**: Applications can be built by combining different packages
- **Team Development**: Different teams can work on different packages

This example shows how the package system enables modular, maintainable, and testable applications with clear separation of concerns.

