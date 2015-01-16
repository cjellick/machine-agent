package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cjellick/machine-agent/events"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func CreateMachine(event *events.Event, replyUrl string) {
	randomName := event.Data["name"].(string) + strconv.FormatInt(time.Now().Unix(), 10)
	log.Printf("Beginning create machine event %v %s", event, randomName)

	//cmd := exec.Command("/Users/cjellick/go/src/github.com/docker/machine/machine", "create", "-d", "virtualbox", randomName)
	cmd := exec.Command("/Users/cjellick/go/src/github.com/docker/machine/machine", "ls")
	r, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	go func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Printf("%s \n", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "There was an error with the scanner reading from exec cmd", err)
		}
	}(r)

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Done create machine.")

	log.Printf("Replying to %s", event.ReplyTo)
	replyEvent := events.NewReplyEvent()
	replyEvent.Name = event.ReplyTo
	replyEvent.PreviousIds = []string{event.Id}

	replyEventJson, err := json.Marshal(replyEvent)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sending: ", string(replyEventJson[:]))
	eventBuffer := bytes.NewBuffer(replyEventJson)
	replyRequest, err := http.NewRequest("POST", replyUrl, eventBuffer)
	replyRequest.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(replyRequest)
	if err != nil {
		log.Fatal("Error sending reply!", err)
	}
	log.Printf("Response code: %s", response.Status)
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	log.Printf("Response body: [%v]", string(body[:]))
}

func ActivateMachine(event *events.Event, replyUrl string) {
	log.Printf("Activating machine [%s]", event.Data["name"].(string))
}

func Ping(event *events.Event, replyUrl string) {
	// No-op ping handler
}
