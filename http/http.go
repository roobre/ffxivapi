package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/swaggo/http-swagger"
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

	h.Handle("/", http.RedirectHandler("/doc/", http.StatusMovedPermanently))
	h.HandleFunc("/character/search", h.search)
	h.HandleFunc("/character/{id}", h.character)

	h.Handle("/swagger.yaml", http.FileServer(http.Dir("http")))
	h.PathPrefix("/doc").Handler(httpSwagger.Handler(httpSwagger.URL("/swagger.yaml")))

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
