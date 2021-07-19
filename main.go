package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sync"
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

func getDefaultListeners() (map[string][]net.Listener, error) {

	listener, err := net.Listen("tcp", ":8443")

	if err != nil {
		return nil, err
	}

	return map[string][]net.Listener{
		"https": {listener},
	}, nil
}

func getListeners() (map[string][]net.Listener, error) {
	listeners, err := activation.ListenersWithNames()

	if err != nil || len(listeners) == 0 {
		log.Printf("Received no listeners from socket activation, defaulting to HTTPS on port 8443")
		listeners, err = getDefaultListeners()
	}

	return listeners, err
}

func getCertificatePaths() (string, string, error) {
	keyPath := flag.String("key", "", "The path to the private key")
	certPath := flag.String("certificate", "", "The path to the certificate")
	flag.Parse()

	if *keyPath == "" || *certPath == "" {
		return "", "", errors.New("Either or both of -key or -certificate not set")
	}

	return *keyPath, *certPath, nil
}

func getTLSConfig(keyPath string, certPath string) (*tls.Config, error) {
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

	return config, nil
}

func createWebServers() (*http.ServeMux, *http.ServeMux) {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("HTTP request from %s\n", req.RemoteAddr)
		hostname := fmt.Sprintf("https://%s", req.Host)
		http.Redirect(w, req, hostname+req.RequestURI, http.StatusMovedPermanently)
	})

	httpsMux := http.NewServeMux()
	httpsMux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("HTTPS request from %s\n", req.RemoteAddr)
		fmt.Fprintf(w, "<h1>%s</h1>", getRandomLunch())
	})

	return httpMux, httpsMux
}

func main() {
	rand.Seed(time.Now().UnixNano())

	log.Printf("LISTEN_FDS is %s\n", os.Getenv("LISTEN_FDS"))
	log.Printf("LISTEN_FDNAMES is %s\n", os.Getenv("LISTEN_FDNAMES"))

	listeners, err := getListeners()
	httpMux, httpsMux := createWebServers()

	if err != nil {
		log.Fatalf("Could not set up listeners: %s", err)
	}

	var wg sync.WaitGroup

	if tlsListeners, ok := listeners["https"]; ok {
		keyPath, certPath, err := getCertificatePaths()

		if err != nil {
			log.Fatalf("Could not load certificates: %s", err)
		}

		tlsConfig, err := getTLSConfig(keyPath, certPath)

		for i := range tlsListeners {
			wg.Add(1)

			go func(l net.Listener) {
				log.Printf("Starting secure web server on port %s\n", l.Addr())
				tl := tls.NewListener(l, tlsConfig)
				log.Fatal(http.Serve(tl, httpsMux))
			}(tlsListeners[i])
		}
	}

	if plainListeners, ok := listeners["http"]; ok {
		for i := range plainListeners {
			wg.Add(1)

			go func(l net.Listener) {
				log.Printf("Starting plaintext web server on port %s\n", l.Addr())
				log.Fatal(http.Serve(l, httpMux))
			}(plainListeners[i])
		}
	}

	wg.Wait()
}
