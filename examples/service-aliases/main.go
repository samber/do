package main

import (
	"fmt"

	"github.com/samber/do/v2"
)

func main() {
	injector := do.New()

	fmt.Println("=== Service Aliases Example ===")
	fmt.Println("Service aliases allow the same concrete implementation to be")
	fmt.Println("registered under multiple interfaces, providing flexibility")
	fmt.Printf("and interface segregation.\n\n")

	// Step 1: Register the concrete database implementation
	fmt.Println("Step 1: Registering concrete Database implementation")
	do.Provide(injector, func(i do.Injector) (*Database, error) {
		return &Database{
			URL:       "postgres://localhost:5432/mydb",
			Connected: false,
		}, nil
	})

	// Step 2: Create aliases for different interfaces
	fmt.Println("Step 2: Creating aliases for different interfaces")
	fmt.Println("  - DatabaseInterface: Full database operations")
	fmt.Println("  - ReadOnlyDatabase: Read-only operations only")
	fmt.Println("  - WriteDatabase: Write operations only")

	_ = do.As[*Database, DatabaseInterface](injector)
	_ = do.As[*Database, ReadOnlyDatabase](injector)
	_ = do.As[*Database, WriteDatabase](injector)

	// Step 3: Register repositories that depend on DatabaseInterface
	fmt.Println("Step 3: Registering repositories (use DatabaseInterface)")
	do.Provide(injector, func(i do.Injector) (*UserRepository, error) {
		return &UserRepository{
			DB: do.MustInvoke[DatabaseInterface](i),
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*OrderRepository, error) {
		return &OrderRepository{
			DB: do.MustInvoke[DatabaseInterface](i),
		}, nil
	})

	// Step 4: Register services that depend on specific interfaces
	fmt.Println("Step 4: Registering services with specific interface requirements")
	do.Provide(injector, func(i do.Injector) (*AuditService, error) {
		return &AuditService{
			DB: do.MustInvoke[ReadOnlyDatabase](i), // Only needs read access
		}, nil
	})

	do.Provide(injector, func(i do.Injector) (*ConnectionManager, error) {
		return &ConnectionManager{
			DB: do.MustInvoke[WriteDatabase](i), // Only needs write access
		}, nil
	})

	// Step 5: Register the main application
	fmt.Println("Step 5: Registering main application")
	do.Provide(injector, func(i do.Injector) (*Application, error) {
		return &Application{
			UserRepo:      do.MustInvoke[*UserRepository](i),
			OrderRepo:     do.MustInvoke[*OrderRepository](i),
			AuditService:  do.MustInvoke[*AuditService](i),
			ConnectionMgr: do.MustInvoke[*ConnectionManager](i),
		}, nil
	})

	fmt.Println("\n=== Service Registration Complete ===")
	fmt.Println("Available services:", injector.ListProvidedServices())

	// Run the application
	fmt.Println("\n=== Running Application ===")
	app := do.MustInvoke[*Application](injector)
	app.Run()

	fmt.Println("\n=== Demonstrating Alias Functionality ===")
	fmt.Println("All aliases point to the same underlying database instance:")

	// Show that all aliases point to the same underlying database instance
	db1 := do.MustInvoke[DatabaseInterface](injector)
	db2 := do.MustInvoke[ReadOnlyDatabase](injector)

	// All should return the same query result since they're the same instance
	fmt.Println("DatabaseInterface query:", db1.Query("SELECT 1"))
	fmt.Println("ReadOnlyDatabase query:", db2.Query("SELECT 1"))

	// The WriteDatabase interface doesn't have Query method, but we can use it for connection
	fmt.Println("WriteDatabase can be used for connection management")

	fmt.Println("\n=== Service Aliases Benefits ===")
	fmt.Println("1. Interface segregation - services only see what they need")
	fmt.Println("2. Single implementation - one concrete type serves multiple interfaces")
	fmt.Println("3. Type safety - compile-time guarantees about available methods")
	fmt.Println("4. Flexibility - easy to change implementations without breaking consumers")
}
