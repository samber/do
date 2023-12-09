package main

import "github.com/cryptoniumX/di"

/**
 * Wheel
 */
type Wheel struct {
}

/**
 * Engine
 */
type Engine struct {
}

/**
 * Car
 */
type Car struct {
	Engine *Engine
	Wheels []*Wheel
}

func (c *Car) Start() {
	println("vroooom")
}

/**
 * Run example
 */
func main() {
	injector := di.New()

	// provide wheels
	di.ProvideNamedValue(injector, "wheel-1", &Wheel{})
	di.ProvideNamedValue(injector, "wheel-2", &Wheel{})
	di.ProvideNamedValue(injector, "wheel-3", &Wheel{})
	di.ProvideNamedValue(injector, "wheel-4", &Wheel{})

	// provide car
	di.Provide(injector, func(i *di.Injector) (*Car, error) {
		car := Car{
			Engine: di.MustInvoke[*Engine](i),
			Wheels: []*Wheel{
				di.MustInvokeNamed[*Wheel](i, "wheel-1"),
				di.MustInvokeNamed[*Wheel](i, "wheel-2"),
				di.MustInvokeNamed[*Wheel](i, "wheel-3"),
				di.MustInvokeNamed[*Wheel](i, "wheel-4"),
			},
		}

		return &car, nil
	})

	// provide engine
	di.Provide(injector, func(i *di.Injector) (*Engine, error) {
		return &Engine{}, nil
	})

	// start car
	car := di.MustInvoke[*Car](injector)
	car.Start()
}
