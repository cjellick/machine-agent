package events

import (
	"bytes"
	"fmt"
	"github.com/cjellick/machine-agent/test_utils"
	"net/http"
	"os"
	//	"strings"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"
)

const eventServerPort string = "8005"
const baseUrl string = "http://localhost:" + eventServerPort
const pushUrl string = baseUrl + "/pushEvent"
const subscribeUrl string = baseUrl + "/subscribe"

func TestSanity(t *testing.T) {
	fmt.Println("Sanity test passed")
}

func TestRouting(t *testing.T) {
	router := NewEventRouter(subscribeUrl)

	eventsReceived := make(chan *Event)
	testHandler := func(event *Event) {
		eventsReceived <- event
	}

	RegisterEventHandler("physicalhost.activate;handler=demo", testHandler)
	go router.Start()

	// Unfortunate hack to allow router to start listening before sending first event
	time.Sleep(300 * time.Millisecond)

	err := prepAndPostEvent("../test_utils/resources/create_virtualbox.json")
	checkError(err, t)

	recievedEvent := <-eventsReceived
	if recievedEvent.Name != "physicalhost.activate;handler=demo" || recievedEvent.Data["driver"] != "virtualbox" {
		t.Fail()
	}

	err = prepAndPostEvent("../test_utils/resources/create_virtualbox.json")
	recievedEvent = <-eventsReceived
	if recievedEvent.Name != "physicalhost.activate;handler=demo" || recievedEvent.Data["driver"] != "virtualbox" {
		t.Fail()
	}
}

func checkError(err error, t *testing.T) {
	if err != nil {
		t.Error()
	}
}
func prepAndPostEvent(eventFile string) (err error) {
	rawEvent, err := ioutil.ReadFile(eventFile)
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
	go test_utils.InitializeServer(eventServerPort)
	result := m.Run()
	// TODO Kill event server
	os.Exit(result)
}
