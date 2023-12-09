package main

import (
	"github.com/cryptoniumX/di"
)

func main() {
	injector := di.New()

	// provide wheels
	di.ProvideNamedValue(injector, "wheel-1", NewWheel())
	di.ProvideNamedValue(injector, "wheel-2", NewWheel())
	di.ProvideNamedValue(injector, "wheel-3", NewWheel())
	di.ProvideNamedValue(injector, "wheel-4", NewWheel())

	// provide car
	di.Provide(injector, NewCar)

	// provide engine
	di.Provide(injector, NewEngine)

	// start car
	car := di.MustInvoke[Car](injector)
	car.Start()
}
