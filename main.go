package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
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

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "<h1>%s</h1>", getRandomLunch())
	})

	fmt.Println("Starting web server on port 8080")
	http.ListenAndServe(":8080", nil)
}