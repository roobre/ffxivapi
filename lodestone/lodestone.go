package lodestone

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Client is an object capable of returning HTML from the Lodestone
type Client interface {
	// Requests returns an io.ReaderCloser from which the HTML response associated to the given query can be read
	Request(query string) (io.ReadCloser, error)
}

// HTTPError is a non-200 status code from the lodestone server implemented as error
type HTTPError int

func (lhe HTTPError) Error() string {
	return fmt.Sprintf("lodestone returned status %d %s", lhe, http.StatusText(int(lhe)))
}

const LodestoneHTTPTimeout = 20 * time.Second

func CanonServerFromRegion(region string) string {
	return "https://" + region + ".finalfantasyxiv.com"
}

// HTTPClient uses an http.Client to connec to to the Lodestone and retrieve data
type HTTPClient struct {
	Server     string
	HTTPClient *http.Client
}

func (hlp *HTTPClient) Request(query string) (io.ReadCloser, error) {
	u := strings.TrimSuffix(hlp.Server, "/") + "/" + strings.TrimPrefix(query, "/")

	request, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("accept-language", "en-US,en;q=0.5")
	request.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; WOW64; rv:77.0) Gecko/20100101 Firefox/81.0")
	request.Header.Add("DNT", "1")

	var response *http.Response
	try := 1
	start := time.Now()
	for {
		response, err = hlp.HTTPClient.Do(request)
		if err != nil { // Request failed hard, return error
			return nil, err
		}

		// Everything went ok, break retry loop
		if response.StatusCode == http.StatusOK {
			break
		}

		// Return error if status code is not retry-able
		if !shouldRetry(response.StatusCode) || time.Since(start) > LodestoneHTTPTimeout {
			return nil, HTTPError(response.StatusCode)
		}

		// Linear backoff, wait between n and n+3 seconds where n is the attempt number
		wait := time.Second * time.Duration(retryMultiplier(response.StatusCode)*float64(1+rand.Intn(try+2)))
		log.Warnf("Lodestone replied with %d, retrying in %fs", response.StatusCode, wait.Seconds())
		time.Sleep(wait)
		try++
	}

	return response.Body, nil
}

// shouldRetry returns whether the non-200 status code is considered transient, and therefore the request should be retried
func shouldRetry(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// retryMultiplier returns a factor for the naive linear backoff algorithm
func retryMultiplier(statusCode int) float64 {
	switch statusCode {
	case http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return 2
	case http.StatusBadGateway:
		return 3
	default:
		return 1
	}
}
