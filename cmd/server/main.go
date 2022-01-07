package main

import (
	"fmt"
	"net/http"
)

const HOST_IP = ""
const HOST_PORT = "8080"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})

	http.ListenAndServe(fmt.Sprintf("%s:%s", HOST_IP, HOST_PORT), nil)
}
