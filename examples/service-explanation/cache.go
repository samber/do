package main

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
