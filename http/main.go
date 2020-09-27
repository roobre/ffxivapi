package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	addr := ":8080"
	if len(os.Args) >= 2 {
		addr = os.Args[1]
	}

	h := NewHTTPApi()
	log.Println("Listening on " + addr)
	log.Fatal(http.ListenAndServe(addr, h))
}

