package main

import (
	"log"

	"github.com/maksadbek/statsfornba/cmd/api/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatal("failed to start the application", err)
	}
}
