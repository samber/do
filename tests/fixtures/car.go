package fixtures

import "github.com/samber/do/v2"

/**
 * driver
 */
func newDriver(i do.Injector) (*Driver, error) {
	d := &Driver{
		seat:   do.MustInvokeNamed[*Seat](i, "seat-1"),
		engine: do.MustInvoke[*Engine](i),
	}

	d.seat.take()
	d.engine.start()

	return d, nil
}

type Driver struct {
	seat   *Seat
	engine *Engine
}

func (d *Driver) Shutdown() {
	d.engine.stop()
	d.seat.release()
}

/**
 * passenger
 */
func newPassenger(seatName string) func(i do.Injector) (*Passenger, error) {
	return func(i do.Injector) (*Passenger, error) {
		p := &Passenger{
			seat: do.MustInvokeNamed[*Seat](i, seatName),
		}

		p.seat.take()

		return p, nil
	}
}

type Passenger struct {
	seat *Seat
}

func (p *Passenger) Shutdown() {
	p.seat.release()
}

/**
 * Seat
 */
type Seat struct {
	busy bool
}

func (s *Seat) take() {
	if s.busy {
		panic("seat should be free")
	}
	s.busy = true
}

func (s *Seat) release() {
	if !s.busy {
		panic("seat should be busy")
	}
	s.busy = false
}

func (s *Seat) Shutdown() {
	if s.busy {
		panic("seat should be free")
	}
}

/**
 * Wheel
 */
type Wheel struct {
	active bool
}

func (w *Wheel) start() {
	if w.active {
		panic("wheel should be stopped")
	}
	w.active = true
}

func (w *Wheel) stop() {
	if !w.active {
		panic("wheel should be started")
	}
	w.active = false
}

func (w *Wheel) Shutdown() {
	if w.active {
		panic("wheel should be stopped")
	}
}

/**
 * engine
 */
func newEngine(i do.Injector) (*Engine, error) {
	return &Engine{
		wheels: []*Wheel{
			do.MustInvokeNamed[*Wheel](i, "wheel-1"),
			do.MustInvokeNamed[*Wheel](i, "wheel-2"),
			do.MustInvokeNamed[*Wheel](i, "wheel-3"),
			do.MustInvokeNamed[*Wheel](i, "wheel-4"),
		},
	}, nil
}

type Engine struct {
	started bool
	wheels  []*Wheel
}

func (e *Engine) start() {
	if e.started {
		panic("engine should be stopped")
	}
	e.started = true

	for _, wheel := range e.wheels {
		wheel.start()
	}
}

func (e *Engine) stop() {
	if !e.started {
		panic("engine should be started")
	}
	e.started = false

	for _, wheel := range e.wheels {
		wheel.stop()
	}
}

func (e *Engine) Shutdown() {
	if e.started {
		panic("engine should be stopped")
	}
}

func GetPackage() (do.Injector, do.Injector, do.Injector) {
	injector := do.New()

	driverScope := injector.Scope("driver")
	passengerScope := injector.Scope("passenger")

	// provide wheels
	do.ProvideNamedValue(injector, "wheel-1", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-2", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-3", &Wheel{})
	do.ProvideNamedValue(injector, "wheel-4", &Wheel{})

	// provide engine
	do.Provide(injector, newEngine)

	// provide seats
	do.ProvideNamedValue(injector, "seat-1", &Seat{})
	do.ProvideNamedValue(injector, "seat-2", &Seat{})
	do.ProvideNamedValue(injector, "seat-3", &Seat{})
	do.ProvideNamedValue(injector, "seat-4", &Seat{})

	// provide driver
	do.Provide(driverScope, newDriver)

	// provide passenger
	do.ProvideNamed(passengerScope, "passenger-1", newPassenger("seat-2"))
	do.ProvideNamed(passengerScope, "passenger-2", newPassenger("seat-3"))
	do.ProvideNamed(passengerScope, "passenger-3", newPassenger("seat-4"))

	return injector, driverScope, passengerScope
}
