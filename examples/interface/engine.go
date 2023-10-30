package main

import "github.com/samber/do/v2"

type Engine interface{}

type engineImplem struct {
}

func NewEngine(i do.Injector) (*engineImplem, error) {
	return &engineImplem{}, nil
}
