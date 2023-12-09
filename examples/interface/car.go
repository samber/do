package main

import "github.com/cryptoniumX/di"

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

func NewCar(i *di.Injector) (Car, error) {
	wheels := []*Wheel{
		di.MustInvokeNamed[*Wheel](i, "wheel-1"),
		di.MustInvokeNamed[*Wheel](i, "wheel-2"),
		di.MustInvokeNamed[*Wheel](i, "wheel-3"),
		di.MustInvokeNamed[*Wheel](i, "wheel-4"),
	}

	engine := di.MustInvoke[Engine](i)

	car := carImplem{
		Engine: engine,
		Wheels: wheels,
	}

	return &car, nil
}
