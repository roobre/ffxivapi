package ffxivapi

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// LodestoneClient is an object capable of returning HTML from the Lodestone
type LodestoneClient interface {
	// Requests returns an io.ReaderCloser from which the HTML response associated to the given query can be read
	Request(query string) (io.ReadCloser, error)
}

// LodestoneHTTPError is a non-200 status code from the lodestone server implemented as error
type LodestoneHTTPError int

func (lhe LodestoneHTTPError) Error() string {
	return fmt.Sprintf("lodestone returned status %d %s", lhe, http.StatusText(int(lhe)))
}

// HTTPLoadestoneProvider uses an http.Client to connecto to the Lodestone and retrieve data
type LodestoneHTTPClient struct {
	Region     string
	HTTPClient *http.Client
}

func (hlp *LodestoneHTTPClient) Request(query string) (io.ReadCloser, error) {
	u := "https://" + hlp.Region + ".finalfantasyxiv.com" + query

	request, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("accept-language", "en-US,en;q=0.5")
	request.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:77.0) Gecko/20100101 Firefox/81.0")
	request.Header.Add("DNT", "1")

	var response *http.Response
	try := 1
	for {
		response, err = hlp.HTTPClient.Do(request)
		if err != nil { // Request failed hard, return error
			return nil, err
		}

		// If TooManyRequests and retries left, sleep then continue
		if try <= 10 && response.StatusCode == http.StatusTooManyRequests {
			// Linear backoff, wait between n and n+3 seconds where n is the attempt number
			retry := time.Duration(1+rand.Intn(try+2)) * time.Second
			log.Printf("Lodestone ratelimit hit, retrying in %fs", retry.Seconds())
			time.Sleep(retry)
			try++
			continue
		}

		// Out of retires or other non-ok status, return error
		if response.StatusCode != http.StatusOK {
			return nil, LodestoneHTTPError(response.StatusCode)
		}

		break
	}

	return response.Body, nil
}
