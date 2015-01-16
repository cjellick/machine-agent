package main

import (
	"github.com/cjellick/machine-agent/events"
	"github.com/cjellick/machine-agent/handlers"
	_ "log"
)

func main() {
	// TODO Implement!

	eventHandlers := map[string]events.EventHandler{
		"physicalhost.create": handlers.CreateMachine,
		"ping":                handlers.Ping}

	router := events.NewEventRouter("goMachineService", 2000, "http://localhost:8080/v1", eventHandlers, 3)
	router.Start(nil)
	// TODO What cleanup/teardown needs done here? killing router?
}
