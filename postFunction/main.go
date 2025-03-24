package main

import (
	"fmt"
	"net/http"
)

func main() {
	port := "80"

	http.HandleFunc("/", EntryPoint)
	fmt.Printf("Listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("ListenAndServe: %v\n", err)
	}
}
