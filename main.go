package main

import (
	"crypto/tls"
	"flag"
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

func getCertificates() (string, string, error) {
	keyPath := flag.String("key", "", "The path to the private key")
	certPath := flag.String("certificate", "", "The path to the certificate")
	flag.Parse()

	if *keyPath == "" || *certPath == "" {
		return "", "", fmt.Errorf("Either or both of -key or -cert not set")
	}

	return *keyPath, *certPath, nil
}

func tlsSetup(keyPath string, certPath string, listener net.Listener) (net.Listener, error) {
	config := &tls.Config{
		Certificates:             make([]tls.Certificate, 1),
		NextProtos:               []string{"h2", "http/1.1"},
		PreferServerCipherSuites: true,
	}

	var err error

	log.Printf("Loading certs from key: %s and cert: %s", keyPath, certPath)

	config.Certificates[0], err = tls.LoadX509KeyPair(
		certPath,
		keyPath,
	)

	if err != nil {
		log.Printf("Failed to configure TLS: %s", err)
		return nil, err
	}

	return tls.NewListener(listener, config), nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	listener, err := getListener()

	if err != nil {
		log.Fatalf("Could not set up listener: %s", err)
	}

	keyPath, certPath, err := getCertificates()

	if err != nil {
		log.Fatalf("Could not load certificates: %s", err)
	}

	tlsListener, err := tlsSetup(keyPath, certPath, listener)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "<h1>%s</h1>", getRandomLunch())
	})

	fmt.Printf("Starting web server on port %s", tlsListener.Addr().String())
	http.Serve(tlsListener, nil)
}