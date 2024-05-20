package main

import "github.com/maksadbek/statsfornba/cmd/consumer/app"

func main() {
	err := app.Run()
	if err != nil {
		panic(err)
	}
}
