package helpers

import "github.com/PushAndRun/bookings/internal/config"

var app *config.AppConfig

func NewHelpers(a *config.AppConfig) {
	app = a
}
