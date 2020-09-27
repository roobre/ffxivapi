package ffxivapi

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"strconv"
)

type FFXIVAPI struct {
	Region     string
	HTTPClient *http.Client
}

func New() *FFXIVAPI {
	return &FFXIVAPI{
		Region:     "eu",
		HTTPClient: http.DefaultClient,
	}
}

type URLParam struct {
	K string
	V string
}

func (api *FFXIVAPI) lodestone(query string, params ...URLParam) (*goquery.Document, error) {
	u := api.url(query)

	if len(params) > 0 {
		u += "?"
		urlValues := url.Values{}
		for _, v := range params {
			urlValues.Add(v.K, v.V)
		}
		u += urlValues.Encode()
	}

	request, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("accept-language", "en-US,en;q=0.5")
	//request.Header.Add("accept-encoding", "gzip, deflate, br")
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
