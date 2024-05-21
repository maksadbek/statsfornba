package main

import (
	"log"

	"github.com/maksadbek/statsfornba/cmd/consumer/app"
)

func main() {
	err := app.Run()
	if err != nil {
		log.Fatal("failed to start the application", err)
	}
}
