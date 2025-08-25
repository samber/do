package main

import "fmt"

// Cache represents a cache service.
type Cache struct {
	Data map[string]interface{}
}

func (c *Cache) Get(key string) interface{} {
	return c.Data[key]
}

func (c *Cache) Set(key string, value interface{}) {
	c.Data[key] = value
}

func (c *Cache) HealthCheck() error {
	return nil
}

func (c *Cache) Shutdown() error {
	fmt.Println("Shutting down cache")
	return nil
}
