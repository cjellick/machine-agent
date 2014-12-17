package test_utils

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var subscriberChannels []chan string

func pushEventHandler(w http.ResponseWriter, req *http.Request) {
	bod, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	body := string(bod[:len(bod)])
	pushToSubscribers(body)
}

func subscribeHandler(w http.ResponseWriter, req *http.Request) {
	resultChan := make(chan string)
	subscriberChannels = append(subscriberChannels, resultChan)
	writeEventToSubscriber(w, resultChan)
}

func pushToSubscribers(message string) {
	for i := range subscriberChannels {
		subscriberChannels[i] <- message
	}
}

func writeEventToSubscriber(w http.ResponseWriter, c chan string) {
	for {
		io.WriteString(w, <-c+"\r\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func InitializeServer(port string) {
	http.HandleFunc("/subscribe", subscribeHandler)
	http.HandleFunc("/pushEvent", pushEventHandler)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Failed to start event server: ", err)
	}
}
