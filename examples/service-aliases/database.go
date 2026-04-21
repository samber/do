package main

import (
	"fmt"
)

// Database represents a concrete database implementation.
type Database struct {
	URL       string
	Connected bool
}

func (db *Database) Connect() error {
	fmt.Printf("Connecting to database: %s\n", db.URL)
	db.Connected = true
	return nil
}

func (db *Database) Query(sql string) string {
	return fmt.Sprintf("Query result from %s: %s", db.URL, sql)
}

// DatabaseInterface defines the interface for database operations.
type DatabaseInterface interface {
	Connect() error
	Query(string) string
}

// ReadOnlyDatabase defines a read-only interface for database operations.
type ReadOnlyDatabase interface {
	Query(string) string
}

// WriteDatabase defines a write interface for database operations.
type WriteDatabase interface {
	Connect() error
}
