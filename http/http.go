package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"roob.re/ffxivapi"
	"strconv"
)

type HTTPApi struct {
	*mux.Router
	xivapi *ffxivapi.FFXIVAPI
}

func NewHTTPApi() *HTTPApi {
	h := &HTTPApi{
		Router: mux.NewRouter(),
		xivapi: ffxivapi.New(),
	}

	h.HandleFunc("/", usage)
	h.HandleFunc("/character/search", h.search)
	h.HandleFunc("/character/{id}", h.character)

	return h
}

func (h *HTTPApi) search(rw http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	world := r.FormValue("world")
	if name == "" || world == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	results, err := h.xivapi.Search(name, world)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Header().Add("content-type", "application/json")

	je := json.NewEncoder(rw)
	je.Encode(results)
}

func (h *HTTPApi) character(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	features := uint(0)
	if r.FormValue("achievements") != "" {
		features |= ffxivapi.FeatureAchievements
	}
	if r.FormValue("classjob") != "" {
		features |= ffxivapi.FeatureClassJob
	}

	results, err := h.xivapi.Character(id, features)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Header().Add("content-type", "application/json")

	je := json.NewEncoder(rw)
	je.Encode(results)
}

func usage(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte(
		`# API Usage:

## Search
  /search?name={name}&world={world}
Example:
  /search?name=Roobre+Shiram&world=Ragnarok

## Character details
  /character/{id}/[?achievements=y]
Examples:
  /character/31688528
  /character/31688528?achievement=y
`,
	))
}
