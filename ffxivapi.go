package ffxivapi

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// FFXIVAPI is the main object, containing the region to be targeted and the HTTP client to use
type FFXIVAPI struct {
	Region     string
	HTTPClient *http.Client
}

type LodestoneHTTPError int

func (lhe LodestoneHTTPError) Error() string {
	return fmt.Sprintf("lodestone returned status code %d", lhe)
}

// New returns a new FFXIVAPI object with http.DefaultClient and the region set to Europe ("eu")
func New() *FFXIVAPI {
	return &FFXIVAPI{
		Region:     "eu",
		HTTPClient: http.DefaultClient,
	}
}

// lodestone queries the given lodestone URL and params (url-encoding them) and returns a goquery document
func (api *FFXIVAPI) lodestone(query string, params map[string]string) (*goquery.Document, error) {
	u := api.url(query)

	if len(params) > 0 {
		u += "?"
		urlValues := url.Values{}
		for k, v := range params {
			urlValues.Add(k, v)
		}
		u += urlValues.Encode()
	}

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
		response, err = api.HTTPClient.Do(request)
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

		// Out of retires or other non-ok stetus, return error
		if response.StatusCode != http.StatusOK {
			return nil, LodestoneHTTPError(response.StatusCode)
		}

		break
	}

	return goquery.NewDocumentFromReader(response.Body)
}

func (api *FFXIVAPI) url(query string) string {
	return "https://" + api.Region + ".finalfantasyxiv.com" + query
}

func silentAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
