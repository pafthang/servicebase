package main

import (
	"log"

	"github.com/pafthang/servicebase"
	appinit "github.com/pafthang/servicebase/app"
)

func main() {
	app := servicebase.New()

	appinit.RegisterTinybaseDefaults(app, app.RootCmd, appinit.TinybaseDefaultsConfig{})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
