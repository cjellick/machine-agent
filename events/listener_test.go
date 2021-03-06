package events

import (
	"bytes"
	"encoding/json"
	"github.com/cjellick/machine-agent/test_utils"
	"io/ioutil"
	//"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

const eventServerPort string = "8005"
const baseUrl string = "http://localhost:" + eventServerPort
const pushUrl string = baseUrl + "/pushEvent"
const subscribeUrl string = baseUrl + "/subscribe"

// Tests the simplest case of successfully receiving, routing, and handling
// three events.
func TestSimpleRouting(t *testing.T) {
	eventsReceived := make(chan *Event)
	testHandler := func(event *Event, replyUrl string) {
		eventsReceived <- event
	}

	eventHandlers := map[string]EventHandler{"physicalhost.create": testHandler}
	router := NewEventRouter("testRouter", 2000, baseUrl, eventHandlers, 3)
	ready := make(chan bool, 1)
	go router.Start(ready)
	// Wait for start to be ready
	<-ready

	preCount := 0
	pre := func(event *Event) {
		event.Id = strconv.Itoa(preCount)
		event.ResourceId = strconv.FormatInt(time.Now().UnixNano(), 10)
		preCount += 1
		event.Name = "physicalhost.create;handler=testRouter"
	}

	// Push 3 events
	for i := 0; i < 3; i++ {
		err := prepAndPostEvent("../test_utils/resources/create_virtualbox.json", pre)
		checkError(err, t)
	}
	receivedEvents := map[string]*Event{}
	for i := 0; i < 3; i++ {
		receivedEvent := awaitEvent(eventsReceived, 100, t)
		if receivedEvent != nil {
			receivedEvents[receivedEvent.Id] = receivedEvent
		}
	}

	for i := 0; i < 3; i++ {
		if _, ok := receivedEvents[strconv.Itoa(i)]; !ok {
			t.Errorf("Didn't get event %v", i)
		}
	}

	router.Stop()
}

// If no workers are available (because they're all busy), an event should simply be dropped.
// This tests that functionality
func TestEventDropping(t *testing.T) {
	eventsReceived := make(chan *Event)
	stopWaiting := make(chan bool)
	testHandler := func(event *Event, replyUrl string) {
		eventsReceived <- event
		<-stopWaiting
	}

	eventHandlers := map[string]EventHandler{"physicalhost.create": testHandler}

	// 2 workers, not 3, means the last event should be droppped
	router := NewEventRouter("testRouter", 2000, baseUrl, eventHandlers, 2)
	ready := make(chan bool, 1)
	go router.Start(ready)
	// Wait for start to be ready
	<-ready

	preCount := 0
	pre := func(event *Event) {
		event.Id = strconv.Itoa(preCount)
		event.ResourceId = strconv.FormatInt(time.Now().UnixNano(), 10)
		preCount += 1
		event.Name = "physicalhost.create;handler=testRouter"
	}

	// Push 3 events
	for i := 0; i < 3; i++ {
		err := prepAndPostEvent("../test_utils/resources/create_virtualbox.json", pre)
		checkError(err, t)
	}
	receivedEvents := map[string]*Event{}
	for i := 0; i < 3; i++ {
		receivedEvent := awaitEvent(eventsReceived, 20, t)
		if receivedEvent != nil {
			receivedEvents[receivedEvent.Id] = receivedEvent
		}
	}

	if len(receivedEvents) != 2 {
		t.Errorf("Unexpected length %v", len(receivedEvents))
	}
	router.Stop()
}

// Tests that when we have more events than workers, workers are added back to the pool
// when they are done doing their work and capable of handling more work.
func TestWorkerReuse(t *testing.T) {
	eventsReceived := make(chan *Event)
	testHandler := func(event *Event, replyUrl string) {
		time.Sleep(10 * time.Millisecond)
		eventsReceived <- event
	}

	eventHandlers := map[string]EventHandler{"physicalhost.create": testHandler}

	router := NewEventRouter("testRouter", 2000, baseUrl, eventHandlers, 1)
	ready := make(chan bool, 1)
	go router.Start(ready)
	// Wait for start to be ready
	<-ready
	preCount := 1
	pre := func(event *Event) {
		event.Id = strconv.Itoa(preCount)
		event.ResourceId = strconv.FormatInt(time.Now().UnixNano(), 10)
		preCount += 1
		event.Name = "physicalhost.create;handler=testRouter"
	}

	// Push 3 events
	receivedEvents := map[string]*Event{}
	for i := 0; i < 2; i++ {
		err := prepAndPostEvent("../test_utils/resources/create_virtualbox.json", pre)
		checkError(err, t)
		receivedEvent := awaitEvent(eventsReceived, 200, t)
		if receivedEvent != nil {
			receivedEvents[receivedEvent.Id] = receivedEvent
		}
	}

	if len(receivedEvents) != 2 {
		t.Errorf("Unexpected length %v", len(receivedEvents))
	}
}

func awaitEvent(eventsReceived chan *Event, millisToWait int, t *testing.T) *Event {
	timeout := make(chan bool, 1)
	timeoutFunc := func() {
		time.Sleep(time.Duration(millisToWait) * time.Millisecond)
		timeout <- true
	}
	go timeoutFunc()

	select {
	case e := <-eventsReceived:
		return e
	case <-timeout:
		return nil
	}
	return nil
}

func checkError(err error, t *testing.T) {
	if err != nil {
		t.Error()
	}
}

type PreFunc func(*Event)

func prepAndPostEvent(eventFile string, preFunc PreFunc) (err error) {
	rawEvent, err := ioutil.ReadFile(eventFile)
	if err != nil {
		return err
	}

	event := &Event{}
	err = json.Unmarshal(rawEvent, &event)
	if err != nil {
		return err
	}
	preFunc(event)
	rawEvent, err = json.Marshal(event)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = json.Compact(buffer, rawEvent)
	if err != nil {
		return err
	}
	http.Post(pushUrl, "application/json", buffer)

	return nil
}

func TestMain(m *testing.M) {
	ready := make(chan string, 1)
	go test_utils.InitializeServer(eventServerPort, ready)
	<-ready
	result := m.Run()
	// TODO Kill event server
	os.Exit(result)
}
