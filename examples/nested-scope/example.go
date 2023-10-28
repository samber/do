package main

import (
	"fmt"

	"github.com/samber/do/v2"
)

/**
 * Run example
 */

func main() {
	injector := do.New()

	driverModule := injector.Scope("driver")
	passengerModule := injector.Scope("passenger")

	bootCarModule(injector)
	bootDriverModule(driverModule)
	bootPassengerModule(passengerModule)

	// 1 driver + 1 passenger
	driver := do.MustInvoke[*Driver](driverModule)
	passenger := do.MustInvokeNamed[*Passenger](passengerModule, "passenger-1")

	driver.TakeASeat()
	passenger.TakeASeat()

	fmt.Println(injector.ShutdownOnSIGTERMOrInterrupt())
}
