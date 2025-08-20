package main

import (
	"fmt"
)

// Application represents the main application
type Application struct {
	Config *Configuration
	DB     *Database
	Logger *Logger
}

func (app *Application) Start() {
	app.Logger.Log("Application starting...")
	app.DB.Connect()
	app.Logger.Log(fmt.Sprintf("Server listening on port %d", app.Config.Port))
}
