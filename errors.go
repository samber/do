package do

import "errors"

var ErrServiceNotFound = errors.New("DI: could not find service")
var ErrCircularDependency = errors.New("DI: circular dependency detected")
var ErrHealthCheckTimeout = errors.New("DI: health check timeout")
