package do

import (
	"fmt"
	"reflect"
)

// DefaultStructTagKey is the default tag key used for struct field injection.
// When using struct injection, fields can be tagged with `do:""` or `do:"service-name"`
// to specify which service should be injected.
const DefaultStructTagKey = "do"

// Provider is a function type that creates and returns a service instance of type T.
// This is the core abstraction for service creation in the dependency injection container.
//
// The provider function receives an Injector instance that can be used to resolve
// dependencies for the service being created.
//
// Example:
//
//	func NewMyService(i do.Injector) (*MyService, error) {
//	    db := do.MustInvoke[*Database](i)
//	    config := do.MustInvoke[*Config](i)
//	    return &MyService{DB: db, Config: config}, nil
//	}
//
//	// Register the provider
//	do.Provide(injector, NewMyService)
type Provider[T any] func(Injector) (T, error)

// NameOf returns the name of the service in the DI container.
// This is highly discouraged to use this function, as your code
// should not declare any dependency explicitly.
//
// The function uses type inference to determine the service name
// based on the generic type parameter T.
//
// Play: https://go.dev/play/p/g549GqBbj-n
//
// Example:
//
//	serviceName := do.NameOf[*Database]()
func NameOf[T any]() string {
	return inferServiceName[T]()
}

// Provide registers a service in the DI container, using type inference.
// The service will be lazily instantiated when first requested.
//
// Play: https://go.dev/play/p/4JutUJ5Rqau
//
// Example:
//
//	do.Provide(injector, func(i do.Injector) (*MyService, error) {
//	    return &MyService{...}, nil
//	})
func Provide[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	ProvideNamed(i, name, provider)
}

// ProvideNamed registers a named service in the DI container.
// This allows you to register multiple services of the same type
// with different names for disambiguation.
//
// The service will be lazily instantiated when first requested.
//
// Play: https://go.dev/play/p/9JuTQhLGIlh
//
// Example:
//
//	do.ProvideNamed(injector, "main-db", func(i do.Injector) (*Database, error) {
//	    return &Database{URL: "postgres://main.acme.dev:5432/db"}, nil
//	})
//	do.ProvideNamed(injector, "backup-db", func(i do.Injector) (*Database, error) {
//	    return &Database{URL: "postgres://backup.acme.dev:5432/db"}, nil
//	})
func ProvideNamed[T any](i Injector, name string, provider Provider[T]) {
	provide(i, name, provider, func(s string, a Provider[T]) serviceWrapper[T] {
		return newServiceLazy(s, a)
	})
}

// ProvideValue registers a value in the DI container, using type inference to determine the service name.
// The value is immediately available and will not be recreated on each request.
//
// Play: https://go.dev/play/p/5TOSiI-c17Y
//
// Example:
//
//	ProvideValue(injector, &MyService{})
func ProvideValue[T any](i Injector, value T) {
	name := inferServiceName[T]()
	ProvideNamedValue(i, name, value)
}

// ProvideNamedValue registers a named value in the DI container.
// This allows you to register multiple values of the same type
// with different names for disambiguation.
//
// The value is immediately available and will not be recreated on each request.
//
// Example:
//
//	do.ProvideNamedValue(injector, "app-config", &Config{Port: 8080})
//	do.ProvideNamedValue(injector, "db-config", &Config{Port: 5432})
func ProvideNamedValue[T any](i Injector, name string, value T) {
	provide(i, name, value, func(s string, a T) serviceWrapper[T] {
		return newServiceEager(s, a)
	})
}

// ProvideTransient registers a factory in the DI container, using type inference to determine the service name.
// The service will be recreated each time it is requested, providing a fresh instance.
//
// Play: https://go.dev/play/p/hzOJwtNXwT9
//
// Example:
//
//	// Each invocation creates a new instance
//	do.ProvideTransient(injector, func(i do.Injector) (string, error) {
//	    return uuid.New().String(), nil
//	})
//
//	// First invocation
//	id1, _ := do.Invoke[string](injector)
//	// Second invocation - different instance
//	id2, _ := do.Invoke[string](injector)
//
//	fmt.Println(id1 != id2) // Output: true
func ProvideTransient[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	ProvideNamedTransient(i, name, provider)
}

// ProvideNamedTransient registers a named factory in the DI container.
// This allows you to register multiple transient services of the same type
// with different names for disambiguation.
//
// The service will be recreated each time it is requested, providing a fresh instance.
//
// Example:
//
//	// Each invocation creates a new instance
//	do.ProvideNamedTransient(injector, "request-id", func(i do.Injector) (string, error) {
//	    return uuid.New().String(), nil
//	})
//
//	// First invocation
//	id1, _ := do.InvokeNamed[string](injector, "request-id")
//	// Second invocation - different instance
//	id2, _ := do.InvokeNamed[string](injector, "request-id")
//
//	fmt.Println(id1 != id2) // Output: true
func ProvideNamedTransient[T any](i Injector, name string, provider Provider[T]) {
	provide(i, name, provider, func(s string, a Provider[T]) serviceWrapper[T] {
		return newServiceTransient(s, a)
	})
}

// provide is an internal helper function that handles the common logic
// for registering services in the DI container. It ensures that:
// - The injector is properly initialized
// - No duplicate service names are registered
// - The service is properly created and stored
// - Logging is performed for successful registration.
func provide[T any, A any](i Injector, name string, valueOrProvider A, serviceCtor func(string, A) serviceWrapper[T]) {
	_i := getInjectorOrDefault(i)
	if _i.serviceExist(name) {
		panic(fmt.Errorf("DI: service `%s` has already been declared", name))
	}

	service := serviceCtor(name, valueOrProvider)
	_i.serviceSet(name, service)

	_i.RootScope().opts.Logf("DI: service %s injected", name)
}

// Override replaces the service in the DI container, using type inference to determine the service name.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function is useful for testing or when you need to replace a service
// that has already been registered. However, be cautious as it may lead to
// resource leaks if the original service was already instantiated.
//
// Play: https://go.dev/play/p/g549GqBbj-n
func Override[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	OverrideNamed(i, name, provider)
}

// OverrideNamed replaces the named service in the DI container.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function allows you to replace a specific named service that has
// already been registered. Use with caution to avoid resource leaks.
//
// Play: https://go.dev/play/p/-gNF1BUEB5Q
func OverrideNamed[T any](i Injector, name string, provider Provider[T]) {
	override(i, name, provider, func(s string, a Provider[T]) serviceWrapper[T] {
		return newServiceLazy(s, a)
	})
}

// OverrideValue replaces the value in the DI container, using type inference to determine the service name.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function replaces an existing value service with a new one.
// The old value will not be properly cleaned up if it was already instantiated.
//
// Play: https://go.dev/play/p/-gNF1BUEB5Q
func OverrideValue[T any](i Injector, value T) {
	name := inferServiceName[T]()
	OverrideNamedValue(i, name, value)
}

// OverrideNamedValue replaces the named value in the DI container.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function allows you to replace a specific named value service.
// Use with caution to avoid resource leaks.
//
// Play: https://go.dev/play/p/-gNF1BUEB5Q
func OverrideNamedValue[T any](i Injector, name string, value T) {
	override(i, name, value, func(s string, a T) serviceWrapper[T] {
		return newServiceEager(s, a)
	})
}

// OverrideTransient replaces the factory in the DI container, using type inference to determine the service name.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function replaces an existing transient service factory with a new one.
// Since transient services are recreated on each request, this is generally safer
// than overriding lazy or eager services.
//
// Play: https://go.dev/play/p/_wYwBADbCaN
func OverrideTransient[T any](i Injector, provider Provider[T]) {
	name := inferServiceName[T]()
	OverrideNamedTransient(i, name, provider)
}

// OverrideNamedTransient replaces the named factory in the DI container.
// Warning: this will not unload/shutdown the previously invoked service.
//
// This function allows you to replace a specific named transient service factory.
// Since transient services are recreated on each request, this is generally safer
// than overriding lazy or eager services.
//
// Play: https://go.dev/play/p/_wYwBADbCaN
func OverrideNamedTransient[T any](i Injector, name string, provider Provider[T]) {
	override(i, name, provider, func(s string, a Provider[T]) serviceWrapper[T] {
		return newServiceTransient(s, a)
	})
}

// override is an internal helper function that handles the common logic
// for overriding services in the DI container. Unlike provide, it allows
// replacing existing services without throwing an error.
func override[T any, A any](i Injector, name string, valueOrProvider A, serviceCtor func(string, A) serviceWrapper[T]) {
	_i := getInjectorOrDefault(i)

	// Note: We don't check if the service exists here, allowing override
	service := serviceCtor(name, valueOrProvider)
	_i.serviceSet(name, service) // @TODO: should we unload/shutdown the previous service ?

	_i.RootScope().opts.Logf("DI: service %s overridden", name)
}

// Invoke retrieves and instantiates a service from the DI container using type inference.
// The service will be created if it hasn't been instantiated yet (for lazy services).
//
// Play: https://go.dev/play/p/4JutUJ5Rqau
//
// Example:
//
//	service, err := do.Invoke[*MyService](injector)
func Invoke[T any](i Injector) (T, error) {
	name := inferServiceName[T]()
	return invokeByName[T](i, name)
}

// InvokeNamed retrieves and instantiates a named service from the DI container.
// This allows you to retrieve specific named services when multiple services
// of the same type are registered.
//
// Play: https://go.dev/play/p/9JuTQhLGIlh
//
// Example:
//
//	// Register multiple databases
//	do.ProvideNamed(injector, "main-db", func(i do.Injector) (*Database, error) {
//	    return &Database{URL: "postgres://main.acme.dev:5432/db"}, nil
//	})
//	do.ProvideNamed(injector, "backup-db", func(i do.Injector) (*Database, error) {
//	    return &Database{URL: "postgres://backup.acme.dev:5432/db"}, nil
//	})
//
//	// Retrieve specific database
//	mainDB, err := do.InvokeNamed[*Database](injector, "main-db")
//	backupDB, err := do.InvokeNamed[*Database](injector, "backup-db")
func InvokeNamed[T any](i Injector, name string) (T, error) {
	if typeIsAssignable[T, any]() {
		v, err := invokeAnyByName(i, name)
		t, _ := v.(T) // just skip if v == nil
		return t, err
	}

	return invokeByName[T](i, name)
}

// MustInvoke retrieves and instantiates a service from the DI container using type inference.
// If the service cannot be retrieved or instantiated, it panics.
//
// This function is useful when you're certain the service exists and want
// to avoid error handling in your code.
//
// Play: https://go.dev/play/p/456pBhI36Q2
//
// Example:
//
//	service := do.MustInvoke[*MyService](injector)
func MustInvoke[T any](i Injector) T {
	return must1(Invoke[T](i))
}

// MustInvokeNamed retrieves and instantiates a named service from the DI container.
// If the service cannot be retrieved or instantiated, it panics.
//
// This function is useful when you're certain the named service exists and want
// to avoid error handling in your code.
//
// Play: https://go.dev/play/p/456pBhI36Q2
//
// Example:
//
//	service := do.MustInvokeNamed[*MyService](injector, "my-service")
func MustInvokeNamed[T any](i Injector, name string) T {
	return must1(InvokeNamed[T](i, name))
}

// InvokeStruct invokes services located in struct properties.
// The struct fields must be tagged with `do:""` or `do:"name"`, where `name` is the service name in the DI container.
// If the service is not found in the DI container, an error is returned.
// If the service is found but not assignable to the struct field, an error is returned.
//
// Play: https://go.dev/play/p/I3_Rznkprpj
//
// Example:
//
//	type App struct {
//	    Database *Database `do:""`
//	    Logger   *Logger   `do:"app-logger"`
//	    Config   *Config   `do:""`
//	}
//
//	// Register services
//	do.Provide(injector, func(i do.Injector) (*Database, error) {
//	    return &Database{}, nil
//	})
//	do.ProvideNamed(injector, "app-logger", func(i do.Injector) (*Logger, error) {
//	    return &Logger{}, nil
//	})
//	do.Provide(injector, func(i do.Injector) (*Config, error) {
//	    return &Config{}, nil
//	})
//
//	// Invoke struct with injected services
//	app, err := do.InvokeStruct[App](injector)
func InvokeStruct[T any](i Injector) (T, error) {
	structName := inferServiceName[T]()
	output := deepEmpty[T]() // if the struct is hidden behind a pointer, we need to init the struct value deep enough
	value := reflect.ValueOf(&output)

	for value.Elem().Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Check if the empty value is a struct (before passing a pointer to reflect.ValueOf).
	// It will be checked in invokeByTags, but the error message is different.
	if value.Kind() != reflect.Pointer || value.Elem().Kind() != reflect.Struct {
		return empty[T](), fmt.Errorf("DI: must be a struct or a pointer to a struct, but got `%s`", structName)
	}

	err := invokeByTags(i, structName, value, true)
	if err != nil {
		return empty[T](), err
	}

	return output, nil
}

// MustInvokeStruct invokes services located in struct properties and panics on error.
// The struct fields must be tagged with `do:""` or `do:"name"`, where `name` is the service name in the DI container.
// If the service is not found in the DI container, it panics.
// If the service is found but not assignable to the struct field, it panics.
//
// Play: https://go.dev/play/p/lRKqRT9TQVf
//
// Example:
//
//	type App struct {
//	    Database *Database `do:""`
//	    Logger   *Logger   `do:"app-logger"`
//	    Config   *Config   `do:""`
//	}
//
//	// Register services
//	do.Provide(injector, func(i do.Injector) (*Database, error) {
//	    return &Database{}, nil
//	})
//	do.ProvideNamed(injector, "app-logger", func(i do.Injector) (*Logger, error) {
//	    return &Logger{}, nil
//	})
//	do.Provide(injector, func(i do.Injector) (*Config, error) {
//	    return &Config{}, nil
//	})
//
//	// Invoke struct with injected services (panics on error)
//	app := do.MustInvokeStruct[App](injector)
func MustInvokeStruct[T any](i Injector) T {
	return must1(InvokeStruct[T](i))
}

/////////////////////////////////////////////////////////////////////////////
// 							Explicit aliases
/////////////////////////////////////////////////////////////////////////////

// As declares an alias for a service, allowing it to be retrieved using a different type.
// This function creates a type alias where the Alias type can be used to retrieve the Initial service.
// The alias is automatically named using the type name of Alias.
//
// Parameters:
//   - i: The injector to register the alias in
//
// Returns an error if the alias cannot be created (e.g., type incompatibility or missing service).
//
// Play: https://go.dev/play/p/T5zKBRZaZhj
//
// Example:
//
//	// Register a concrete service
//	do.Provide(injector, func(i do.Injector) (*PostgresqlDatabase, error) {
//	    return &PostgresqlDatabase{}, nil
//	})
//
//	// Create an alias so it can be retrieved as an interface
//	do.As[*PostgresqlDatabase, Database](injector)
//
//	// Now both work:
//	db := do.MustInvoke[*PostgresqlDatabase](injector)
//	db := do.MustInvoke[Database](injector)
func As[Initial any, Alias any](i Injector) error {
	initialName := NameOf[Initial]()
	aliasName := NameOf[Alias]()

	return AsNamed[Initial, Alias](i, initialName, aliasName)
}

// MustAs declares an alias for a service and panics if an error occurs.
// This is a convenience function that wraps As and panics on error.
//
// Parameters:
//   - i: The injector to register the alias in
//
// Panics if the alias cannot be created (e.g., type incompatibility or missing service).
//
// Play: https://go.dev/play/p/_wGjnRJfwV8
//
// Example:
//
//	do.MustAs[*PostgresqlDatabase, Database](injector)
func MustAs[Initial any, Alias any](i Injector) {
	must0(As[Initial, Alias](i))
}

// AsNamed declares a named alias for a named service.
// This function allows you to create aliases with custom names for both the initial service and the alias.
//
// Parameters:
//   - i: The injector to register the alias in
//   - initial: The name of the existing service to alias
//   - alias: The name for the new alias service
//
// Returns an error if the alias cannot be created (e.g., type incompatibility or missing service).
//
// Play: https://go.dev/play/p/h1R5rxKizwR
//
// Example:
//
//	// Register a service with a custom name
//	do.ProvideNamed(injector, "my-db", func(i do.Injector) (*PostgresqlDatabase, error) {
//	    return &PostgresqlDatabase{}, nil
//	})
//
//	// Create an alias with custom names
//	do.AsNamed[*PostgresqlDatabase, Database](injector, "my-db", "db-interface")
//
//	// Retrieve using the alias name
//	db := do.MustInvokeNamed[Database](injector, "db-interface")
func AsNamed[Initial any, Alias any](i Injector, initial string, alias string) error {
	// first, we check if Initial can be cast to Alias
	if !genericCanCastToGeneric[Initial, Alias]() {
		return fmt.Errorf("DI: `%s` does not implement `%s`", initial, alias)
	}

	_i := getInjectorOrDefault(i)
	if ok := _i.serviceExistRec(initial); !ok {
		return fmt.Errorf("DI: service `%s` has not been declared", initial)
	}

	provide(i, alias, nil, func(_ string, _ any) serviceWrapper[Alias] {
		return newServiceAlias[Initial, Alias](alias, i, initial)
	})

	return nil
}

// MustAsNamed declares a named alias for a named service and panics if an error occurs.
// This is a convenience function that wraps AsNamed and panics on error.
//
// Parameters:
//   - i: The injector to register the alias in
//   - initial: The name of the existing service to alias
//   - alias: The name for the new alias service
//
// Panics if the alias cannot be created (e.g., type incompatibility or missing service).
//
// Play: https://go.dev/play/p/8QI2XUm9yLH
//
// Example:
//
//	do.MustAsNamed[*PostgresqlDatabase, Database](injector, "my-db", "db-interface")
func MustAsNamed[Initial any, Alias any](i Injector, initial string, alias string) {
	must0(AsNamed[Initial, Alias](i, initial, alias))
}

/////////////////////////////////////////////////////////////////////////////
// 							Implicit aliases
/////////////////////////////////////////////////////////////////////////////

// InvokeAs invokes a service in the DI container by finding the first service that matches the provided type or interface.
// This function searches through all registered services to find one that can be cast to the requested type T.
// It's useful when you want to retrieve a service by interface without explicitly creating aliases.
//
// Parameters:
//   - i: The injector to search for the service
//
// Returns the service instance and any error that occurred during invocation.
//
// Play: https://go.dev/play/p/1uXvNenBbgk
//
// Example:
//
//	// Register a concrete service
//	do.Provide(injector, func(i do.Injector) (*PostgresqlDatabase, error) {
//	    return &PostgresqlDatabase{}, nil
//	})
//
//	// Retrieve by interface
//	db, err := do.InvokeAs[Database](injector)
func InvokeAs[T any](i Injector) (T, error) {
	return invokeByGenericType[T](i)
}

// MustInvokeAs invokes a service in the DI container by finding the first service that matches the provided type or interface.
// This function panics if an error occurs during invocation.
// It's useful when you want to retrieve a service by interface without explicitly creating aliases.
//
// Parameters:
//   - i: The injector to search for the service
//
// Returns the service instance.
// Panics if the service cannot be found or invoked.
//
// Play: https://go.dev/play/p/29gb2TJG4m5
//
// Example:
//
//	// Register a concrete service
//	do.Provide(injector, func(i do.Injector) (*PostgresqlDatabase, error) {
//	    return &PostgresqlDatabase{}, nil
//	})
//
//	// Retrieve by interface
//	db := do.MustInvokeAs[Database](injector)
func MustInvokeAs[T any](i Injector) T {
	return must1(InvokeAs[T](i))
}

// InvokeAsAll invokes all services in the DI container that match the provided type or interface.
// This function searches through all registered services to find all that can be cast to the requested type T.
// Returns a slice of all matching services in deterministic order by service name.
//
// Parameters:
//   - i: The injector to search for services
//
// Returns a slice of service instances and any error that occurred during invocation.
// If no services match, returns an empty slice and no error.
// If some services fail to invoke, returns successfully invoked services with an error describing the failures.
//
// Play: https://go.dev/play/p/spRqDOsXQLs
//
// Example:
//
//	// Register multiple database implementations
//	do.Provide(injector, func(i do.Injector) (*PostgresDB, error) {
//	    return &PostgresDB{}, nil
//	})
//	do.Provide(injector, func(i do.Injector) (*MySQLDB, error) {
//	    return &MySQLDB{}, nil
//	})
//	// Both implement Database interface
//
//	// Invoke all databases
//	databases, err := do.InvokeAsAll[Database](injector)
//	// databases contains both PostgresDB and MySQLDB instances
func InvokeAsAll[T any](i Injector) ([]T, error) {
	return invokeAsAllByGenericType[T](i)
}

// MustInvokeAsAll invokes all services in the DI container that match the provided type or interface.
// This function panics if an error occurs during invocation.
// Returns a slice of all matching service instances in deterministic order.
//
// Parameters:
//   - i: The injector to search for services
//
// Returns a slice of service instances.
// Panics if any service cannot be found or invoked.
//
// Play: https://go.dev/play/p/spRqDOsXQLs
//
// Example:
//
//	// Register multiple repositories
//	do.Provide(injector, func(i do.Injector) (*UserRepository, error) {
//	    return &UserRepository{}, nil
//	})
//	do.Provide(injector, func(i do.Injector) (*ProductRepository, error) {
//	    return &ProductRepository{}, nil
//	})
//	// Both implement Repository interface
//
//	// Invoke all repositories (panics on error)
//	repositories := do.MustInvokeAsAll[Repository](injector)
func MustInvokeAsAll[T any](i Injector) []T {
	return must1(InvokeAsAll[T](i))
}

/////////////////////////////////////////////////////////////////////////////
// 							Package-level declaration
/////////////////////////////////////////////////////////////////////////////

// Package creates a function that executes multiple service registration functions.
// This function is used to group related service registrations into reusable packages
// that can be applied to any injector instance.
//
// Parameters:
//   - services: Variable number of service registration functions to execute
//
// Returns a function that can be passed to New(), NewWithOpts(), or Scope() to register all services.
//
// Play: https://go.dev/play/p/kmf8aOVyj96
//
// Example:
//
//		 // pkg/database/package.go
//			// Create a global `Package` variable
//			var Package = do.Package(
//			    do.Lazy[*Database](func(i do.Injector) (*Database, error) {
//			        return &Database{}, nil
//			    }),
//			    do.Lazy[*ConnectionPool](func(i do.Injector) (*ConnectionPool, error) {
//			        return &ConnectionPool{}, nil
//			    }),
//	         ...
//			)
//
//		 // main.go
//			// Apply the package to an injector
//			injector := do.New(database.Package)
func Package(services ...func(i Injector)) func(Injector) {
	return func(injector Injector) {
		for i := range services {
			services[i](injector)
		}
	}
}

// Lazy creates a function that registers a lazy service using the default service name.
// This function is a convenience wrapper for creating lazy service registration functions
// that can be used in packages.
//
// Parameters:
//   - p: The provider function that creates the service instance
//
// Returns a function that registers the service as lazy when executed.
//
// Play: https://go.dev/play/p/M6-wd1qt-GZ
//
// Example:
//
//	// Global to a package
//	var Package = do.Package(
//		do.Lazy[*Database](func(i do.Injector) (*Database, error) {
//	    	return &Database{}, nil
//		}),
//	)
func Lazy[T any](p Provider[T]) func(Injector) {
	return func(injector Injector) {
		Provide(injector, p)
	}
}

// LazyNamed creates a function that registers a lazy service with a custom name.
// This function is a convenience wrapper for creating named lazy service registration functions
// that can be used in packages.
//
// Parameters:
//   - serviceName: The custom name for the service
//   - p: The provider function that creates the service instance
//
// Returns a function that registers the service as lazy with the specified name when executed.
//
// Play: https://go.dev/play/p/xAs-exXR9Sz
//
// Example:
//
//	// Global to a package
//	var Package = do.Package(
//		do.LazyNamed[*Database]("main-db", func(i do.Injector) (*Database, error) {
//	    	return &Database{}, nil
//		}),
//	)
func LazyNamed[T any](serviceName string, p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideNamed(injector, serviceName, p)
	}
}

// Eager creates a function that registers an eager service using the default service name.
// This function is a convenience wrapper for creating eager service registration functions
// that can be used in packages. The service value is provided directly.
//
// Parameters:
//   - value: The service instance to register eagerly
//
// Returns a function that registers the service as eager when executed.
//
// Play: https://go.dev/play/p/M6-wd1qt-GZ
//
// Example:
//
//	// Global to a package
//	var Package = do.Package(
//		do.Eager[*Config](&Config{Port: 8080})
//	)
func Eager[T any](value T) func(Injector) {
	return func(injector Injector) {
		ProvideValue(injector, value)
	}
}

// EagerNamed creates a function that registers an eager service with a custom name.
// This function is a convenience wrapper for creating named eager service registration functions
// that can be used in packages. The service value is provided directly.
//
// Parameters:
//   - serviceName: The custom name for the service
//   - value: The service instance to register eagerly
//
// Returns a function that registers the service as eager with the specified name when executed.
//
// Play: https://go.dev/play/p/pvByI4EkFEJ
//
// Example:
//
//	// Global to a package
//	var Package = do.Package(
//		do.EagerNamed[*Config]("app-config", &Config{Port: 8080})
//	)
func EagerNamed[T any](serviceName string, value T) func(Injector) {
	return func(injector Injector) {
		ProvideNamedValue(injector, serviceName, value)
	}
}

// Transient creates a function that registers a transient service using the default service name.
// This function is a convenience wrapper for creating transient service registration functions
// that can be used in packages.
//
// Parameters:
//   - p: The provider function that creates the service instance
//
// Returns a function that registers the service as transient when executed.
//
// Play: https://go.dev/play/p/M6-wd1qt-GZ
//
// Example:
//
//	// Global to a package
//	var Package = do.Package(
//		do.Transient[*Logger](func(i do.Injector) (*Logger, error) {
//	    	return &Logger{}, nil
//		})
//	)
func Transient[T any](p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideTransient(injector, p)
	}
}

// TransientNamed creates a function that registers a transient service with a custom name.
// This function is a convenience wrapper for creating named transient service registration functions
// that can be used in packages.
//
// Parameters:
//   - serviceName: The custom name for the service
//   - p: The provider function that creates the service instance
//
// Returns a function that registers the service as transient with the specified name when executed.
//
// Play: https://go.dev/play/p/d9IJOAbKRUH
//
// Example:
//
//	// Global to a package
//	var Package = do.Package(
//		do.TransientNamed[*Logger]("request-logger", func(i do.Injector) (*Logger, error) {
//	    	return &Logger{}, nil
//		})
//	)
func TransientNamed[T any](serviceName string, p Provider[T]) func(Injector) {
	return func(injector Injector) {
		ProvideNamedTransient(injector, serviceName, p)
	}
}

// Bind creates a function that creates a type alias between two types.
// This function is a convenience wrapper for creating service binding functions
// that can be used in packages.
//
// Parameters:
//   - Initial: The original type to bind from
//   - Alias: The type to bind to (must be implemented by Initial)
//
// Returns a function that creates the type alias when executed.
// Panics if the binding cannot be created.
//
// Play: https://go.dev/play/p/j69I52whJr2
//
// Example:
//
//	// Global to a package
//	var Package = do.Package(
//		do.Bind[*Database, DatabaseInterface]()
//	)
func Bind[Initial any, Alias any]() func(Injector) {
	return func(injector Injector) {
		MustAs[Initial, Alias](injector)
	}
}

// BindNamed creates a function that creates a named type alias between two types.
// This function is a convenience wrapper for creating named service binding functions
// that can be used in packages.
//
// Parameters:
//   - initial: The name of the original service
//   - alias: The name for the alias service
//   - Initial: The original type to bind from
//   - Alias: The type to bind to (must be implemented by Initial)
//
// Returns a function that creates the named type alias when executed.
// Panics if the binding cannot be created.
//
// Play: https://go.dev/play/p/m2gW7s9_qgq
//
// Example:
//
//	// Global to a package
//	var Package = do.Package(
//		do.BindNamed[*Database, DatabaseInterface]("main-db", "db-interface")
//	)
func BindNamed[Initial any, Alias any](initial string, alias string) func(Injector) {
	return func(injector Injector) {
		MustAsNamed[Initial, Alias](injector, initial, alias)
	}
}
