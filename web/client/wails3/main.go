package main

import (
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	app := NewApp()

	w := application.New(application.Options{
		Name:        "PocketPaw",
		Description: "PocketPaw desktop client (Wails 3 migration)",
		Services: []application.Service{
			application.NewService(app),
		},
	})

	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
