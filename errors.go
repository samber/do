package do

import (
	"errors"
	"fmt"
	"strings"
)

var ErrServiceNotFound = errors.New("DI: could not find service")
var ErrServiceNotMatch = errors.New("DI: could not find service satisfying interface")
var ErrCircularDependency = errors.New("DI: circular dependency detected")
var ErrHealthCheckTimeout = errors.New("DI: health check timeout")

// newShutdownErrors creates a new ShutdownErrors instance for collecting shutdown errors.
// This function initializes an empty map to store errors that occur during service shutdown.
//
// Returns a new ShutdownErrors instance ready for error collection.
func newShutdownErrors() *ShutdownErrors {
	return &ShutdownErrors{}
}

// ShutdownErrors is a map that collects errors that occur during service shutdown.
// This type is used to aggregate multiple shutdown errors from different services
// and provide a comprehensive error report.
//
// The map key is an EdgeService that uniquely identifies the service that failed to shutdown,
// and the value is the error that occurred during shutdown.
type ShutdownErrors map[EdgeService]error

// Add adds an error to the ShutdownErrors collection for a specific service.
// This method is used to record errors that occur during service shutdown.
// If the error is nil, no entry is added to the collection.
//
// Parameters:
//   - scopeID: The unique identifier of the scope containing the service
//   - scopeName: The human-readable name of the scope containing the service
//   - serviceName: The name of the service that failed to shutdown
//   - err: The error that occurred during shutdown (nil errors are ignored)
func (e *ShutdownErrors) Add(scopeID string, scopeName string, serviceName string, err error) {
	if err != nil {
		(*e)[newEdgeService(scopeID, scopeName, serviceName)] = err
	}
}

// Len returns the number of non-nil errors in the ShutdownErrors collection.
// This method provides a count of actual shutdown failures.
//
// Returns the number of services that failed to shutdown properly.
func (e ShutdownErrors) Len() int {
	out := 0
	for _, v := range e {
		if v != nil {
			out++
		}
	}
	return out
}

// Error returns a formatted string representation of all shutdown errors.
// This method implements the error interface and provides a human-readable
// summary of all shutdown failures.
//
// Returns a formatted string containing all shutdown errors, or a message
// indicating no errors if the collection is empty.
//
// Example output:
//
//	"DI: shutdown errors:
//	  - root > database: connection refused
//	  - api > logger: failed to flush logs"
func (e ShutdownErrors) Error() string {
	lines := []string{}
	for k, v := range e {
		if v != nil {
			lines = append(lines, fmt.Sprintf("  - %s > %s: %s", k.ScopeName, k.Service, v.Error()))
		}
	}

	if len(lines) == 0 {
		return "DI: no shutdown errors"
	}

	return "DI: shutdown errors:\n" + strings.Join(lines, "\n")
}

func mergeShutdownErrors(ins ...*ShutdownErrors) *ShutdownErrors {
	out := newShutdownErrors()

	for _, in := range ins {
		if in == nil {
			continue
		}

		for k, v := range *in {
			(*out)[k] = v
		}
	}

	return out
}
