package main

import (
	"fmt"
	"log"

	"github.com/samber/do/v2"
)

// Database interface
type Database interface {
	Name() string
}

// PostgresDB implementation
type PostgresDB struct{}

func (p *PostgresDB) Name() string { return "postgres" }

// MySQLDB implementation
type MySQLDB struct{}

func (m *MySQLDB) Name() string { return "mysql" }

func main() {
	injector := do.New()

	// Register multiple database implementations
	do.Provide(injector, func(i do.Injector) (*PostgresDB, error) {
		return &PostgresDB{}, nil
	})
	do.Provide(injector, func(i do.Injector) (*MySQLDB, error) {
		return &MySQLDB{}, nil
	})

	// Invoke all databases
	databases, err := do.InvokeAsAll[Database](injector)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d databases:\n", len(databases))
	for _, db := range databases {
		fmt.Printf("- %s\n", db.Name())
	}
	// Output:
	// Found 2 databases:
	// - mysql
	// - postgres
}
