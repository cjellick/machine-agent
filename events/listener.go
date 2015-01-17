package events

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/cjellick/machine-agent/locks"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type EventHandler func(*Event, string)

type EventRouter struct {
	name                string
	priority            int
	registerUrl         string
	subscribeUrl        string
	replyUrl            string
	eventHandlers       map[string]EventHandler
	workerCount         int
	eventStreamResponse *http.Response
}

func (router *EventRouter) Start(ready chan<- bool) (err error) {
	workers := make(chan *Worker, router.workerCount)
	for i := 0; i < router.workerCount; i++ {
		w := newWorker(router.replyUrl)
		workers <- w
	}
	registerForm := url.Values{}
	subscribeForm := url.Values{}
	// TODO Understand need/function of UUID
	registerForm.Set("uuid", router.name)
	registerForm.Set("name", router.name)
	registerForm.Set("priority", strconv.Itoa(router.priority))

	eventHandlerSuffix := ";handler=" + router.name
	handlers := map[string]EventHandler{}

	if pingHandler, ok := router.eventHandlers["ping"]; ok {
		// Ping doesnt need registered in the POST and
		// ping events don't have the handler suffix. If we
		// start handling other non-suffix events,
		// we might consider improving this.
		handlers["ping"] = pingHandler
	}

	for event, handler := range router.eventHandlers {
		registerForm.Add("processNames", event)
		fullEventKey := event + eventHandlerSuffix
		subscribeForm.Add("eventNames", fullEventKey)
		handlers[fullEventKey] = handler
	}

	regResponse, err := http.PostForm(router.registerUrl, registerForm)
	if err != nil {
		return err
	}
	defer regResponse.Body.Close()

	if ready != nil {
		ready <- true
	}

	// TODO Harden. Add reconnect logic.
	eventStream, err := http.PostForm(router.subscribeUrl, subscribeForm)
	if err != nil {
		return err
	}
	log.Println("Connection established.")
	router.eventStreamResponse = eventStream
	defer eventStream.Body.Close()

	scanner := bufio.NewScanner(eventStream.Body)
	for scanner.Scan() {
		line := scanner.Bytes()

		// TODO Ensure this wont break eventing paradigm
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		select {
		case worker := <-workers:
			go worker.DoWork(line, router.replyUrl, handlers, workers)
		default:
			log.Printf("No workers available dropping event.")
		}
	}

	return nil
}

func (router *EventRouter) Stop() (err error) {
	router.eventStreamResponse.Body.Close()
	return nil
}

type Worker struct {
}

func (w *Worker) DoWork(rawEvent []byte, replyUrl string, eventHandlers map[string]EventHandler,
	workers chan *Worker) {
	defer func() { workers <- w }()

	event := &Event{}
	err := json.Unmarshal(rawEvent, &event)
	if err != nil {
		log.Printf("Error unmarshalling event: %v", err)
		return
	}

	log.Printf("Received event: %v", event.Name)
	unlocker := locks.Lock(event.ResourceId)
	if unlocker == nil {
		log.Printf("Resouce [%v] locked. Dropping event.", event.ResourceId)
		return
	}
	defer unlocker.Unlock()

	if fn, ok := eventHandlers[event.Name]; ok {
		fn(event, replyUrl)
	} else {
		log.Printf("No handler registered for event %v", event.Name)
	}
}

func NewEventRouter(name string, priority int, baseUrl string,
	eventHandlers map[string]EventHandler, workerCount int) *EventRouter {
	return &EventRouter{
		name:          name,
		priority:      priority,
		registerUrl:   baseUrl + "/externalhandlers",
		subscribeUrl:  baseUrl + "/subscribe",
		replyUrl:      baseUrl + "/publish",
		eventHandlers: eventHandlers,
		workerCount:   workerCount,
	}
}

func newWorker(replyUrl string) *Worker {
	return &Worker{}
}
