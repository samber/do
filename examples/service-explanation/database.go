package main

import (
	"fmt"
)

// Database represents a database service
type Database struct {
	Config *Configuration
	URL    string
}

func (db *Database) Connect() error {
	fmt.Printf("Connecting to database: %s\n", db.URL)
	return nil
}

func (db *Database) Query(sql string) string {
	return fmt.Sprintf("Query result: %s", sql)
}
