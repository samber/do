package main

import (
	"github.com/samber/do/v2"
)

func main() {
	injector := do.New()

	// provide wheels
	do.ProvideNamedValue(injector, "wheel-1", NewWheel())
	do.ProvideNamedValue(injector, "wheel-2", NewWheel())
	do.ProvideNamedValue(injector, "wheel-3", NewWheel())
	do.ProvideNamedValue(injector, "wheel-4", NewWheel())

	// provide car
	do.Provide(injector, NewCar)

	// provide engine
	do.Provide(injector, NewEngine)

	// magic happens here
	_ = do.As[*carImplem, Car](injector)
	_ = do.As[*engineImplem, Engine](injector)

	// start car
	car := do.MustInvoke[Car](injector)
	car.Start()
}
