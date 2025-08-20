package database

import (
	"fmt"
)

// Database represents a database connection
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
	return fmt.Sprintf("Query result: %s", sql)
}
