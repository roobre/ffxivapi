package lockingcache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// LockingCache is a middleware that caches HTTP responses, and serializes requests to achieve cache usage of parallel requests
type LockingCache struct {
	MaxEntries int

	mu    sync.Mutex
	cache map[string]*LockingEntry
}

func New() *LockingCache {
	return &LockingCache{
		cache: map[string]*LockingEntry{},
	}
}

type LockingEntry struct {
	sync.RWMutex
	cacheEntry
}

func (lc *LockingCache) Cache(hf http.HandlerFunc, age time.Duration) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Lock for searching
		lc.mu.Lock()
		entry, found := lc.cache[lc.key(r)]

		if found {
			if time.Since(entry.Created) < age {
				// Entry found and valid, lock the entry, release search
				entry.RLock()
				lc.mu.Unlock()
				defer entry.RUnlock()

				log.Printf("Serving %s from cache", lc.key(r))
				err := entry.Echo(rw)
				if err != nil {
					log.Printf("error writing cached request for %v: %v", r, err)
				}
				return
			} else {
				delete(lc.cache, lc.key(r))
			}
		}

		// Allocate new entry, lock it, put it in the map, and unlock global map
		entry = &LockingEntry{}
		lc.add(r, entry)
		entry.Lock()
		defer entry.Unlock()
		lc.mu.Unlock()

		entry.Created = time.Now()
		entry.Record()

		entry.Header().Add("cache-control", fmt.Sprintf("max-age=%d", int(age.Seconds())))
		hf(entry, r)

		err := entry.Echo(rw)
		if err != nil {
			log.Printf("error writing cached request for %s: %v", lc.key(r), err)
		}
	}
}

func (lc *LockingCache) key(r *http.Request) string {
	return r.RequestURI
}

func (lc *LockingCache) add(r *http.Request, entry *LockingEntry) {
	if lc.MaxEntries != 0 && len(lc.cache) >= lc.MaxEntries {
		t := time.Now()
		oldest := ""
		for k, v := range lc.cache {
			if v.Created.Before(t) {
				t = v.Created
				oldest = k
			}
		}

		delete(lc.cache, oldest)
	}

	lc.cache[lc.key(r)] = entry
}

// cacheEntry keeps track of a cached response, as well as implementing http.ResponseWriter
type cacheEntry struct {
	httptest.ResponseRecorder
	Created time.Time

	bodyBuffer []byte
}

// Record prepares a cacheEntry to record data
func (e *cacheEntry) Record() {
	e.Body = &bytes.Buffer{}
}

// Echo writes the cached response to the given writer
func (e *cacheEntry) Echo(rw http.ResponseWriter) error {
	response := e.Result()

	for k, vs := range response.Header {
		for _, v := range vs {
			rw.Header().Add(k, v)
		}
	}

	rw.WriteHeader(response.StatusCode)

	if e.bodyBuffer == nil {
		e.bodyBuffer, _ = ioutil.ReadAll(response.Body)
	}

	_, err := rw.Write(e.bodyBuffer)
	return err
}
