package main

import "fmt"

// UserRepository represents a repository that uses the database.
type UserRepository struct {
	DB DatabaseInterface
}

func (r *UserRepository) GetUser(id string) string {
	return r.DB.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", id))
}

// OrderRepository represents a repository that uses the database.
type OrderRepository struct {
	DB DatabaseInterface
}

func (r *OrderRepository) GetOrder(id string) string {
	return r.DB.Query(fmt.Sprintf("SELECT * FROM orders WHERE id = %s", id))
}
