package do

import "context"

// Healthchecker is an interface that services can implement to provide health checking capabilities.
// Services implementing this interface can be health-checked by the DI container to ensure they
// are functioning correctly.
//
// The HealthCheck method should perform a quick check to verify the service is healthy.
// It should return nil if the service is healthy, or an error describing the health issue.
//
// Example:
//
//	type Database struct {
//	    conn *sql.DB
//	}
//
//	func (db *Database) HealthCheck() error {
//	    return db.conn.Ping()
//	}
type Healthchecker interface {
	HealthCheck() error
}

// HealthcheckerWithContext is an interface that services can implement to provide health checking
// capabilities with context support. This allows for timeout and cancellation control during
// health checks.
//
// The HealthCheck method should perform a quick check to verify the service is healthy.
// It should respect the provided context for cancellation and timeout.
// It should return nil if the service is healthy, or an error describing the health issue.
//
// Example:
//
//	type Database struct {
//	    conn *sql.DB
//	}
//
//	func (db *Database) HealthCheck(ctx context.Context) error {
//	    return db.conn.PingContext(ctx)
//	}
type HealthcheckerWithContext interface {
	HealthCheck(context.Context) error
}

// Shutdowner is an interface that services can implement to provide graceful shutdown capabilities.
// Services implementing this interface will be called during container shutdown to perform
// cleanup operations.
//
// The Shutdown method should perform any necessary cleanup, such as closing connections,
// flushing buffers, or stopping background processes.
//
// Example:
//
//	type Logger struct {
//	    file *os.File
//	}
//
//	func (l *Logger) Shutdown() {
//	    l.file.Close()
//	}
type Shutdowner interface {
	Shutdown()
}

// ShutdownerWithError is an interface that services can implement to provide graceful shutdown
// capabilities with error reporting. This allows services to report any errors that occur
// during shutdown.
//
// The Shutdown method should perform any necessary cleanup and return an error if the
// shutdown process fails.
//
// Example:
//
//	type Database struct {
//	    conn *sql.DB
//	}
//
//	func (db *Database) Shutdown() error {
//	    return db.conn.Close()
//	}
type ShutdownerWithError interface {
	Shutdown() error
}

// ShutdownerWithContext is an interface that services can implement to provide graceful shutdown
// capabilities with context support. This allows for timeout and cancellation control during
// shutdown operations.
//
// The Shutdown method should perform any necessary cleanup and respect the provided context
// for cancellation and timeout.
//
// Example:
//
//	type Server struct {
//	    srv *http.Server
//	}
//
//	func (s *Server) Shutdown(ctx context.Context) {
//	    s.srv.Shutdown(ctx)
//	}
type ShutdownerWithContext interface {
	Shutdown(context.Context)
}

// ShutdownerWithContextAndError is an interface that services can implement to provide graceful
// shutdown capabilities with both context support and error reporting. This is the most flexible
// shutdown interface, allowing for timeout control and error reporting.
//
// The Shutdown method should perform any necessary cleanup, respect the provided context
// for cancellation and timeout, and return an error if the shutdown process fails.
//
// Example:
//
//	type Server struct {
//	    srv *http.Server
//	}
//
//	func (s *Server) Shutdown(ctx context.Context) error {
//	    return s.srv.Shutdown(ctx)
//	}
type ShutdownerWithContextAndError interface {
	Shutdown(context.Context) error
}

// HealthCheck returns a service status, using type inference to determine the service name.
// This function performs a health check on a service by inferring its name from the type T.
// The service must implement either Healthchecker or HealthcheckerWithContext interface.
//
// Parameters:
//   - i: The injector containing the service
//
// Returns an error if the health check fails, or nil if the service is healthy.
//
// Example:
//
//	err := do.HealthCheck[*Database](injector)
//	if err != nil {
//	    log.Printf("Database health check failed: %v", err)
//	}
func HealthCheck[T any](i Injector) error {
	name := inferServiceName[T]()
	return HealthCheckNamedWithContext(context.Background(), i, name)
}

// HealthCheckWithContext returns a service status, using type inference to determine the service name.
// This function performs a health check on a service with context support for timeout and cancellation.
// The service must implement either Healthchecker or HealthcheckerWithContext interface.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - i: The injector containing the service
//
// Returns an error if the health check fails, or nil if the service is healthy.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	err := do.HealthCheckWithContext[*Database](ctx, injector)
//	if err != nil {
//	    log.Printf("Database health check failed: %v", err)
//	}
func HealthCheckWithContext[T any](ctx context.Context, i Injector) error {
	name := inferServiceName[T]()
	return HealthCheckNamedWithContext(ctx, i, name)
}

// HealthCheckNamed returns a service status for a named service.
// This function performs a health check on a service with the specified name.
// The service must implement either Healthchecker or HealthcheckerWithContext interface.
//
// Parameters:
//   - i: The injector containing the service
//   - name: The name of the service to health check
//
// Returns an error if the health check fails, or nil if the service is healthy.
//
// Example:
//
//	err := do.HealthCheckNamed(injector, "main-database")
//	if err != nil {
//	    log.Printf("Main database health check failed: %v", err)
//	}
func HealthCheckNamed(i Injector, name string) error {
	return HealthCheckNamedWithContext(context.Background(), i, name)
}

// HealthCheckNamedWithContext returns a service status for a named service with context support.
// This function performs a health check on a service with the specified name and context support
// for timeout and cancellation. The service must implement either Healthchecker or
// HealthcheckerWithContext interface.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - i: The injector containing the service
//   - name: The name of the service to health check
//
// Returns an error if the health check fails, or nil if the service is healthy.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	err := do.HealthCheckNamedWithContext(ctx, injector, "main-database")
//	if err != nil {
//	    log.Printf("Main database health check failed: %v", err)
//	}
func HealthCheckNamedWithContext(ctx context.Context, i Injector, name string) error {
	// @TODO: should we queue the health check into the healthcheck pool ?
	return getInjectorOrDefault(i).serviceHealthCheck(ctx, name)
}

// Shutdown stops a service, using type inference to determine the service name.
// This function performs a graceful shutdown on a service by inferring its name from the type T.
// The service must implement one of the Shutdowner interfaces.
//
// Parameters:
//   - i: The injector containing the service
//
// Returns an error if the shutdown fails, or nil if the shutdown was successful.
//
// Example:
//
//	err := do.Shutdown[*Database](injector)
//	if err != nil {
//	    log.Printf("Database shutdown failed: %v", err)
//	}
func Shutdown[T any](i Injector) error {
	name := inferServiceName[T]()
	return ShutdownNamedWithContext(context.Background(), i, name)
}

// ShutdownWithContext stops a service, using type inference to determine the service name.
// This function performs a graceful shutdown on a service with context support for timeout
// and cancellation. The service must implement one of the Shutdowner interfaces.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - i: The injector containing the service
//
// Returns an error if the shutdown fails, or nil if the shutdown was successful.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	err := do.ShutdownWithContext[*Database](ctx, injector)
//	if err != nil {
//	    log.Printf("Database shutdown failed: %v", err)
//	}
func ShutdownWithContext[T any](ctx context.Context, i Injector) error {
	name := inferServiceName[T]()
	return ShutdownNamedWithContext(ctx, i, name)
}

// ShutdownNamed stops a named service.
// This function performs a graceful shutdown on a service with the specified name.
// The service must implement one of the Shutdowner interfaces.
//
// Parameters:
//   - i: The injector containing the service
//   - name: The name of the service to shutdown
//
// Returns an error if the shutdown fails, or nil if the shutdown was successful.
//
// Example:
//
//	err := do.ShutdownNamed(injector, "main-database")
//	if err != nil {
//	    log.Printf("Main database shutdown failed: %v", err)
//	}
func ShutdownNamed(i Injector, name string) error {
	return ShutdownNamedWithContext(context.Background(), i, name)
}

// ShutdownNamedWithContext stops a named service with context support.
// This function performs a graceful shutdown on a service with the specified name and context
// support for timeout and cancellation. The service must implement one of the Shutdowner interfaces.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - i: The injector containing the service
//   - name: The name of the service to shutdown
//
// Returns an error if the shutdown fails, or nil if the shutdown was successful.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	err := do.ShutdownNamedWithContext(ctx, injector, "main-database")
//	if err != nil {
//	    log.Printf("Main database shutdown failed: %v", err)
//	}
func ShutdownNamedWithContext(ctx context.Context, i Injector, name string) error {
	return getInjectorOrDefault(i).serviceShutdown(ctx, name)
}

// MustShutdown stops a service, using type inference to determine the service name. It panics on error.
// This function performs a graceful shutdown on a service by inferring its name from the type T.
// If the shutdown fails, this function will panic.
//
// Parameters:
//   - i: The injector containing the service
//
// Panics if the shutdown fails.
//
// Example:
//
//	do.MustShutdown[*Database](injector)
func MustShutdown[T any](i Injector) {
	must0(Shutdown[T](i))
}

// MustShutdownWithContext stops a service, using type inference to determine the service name. It panics on error.
// This function performs a graceful shutdown on a service with context support for timeout
// and cancellation. If the shutdown fails, this function will panic.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - i: The injector containing the service
//
// Panics if the shutdown fails.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	do.MustShutdownWithContext[*Database](ctx, injector)
func MustShutdownWithContext[T any](ctx context.Context, i Injector) {
	must0(ShutdownWithContext[T](ctx, i))
}

// MustShutdownNamed stops a named service. It panics on error.
// This function performs a graceful shutdown on a service with the specified name.
// If the shutdown fails, this function will panic.
//
// Parameters:
//   - i: The injector containing the service
//   - name: The name of the service to shutdown
//
// Panics if the shutdown fails.
//
// Example:
//
//	do.MustShutdownNamed(injector, "main-database")
func MustShutdownNamed(i Injector, name string) {
	must0(ShutdownNamed(i, name))
}

// MustShutdownNamedWithContext stops a named service. It panics on error.
// This function performs a graceful shutdown on a service with the specified name and context
// support for timeout and cancellation. If the shutdown fails, this function will panic.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - i: The injector containing the service
//   - name: The name of the service to shutdown
//
// Panics if the shutdown fails.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	do.MustShutdownNamedWithContext(ctx, injector, "main-database")
func MustShutdownNamedWithContext(ctx context.Context, i Injector, name string) {
	must0(ShutdownNamedWithContext(ctx, i, name))
}
