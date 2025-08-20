package services

import (
	"fmt"
)

// OrderService represents an order service
type OrderService struct {
	DB     Database
	Cache  Cache
	Logger Logger
}

func (o *OrderService) GetOrder(id string) string {
	o.Logger.Log(fmt.Sprintf("Getting order with ID: %s", id))

	// Try cache first
	if cached := o.Cache.Get("order:" + id); cached != nil {
		return fmt.Sprintf("Cached order: %v", cached)
	}

	// Query database
	result := o.DB.Query(fmt.Sprintf("SELECT * FROM orders WHERE id = %s", id))

	// Cache the result
	o.Cache.Set("order:"+id, result)

	return result
}
