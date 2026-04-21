package main

import (
	"github.com/samber/do/v2"
)

/**
 * Passenger
 */
type Passenger struct {
	Seat *Seat
}

func (p *Passenger) TakeASeat() {
	println("passenger enters the car")
	p.Seat.Warm()
}

func (p *Passenger) Shutdown() {
	println("passenger leaves the car")
}

/**
 * Run example
 */
func bootPassengerModule(injector do.Injector) {
	// provide passenger
	do.ProvideNamed(injector, "passenger-1", func(i do.Injector) (*Passenger, error) {
		return &Passenger{
			Seat: do.MustInvokeNamed[*Seat](i, "seat-2"),
		}, nil
	})
	do.ProvideNamed(injector, "passenger-2", func(i do.Injector) (*Passenger, error) {
		return &Passenger{
			Seat: do.MustInvokeNamed[*Seat](i, "seat-3"),
		}, nil
	})
	do.ProvideNamed(injector, "passenger-3", func(i do.Injector) (*Passenger, error) {
		return &Passenger{
			Seat: do.MustInvokeNamed[*Seat](i, "seat-4"),
		}, nil
	})
}
