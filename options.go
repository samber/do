package do

type InjectorOpts struct {
	HookAfterRegistration func(scope *Scope, serviceName string)
	HookAfterShutdown     func(scope *Scope, serviceName string)

	Logf func(format string, args ...any)
}
