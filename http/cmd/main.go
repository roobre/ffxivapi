package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"roob.re/ffxivapi"
	ffxivapihttp "roob.re/ffxivapi/http"
	"roob.re/ffxivapi/lodestone"
	"roob.re/tcache"
	"syscall"
	"time"
)

func main() {
	addr := ":8080"
	if len(os.Args) >= 2 {
		addr = os.Args[1]
	}
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	region := "eu"
	if envRegion := os.Getenv("FFXIVAPI_REGION"); envRegion != "" {
		region = envRegion
	}

	api := ffxivapi.New()

	// If FFXIVAPI_NOCACHE does not exist (== "")
	if os.Getenv("FFXIVAPI_NOCACHE") == "" {
		api.Lodestone = &lodestone.HTTPClient{
			Region: region,
			HTTPClient: &http.Client{
				Transport: &lodestone.TCacheRoundTripper{
					RoundTripper: http.DefaultTransport,
					Cache:        tcache.New(tcache.NewMemStorage()),
					MaxAge:       15 * time.Minute,
				},
			},
		}
	}

	h := ffxivapihttp.NewWithApi(api)

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
