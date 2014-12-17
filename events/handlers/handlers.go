package handlers

import (
	"github.com/cjellick/machine-agent/events"
	"log"
	"os/exec"
	"strconv"
	"time"
)

func init() {
	events.RegisterEventHandler("event.gosleep", goSleeper)
}

func goSleeper(event *events.Event) {
	time.Sleep(time.Duration(event.Sleep) * time.Second)
	log.Printf("Done sleeping for %v seconds", event.Sleep)
}

func execSleeper(event *events.Event) {
	cmd := exec.Command("sleep", strconv.Itoa(event.Sleep))

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("waiting for sleep command to finish.")
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Done sleeping.")

}
