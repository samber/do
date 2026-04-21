package main

import (
	"github.com/samber/do/v2"
)

/**
 * Seat
 */
type Seat struct{}

func (s *Seat) Warm() {
	println("🔥🔥🔥")
}

func (s *Seat) Shutdown() {
	println("stopping seat")
}

/**
 * Wheel
 */
type Wheel struct{}

func (w *Wheel) Shutdown() {
	println("stopping wheel")
}

/**
 * Engine
 */
type Engine struct {
	Wheels []*Wheel
}

func (e *Engine) Start() {
	println("vroooom")
}

func (e *Engine) Shutdown() {
	println("stopping engine")
}

func bootCarModule(injector do.Injector) {
	// provide wheels
	do.ProvideNamedValue(injector, "wheel-1", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-2", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-3", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-4", &Wheel{})

	// provide engine
	do.Provide(injector, func(i do.Injector) (*Engine, error) {
		return &Engine{
			Wheels: []*Wheel{
				do.MustInvokeNamed[*Wheel](i, "wheel-1"),
				do.MustInvokeNamed[*Wheel](i, "wheel-2"),
				do.MustInvokeNamed[*Wheel](i, "wheel-3"),
				do.MustInvokeNamed[*Wheel](i, "wheel-4"),
			},
		}, nil
	})

	// provide seats
	do.ProvideNamedValue(injector, "seat-1", &Seat{})
	do.ProvideNamedValue(injector, "seat-2", &Seat{})
	do.ProvideNamedValue(injector, "seat-3", &Seat{})
	do.ProvideNamedValue(injector, "seat-4", &Seat{})
}
