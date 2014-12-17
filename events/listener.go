package events

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
)

type EventHandler func(*Event)

var handlerMap = make(map[string]EventHandler)

func RegisterEventHandler(eventName string, handler func(*Event)) {
	handlerMap[eventName] = handler
}

type EventRouter struct {
	url string
}

func (client *EventRouter) Start() {
	// TODO Harden. Add reconnect logic.
	resp, err := http.Get(client.url)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		routeEvent(line)
	}
}

func routeEvent(rawEvent []byte) {
	event := &Event{}
	err := json.Unmarshal(rawEvent, &event)
	if err != nil {
		log.Fatal(err)
	}
	if fn, ok := handlerMap[event.Name]; ok {
		log.Printf("Routing event %v", event.Name)
		// TODO Refactor to a worker model so that system isn't overwhelmed with goroutines
		go fn(event)
	} else {
		log.Printf("No handler registered for event %v", event.Name)
	}
}

func NewEventRouter(url string) *EventRouter {
	return &EventRouter{
		url: url,
	}
}
