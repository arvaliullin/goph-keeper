// Package main запускает CLI-клиент GophKeeper.
package main

import (
	"github.com/arvaliullin/goph-keeper/internal/client/app"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	app.SetBuildInfo(buildVersion, buildDate)
	app.Execute()
}
