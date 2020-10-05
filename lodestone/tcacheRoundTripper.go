package lodestone

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"net/http"
	"roob.re/tcache"
	"time"
)

type TCacheRoundTripper struct {
	RoundTripper http.RoundTripper
	Cache        *tcache.Cache
	MaxAge       time.Duration
}

func (trt *TCacheRoundTripper) RoundTrip(rq *http.Request) (response *http.Response, err error) {
	if rq.Method != http.MethodGet {
		return trt.roundTrip(rq)
	}

	url := rq.URL.String()

	err = trt.Cache.Access(url, trt.MaxAge, tcache.Handler{
		Then: func(r io.Reader) error {
			fmt.Printf("url=\"%s\" cache=\"%s\"\n", color.CyanString(url), color.GreenString("hit"))

			response, err = http.ReadResponse(bufio.NewReader(r), nil)
			return err
		},
		Else: func(w io.Writer) error {
			fmt.Printf("url=\"%s\" cache=\"%s\"\n", color.CyanString(url), color.YellowString("miss"))

			response, err = trt.roundTrip(rq)
			if err != nil {
				return err
			}

			newBody := &bytes.Buffer{}
			origBody := response.Body
			response.Body = ioutil.NopCloser(io.TeeReader(origBody, newBody))
			err = response.Write(w)
			_ = origBody.Close()
			response.Body = ioutil.NopCloser(newBody)
			return err
		},
	})

	return response, err
}

func (trt *TCacheRoundTripper) roundTrip(r *http.Request) (response *http.Response, err error) {
	return trt.RoundTripper.RoundTrip(r)
}
