package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Poll struct {
	resultChan chan string
	handler    func(w http.ResponseWriter, c chan string)
	w          http.ResponseWriter
}

var polls []Poll

func PushHandler(w http.ResponseWriter, req *http.Request) {

	bod, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	body := string(bod[:len(bod)])
	pushIt(body)
}

func pushIt(message string) {
	for i := range polls {
		polls[i].resultChan <- message
	}
}

func doResp(w http.ResponseWriter, c chan string) {
	for {
		io.WriteString(w, <-c+"\r\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func PollResponse(w http.ResponseWriter, req *http.Request) {
	poll := Poll{make(chan string), doResp, w}
	polls = append(polls, poll)
	poll.handler(w, poll.resultChan)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/poll", PollResponse)
	http.HandleFunc("/push", PushHandler)
	err := http.ListenAndServe(":8005", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
