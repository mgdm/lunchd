package main

import (
        "fmt"
        "log"
        "math/rand"
        "net"
        "net/http"
        "time"

        "github.com/coreos/go-systemd/activation"
)

var lunchOptions = []string{
        "Sandwich",
        "Soup",
        "Salad",
        "Burger",
        "Sushi",
}

func getRandomLunch() string {
        return lunchOptions[rand.Intn(len(lunchOptions))]
}

func getListener() (net.Listener, error) {
	listeners, err := activation.Listeners()

	if err != nil || len(listeners) != 1 {
		log.Printf("Excpected one listener, got %d: %s", len(listeners), err)

		listener, err := net.Listen("tcp", ":8080")
		return listener, err
	}

	return listeners[0], err
}

func main() {
        rand.Seed(time.Now().UnixNano())

        listener, err := getListener()

        if err != nil {
                log.Panicf("Could not set up listener: %s", err)
        }

        http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
                fmt.Fprintf(w, "<h1>%s</h1>", getRandomLunch())
        })

        log.Printf("Starting web server on port %s", listener.Addr().String())
        http.Serve(listener, nil)
}