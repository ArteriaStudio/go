package main

import (
	"fmt"
	"net/http"

	"postFunction"
)

func main() {
	port := "80"

	http.HandleFunc("/", postFunction.EntryPoint)
	fmt.Printf("Listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("ListenAndServe: %v\n", err)
	}
}
