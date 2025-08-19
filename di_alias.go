package do

import "fmt"

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
	if !canCastTo[Initial, Alias]() {
		return fmt.Errorf("DI: `%s` does not implement `%s`", alias, initial)
	}

	_i := getInjectorOrDefault(i)
	if ok := _i.serviceExistRec(initial); !ok {
		return fmt.Errorf("DI: service `%s` has not been declared", initial)
	}

	provide(i, alias, nil, func(_ string, _ any) Service[Alias] {
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
