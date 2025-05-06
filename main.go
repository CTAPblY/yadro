package main

import (
	"fmt"

	"github.com/CTAPblY/yadro/handler"
	"github.com/CTAPblY/yadro/loader"
)

func main() {
	config, err := loader.LoadConfig("config.json")

	if err != nil {
		fmt.Println(err)
	}
	events, err := loader.LoadEvents("events")
	if err != nil {
		fmt.Println(err)
	}
	report, err := handler.HandleEvents(config, events)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(report)
}
