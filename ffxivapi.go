package ffxivapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strconv"
)

// FFXIVAPI is the main object, containing the region to be targeted and the HTTP client to use
type FFXIVAPI struct {
	Region     string
	HTTPClient *http.Client
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

	response, err := api.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("lodestone request failed with status code %d", response.StatusCode))
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
