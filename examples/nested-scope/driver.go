package main

import (
	"github.com/samber/do/v2"
)

/**
 * Driver
 */
type Driver struct {
	Seat   *Seat
	Engine *Engine
}

func (d *Driver) TakeASeat() {
	println("driver enters the car")
	d.Engine.Start()
}

func (d *Driver) Shutdown() {
	println("driver leaves the car")
}

func bootDriverModule(injector do.Injector) {
	// provide driver
	do.Provide(injector, func(i do.Injector) (*Driver, error) {
		return &Driver{
			Seat:   do.MustInvokeNamed[*Seat](i, "seat-1"),
			Engine: do.MustInvoke[*Engine](i),
		}, nil
	})
}
