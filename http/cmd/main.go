package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	ffxivapihttp "roob.re/ffxivapi/http"
	"syscall"
)

func main() {
	addr := ":8080"
	if len(os.Args) >= 2 {
		addr = os.Args[1]
	}
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	h := ffxivapihttp.NewHTTPApi()
	s := &http.Server{
		Addr:    addr,
		Handler: h,
	}

	go func() {
		log.Println("Listening in " + addr + "...")
		err := s.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	signal := <-sigChan
	log.Printf("Caught %s, shutting down...", signal.String())

	err := s.Shutdown(context.Background())
	if err != nil {
		log.Println(err)
	}
}
