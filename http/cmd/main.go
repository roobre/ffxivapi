package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"roob.re/ffxivapi"
	ffxivapihttp "roob.re/ffxivapi/http"
	"roob.re/ffxivapi/lodestone"
	"roob.re/tcache"
	"strings"
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

	loglvl := log.InfoLevel
	if envLevel := os.Getenv("FFXIVAPI_LOGLVL"); envLevel != "" {
		loglvl, _ = log.ParseLevel(envLevel)
	}
	log.SetLevel(loglvl)

	region := "eu"
	if envRegion := os.Getenv("FFXIVAPI_REGION"); envRegion != "" {
		region = envRegion
	}
	log.Infof("Using region %s", region)

	server := lodestone.CanonServerFromRegion(region)
	if envServer := os.Getenv("FFXIVAPI_SERVER"); envServer != "" {
		if !strings.HasPrefix(envServer, "http") {
			log.Fatal("FFXIVAPI_SERVER must start with http")
		}

		server = envServer
	}
	log.Infof("Using lodestone server %s", server)

	// If FFXIVAPI_NOCACHE does not exist (== "")
	var client = http.DefaultClient
	if os.Getenv("FFXIVAPI_NOCACHE") == "" {
		log.Info("Using tcache-based caching client")

		client = &http.Client{
			Transport: &lodestone.TCacheRoundTripper{
				RoundTripper: http.DefaultTransport,
				Cache:        tcache.New(tcache.NewMemStorage()),
				MaxAge:       15 * time.Minute,
			},
		}
	}

	api := ffxivapi.New()
	api.Lodestone = &lodestone.HTTPClient{
		Server:     server,
		HTTPClient: client,
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
