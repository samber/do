package main

import (
	"context"
	"log"

	"github.com/samber/do/v2"
)

/**
 * Wheel
 */
type Wheel struct {
}

/**
 * AutoPilot
 */
type AutoPilot struct {
}

func (a *AutoPilot) ShutdownWithContext(ctx context.Context) error {
	return nil
}

/**
 * Engine
 */
type Engine struct {
}

func (c *Engine) Shutdown() error {
	println("engine stopped")
	return nil
}

/**
 * Car
 */
type Car struct {
	AutoPilot *AutoPilot
	Engine    *Engine
	Wheels    []*Wheel
}

func (c *Car) Shutdown() error {
	println("car stopped")
	return nil
}

func (c *Car) Start() {
	println("vroooom")
}

/**
 * Run example
 */
func main() {
	injector := do.New()

	// provide wheels
	do.ProvideNamedValue(injector, "wheel-1", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-2", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-3", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-4", &Wheel{})

	// provide car
	do.Provide(injector, func(i do.Injector) (*Car, error) {
		car := Car{
			AutoPilot: do.MustInvoke[*AutoPilot](i),
			Engine:    do.MustInvoke[*Engine](i),
			Wheels: []*Wheel{
				do.MustInvokeNamed[*Wheel](i, "wheel-1"),
				do.MustInvokeNamed[*Wheel](i, "wheel-2"),
				do.MustInvokeNamed[*Wheel](i, "wheel-3"),
				do.MustInvokeNamed[*Wheel](i, "wheel-4"),
			},
		}

		return &car, nil
	})

	// provide engine
	do.Provide(injector, func(i do.Injector) (*Engine, error) {
		return &Engine{}, nil
	})

	// provide autopilot
	do.Provide(injector, func(i do.Injector) (*AutoPilot, error) {
		return &AutoPilot{}, nil
	})

	// start car
	car := do.MustInvoke[*Car](injector)
	car.Start()

	_, err := injector.ShutdownOnSignals()
	for _, e := range err {
		if e != nil {
			log.Fatal(e.Error())
		}
	}
}
