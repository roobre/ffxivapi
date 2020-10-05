package ffxivapi // import "roob.re/ffxivapi"

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"roob.re/ffxivapi/lodestone"
	"roob.re/tcache"
	"strconv"
	"time"
)

// FFXIVAPI is the main object, containing the region to be targeted and the HTTP client to use
type FFXIVAPI struct {
	Lodestone lodestone.Client
}

// New returns a new FFXIVAPI object with http.DefaultClient and the region set to Europe ("eu")
func New() *FFXIVAPI {
	return &FFXIVAPI{
		Lodestone: &lodestone.HTTPClient{
			Region: "eu",
			HTTPClient: &http.Client{
				Transport: &lodestone.TCacheRoundTripper{
					RoundTripper: http.DefaultTransport,
					Cache:        tcache.New(tcache.NewMemStorage()),
					MaxAge:       15 * time.Minute,
				},
			},
		},
	}
}

// lodestone queries the given lodestone URL and params (url-encoding them) and returns a goquery document
func (api *FFXIVAPI) lodestone(query string, params map[string]string) (*goquery.Document, error) {
	if len(params) > 0 {
		query += "?"
		urlValues := url.Values{}
		for k, v := range params {
			urlValues.Add(k, v)
		}
		query += urlValues.Encode()
	}

	response, err := api.Lodestone.Request(query)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromReader(response)
}

// silentAtoi discards error from atoi, used to assign numbers assumed to be correctly-formatted into inline initializers
func silentAtoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
