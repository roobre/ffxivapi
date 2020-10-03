package lodestone

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"roob.re/tcache"
	"time"
)

type TCacheRoundTripper struct {
	RoundTripper http.RoundTripper
	Cache        tcache.Cache
	MaxAge       time.Duration
}

func (trt *TCacheRoundTripper) RoundTrip(r *http.Request) (response *http.Response, err error) {
	if r.Method != http.MethodGet {
		return trt.roundTrip(r)
	}

	k := r.URL.RequestURI()

	err = trt.Cache.From(r.Host).Access(k, trt.MaxAge, tcache.Handler{
		Then: func(r io.Reader) error {
			log.Println("cache hit for " + k)

			response, err = http.ReadResponse(bufio.NewReader(r), nil)
			return err
		},
		Else: func(w io.Writer) error {
			log.Println("cache miss for " + k)

			response, err = trt.roundTrip(r)
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
