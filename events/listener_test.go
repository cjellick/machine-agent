package events

import (
	"fmt"
	"github.com/cjellick/machine-agent/test_utils"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const eventServerPort string = "8005"
const baseUrl string = "http://localhost:" + eventServerPort
const pushUrl string = baseUrl + "/pushEvent"
const subscribeUrl string = baseUrl + "/subscribe"

func TestSanity(t *testing.T) {
	fmt.Println("Test passed")
}

func TestConnectivity(t *testing.T) {
	router := NewEventRouter(subscribeUrl)

	eventsReceived := make(chan *Event)
	testHandler := func(event *Event) {
		eventsReceived <- event
	}

	RegisterEventHandler("event.test", testHandler)
	go router.Start()

	// Unfortunate hack to allow router to start listening before sending first event
	time.Sleep(300 * time.Millisecond)
	reader := strings.NewReader("{\"name\": \"event.test\", \"sleep\": 9, \"data\": \"first\"}")
	http.Post(pushUrl, "application/json", reader)

	recievedEvent := <-eventsReceived
	if recievedEvent.Name != "event.test" || recievedEvent.Data != "first" {
		t.Fail()
	}

	reader = strings.NewReader("{\"name\": \"event.test\", \"sleep\": 1, \"data\": \"second\"}")
	http.Post(pushUrl, "application/json", reader)
	recievedEvent = <-eventsReceived
	if recievedEvent.Name != "event.test" || recievedEvent.Data != "second" {
		t.Fail()
	}

}

func TestMain(m *testing.M) {
	go test_utils.InitializeServer(eventServerPort)
	result := m.Run()
	// TODO Kill event server
	os.Exit(result)
}
