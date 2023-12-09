package main

import "github.com/cryptoniumX/di"

type Engine interface{}

type engineImplem struct {
}

func NewEngine(i *di.Injector) (Engine, error) {
	return &engineImplem{}, nil
}
