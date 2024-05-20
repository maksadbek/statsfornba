package main

import "github.com/maksadbek/statsfornba/cmd/api/app"

func main() {
	err := app.Run()
	if err != nil {
		panic(err)
	}
}
