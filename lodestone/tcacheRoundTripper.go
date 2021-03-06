package lodestone

import (
	"bufio"
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"roob.re/tcache"
	"time"
)

var errResponseNotOk = errors.New("not caching response as status code is >= 400")

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

	logpath := rq.URL.Path
	if rq.URL.RawQuery != "" {
		logpath += "?" + rq.URL.RawQuery
	}

	err = trt.Cache.Access(url, trt.MaxAge, tcache.Handler{
		Then: func(r io.Reader) error {
			log.Debug("hit " + logpath)
			response, err = http.ReadResponse(bufio.NewReader(r), nil)
			return err
		},
		Else: func(w io.Writer) error {
			log.Debug("miss " + logpath)

			response, err = trt.roundTrip(rq)
			if err != nil {
				return err
			}

			// Do not cache errored responses
			if response.StatusCode >= 400 {
				return errResponseNotOk
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

	if err == errResponseNotOk {
		return response, nil
	}
	return response, err
}

func (trt *TCacheRoundTripper) roundTrip(r *http.Request) (response *http.Response, err error) {
	return trt.RoundTripper.RoundTrip(r)
}
