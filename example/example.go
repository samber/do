package main

import (
	"fmt"

	"github.com/samber/do"
)

/**
 * Wheel
 */
type Wheel struct {
}

/**
 * Engine
 */
type Engine interface{}

type engineImplem struct {
}

func (c *engineImplem) HealthCheck() error {
	return fmt.Errorf("engine broken")
}

func (c *engineImplem) Shutdown() error {
	println("engine stopped")
	return nil
}

/**
 * Car
 */
type Car interface {
	Start()
}

type carImplem struct {
	Engine Engine
	Wheels []*Wheel
}

func (c *carImplem) Shutdown() error {
	println("car stopped")
	return nil
}

func (c *carImplem) Start() {
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
	do.Provide(injector, func(i *do.Injector) (Car, error) {
		wheels := []*Wheel{
			do.MustInvokeNamed[*Wheel](i, "wheel-1"),
			do.MustInvokeNamed[*Wheel](i, "wheel-2"),
			do.MustInvokeNamed[*Wheel](i, "wheel-3"),
			do.MustInvokeNamed[*Wheel](i, "wheel-4"),
		}

		engine := do.MustInvoke[Engine](i)

		car := carImplem{
			Engine: engine,
			Wheels: wheels,
		}

		return &car, nil
	})

	// provide engine
	do.Provide(injector, func(i *do.Injector) (Engine, error) {
		return &engineImplem{}, nil
	})

	// start car
	car := do.MustInvoke[Car](injector)

	car.Start()

	fmt.Println(do.HealthCheck[Engine](injector))

	injector.Shutdown()
}
