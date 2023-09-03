package main

import "github.com/samber/do"

type Car interface {
	Start()
}

type carImplem struct {
	Engine Engine
	Wheels []*Wheel
}

func (c *carImplem) Start() {
	println("vroooom")
}

func NewCar(i do.Injector) (Car, error) {
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
}
