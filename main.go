package main

import (
	"fmt"
	"io"
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
	fmt.Println("Query:", r.URL.Query())
	fmt.Println("Headers:", r.Header)

	// Limit request bodies to 1 MB.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not read request body", http.StatusBadRequest)
		return
	}

	fmt.Println("Body:", string(body))

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
