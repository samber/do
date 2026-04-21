package main

import (
	"fmt"
)

// Database represents a database connection.
type Database struct {
	Config *Configuration
	URL    string
}

func (db *Database) Connect() {
	fmt.Printf("Connecting to database: %s\n", db.URL)
}
