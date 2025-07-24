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

func newShutdownErrors() *ShutdownErrors {
	return &ShutdownErrors{}
}

type ShutdownErrors map[EdgeService]error

func (e *ShutdownErrors) Add(scopeID string, scopeName string, serviceName string, err error) {
	if err != nil {
		(*e)[newEdgeService(scopeID, scopeName, serviceName)] = err
	}
}

func (e ShutdownErrors) Len() int {
	out := 0
	for _, v := range e {
		if v != nil {
			out++
		}
	}
	return out
}

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
		if in != nil {
			se := &ShutdownErrors{}
			if ok := errors.As(in, &se); ok {
				for k, v := range *se {
					(*out)[k] = v
				}
			}
		}
	}

	return out
}
