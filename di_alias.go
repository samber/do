package do

import "fmt"

// As declares an alias for a service.
func As[Initial any, Alias any](i Injector) error {
	initialName := Name[Initial]()
	aliasName := Name[Alias]()

	return AsNamed[Initial, Alias](i, initialName, aliasName)
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

	provide(i, alias, 42, func(_ string, _ int) Service[Alias] {
		return newServiceAlias[Alias](alias, i, initial)
	})

	return nil
}
