package main

import (
	// "fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	// "time"
	// "strconv"
)

var messages chan string = make(chan string, 100)

func PushHandler(w http.ResponseWriter, req *http.Request) {

	bod, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	body := string(bod[:len(bod)])
	messages <- body
}

func doResponse(w http.ResponseWriter) {
	io.WriteString(w, <-messages+"\r\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	doResponse(w)
}
func PollResponse(w http.ResponseWriter, req *http.Request) {
	doResponse(w)
	/*
		io.WriteString(w, <-messages)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		time.Sleep(2 * time.Second)
		fmt.Fprintf(w, "\n\rbye2\r\n")
	*/

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

/*
package main

import (
	"fmt"
	//	"io/ioutil"
	//	"log"
	"net/http"
	"time"
)

func handler(response http.ResponseWriter, r *http.Request) {
	This works for flushing and sleeping
	fmt.Fprintf(response, "YOOOOOO\r\n")
	if f, ok := response.(http.Flusher); ok {
		f.Flush()
	}
	time.Sleep(1 * time.Second)
	fmt.Fprintf(response, "hey\r\n")
	if f, ok := response.(http.Flusher); ok {
		f.Flush()
	}
	time.Sleep(2 * time.Second)
	fmt.Fprintf(response, "bye2\r\n")
}

func main() {
	lpchan := make(chan chan string)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8888", nil)
}
*/
