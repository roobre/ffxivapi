package httpcache

import (
	"encoding/gob"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"roob.re/tcache"
	"roob.re/tcache/mapcache"
	"time"
)

func init() {
	gob.Register(cachedResponse{})
}

// HTTPCache is a middleware that caches HTTP responses, and serializes requests to achieve cache usage of parallel requests
type HTTPCache struct {
	cache tcache.Cache
}

func New() *HTTPCache {
	return &HTTPCache{
		cache: mapcache.New(),
	}
}

func (hc *HTTPCache) Cache(handler http.HandlerFunc, age time.Duration) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		hc.cache.From("http").Access(r.RequestURI, age, tcache.Handler{
			Then: func(r io.Reader) error {
				cachedResponse := cachedResponse{}
				err := gob.NewDecoder(r).Decode(&cachedResponse)
				if err != nil {
					log.Println(err)
				}

				err = cachedResponse.Write(rw)
				if err != nil {
					log.Printf("error writing cached request for %v: %v", r, err)
				}

				return nil
			},
			Else: func(w io.Writer) error {
				recorder := httptest.NewRecorder()
				handler(recorder, r)
				cachedResponse := Cache(recorder.Result())

				err := cachedResponse.Write(rw)
				if err != nil {
					log.Printf("error writing cached request for %s: %v", r.RequestURI, err)
				}

				return gob.NewEncoder(w).Encode(cachedResponse)
			},
		})
	}
}

// cachedResponse keeps track of a cached response, as well as implementing http.ResponseWriter
type cachedResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Cache creates a cached HTTP response from an http.Response
func Cache(response *http.Response) *cachedResponse {
	cr := &cachedResponse{}
	cr.StatusCode = response.StatusCode
	cr.Headers = response.Header
	cr.Body, _ = ioutil.ReadAll(response.Body)

	return cr
}

// Echo writes the cached response to the given writer
func (e *cachedResponse) Write(rw http.ResponseWriter) error {
	for k, vs := range e.Headers {
		for _, v := range vs {
			rw.Header().Add(k, v)
		}
	}

	rw.WriteHeader(e.StatusCode)

	_, err := rw.Write(e.Body)
	return err
}
