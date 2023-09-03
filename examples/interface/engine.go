package main

import "github.com/samber/do"

type Engine interface{}

type engineImplem struct {
}

func NewEngine(i do.Injector) (Engine, error) {
	return &engineImplem{}, nil
}
