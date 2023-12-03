package do

import "time"

type InjectorOpts struct {
	HookAfterRegistration func(scope *Scope, serviceName string)
	HookAfterShutdown     func(scope *Scope, serviceName string)

	Logf func(format string, args ...any)

	HealthCheckParallelism   uint          // default: all jobs are executed in parallel
	HealthCheckGlobalTimeout time.Duration // default: no timeout
	HealthCheckTimeout       time.Duration // default: no timeout
}
