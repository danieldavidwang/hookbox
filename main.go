package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", handleRequest)

	fmt.Println("Hookbox listening on http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("server error:", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request!")
	fmt.Println("Method:", r.Method)
	fmt.Println("Path:", r.URL.Path)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received\n"))
}
