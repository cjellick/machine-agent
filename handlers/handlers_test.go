package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/cjellick/machine-agent/events"
	"github.com/cjellick/machine-agent/test_utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

const eventServerPort string = "8006"
const baseUrl string = "http://localhost:" + eventServerPort
const pushUrl string = baseUrl + "/pushEvent"
const subscribeUrl string = baseUrl + "/subscribe"

func TestSanity(t *testing.T) {
	log.Println("Handler sanity test passed")
}

func setUp() *events.EventRouter {
	eventHandlers := map[string]events.EventHandler{
		"physicalhost.create": CreateMachine,
		//"post.physicalhost.activate": ActivateMachine,
		//"post.physicalhost.create":   CreateMachine,
		//"post.physicalhost.activate": ActivateMachine,
	}

	router := events.NewEventRouter("dockerMachineAgent", 2000, baseUrl, eventHandlers, 2)
	ready := make(chan bool, 1)
	go router.Start(ready)
	// Wait for start to be ready
	<-ready

	return router
}

func TestHandler(t *testing.T) {
	setUp()
	err := prepAndPostEvent("../test_utils/resources/create_virtualbox.json")
	checkError(err, t)
	time.Sleep(500 * time.Millisecond)
}

func checkError(err error, t *testing.T) {
	if err != nil {
		log.Fatal(err)
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
	ready := make(chan string, 1)
	go test_utils.InitializeServer(eventServerPort, ready)
	<-ready
	result := m.Run()
	// TODO Kill event server
	os.Exit(result)
}
