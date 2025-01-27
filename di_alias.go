package do

import "fmt"

/////////////////////////////////////////////////////////////////////////////
// 							Explicit aliases
/////////////////////////////////////////////////////////////////////////////

// As declares an alias for a service.
func As[Initial any, Alias any](i Injector) error {
	initialName := NameOf[Initial]()
	aliasName := NameOf[Alias]()

	return AsNamed[Initial, Alias](i, initialName, aliasName)
}

// MustAs declares an alias for a service. It panics on error.
func MustAs[Initial any, Alias any](i Injector) {
	must0(As[Initial, Alias](i))
}

// AsNamed declares a named alias for a named service.
func AsNamed[Initial any, Alias any](i Injector, initial string, alias string) error {
	// first, we check if Initial can be cast to Alias
	_, ok := any(empty[Initial]()).(Alias)
	if !ok {
		return fmt.Errorf("DI: `%s` is not `%s`", initial, alias)
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

// AsNamed declares a named alias for a named service. It panics on error.
func MustAsNamed[Initial any, Alias any](i Injector, initial string, alias string) {
	must0(AsNamed[Initial, Alias](i, initial, alias))
}

/////////////////////////////////////////////////////////////////////////////
// 							Implicit aliases
/////////////////////////////////////////////////////////////////////////////

// InvokeAs invokes a service in the DI container. The last registered service matching the provided type or interface will be invoked.
func InvokeAs[T any](i Injector) (T, error) {
	return invokeByGenericType[T](i)
}

// MustInvokeAs invokes a service in the DI container. The last registered service matching the provided type or interface will be invoked. It panics on error.
func MustInvokeAs[T any](i Injector) T {
	return must1(InvokeAs[T](i))
}
