package http

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/swaggo/http-swagger"
	"net/http"
	"roob.re/ffxivapi"
	"roob.re/ffxivapi/http/httpcache"
	"strconv"
	"time"
)

type HTTPApi struct {
	*mux.Router
	xivapi *ffxivapi.FFXIVAPI
	lc     *httpcache.HTTPCache
}

func NewHTTPApi() *HTTPApi {
	h := &HTTPApi{
		Router: mux.NewRouter(),
		xivapi: ffxivapi.New(),
		lc:     httpcache.New(),
	}

	h.Handle("/", http.RedirectHandler("/doc/", http.StatusMovedPermanently))
	h.HandleFunc("/character/search", h.lc.Cache(h.search, 1*time.Hour))
	h.HandleFunc("/character/{id}", h.lc.Cache(h.character, 15*time.Minute))
	h.HandleFunc("/character/{id}/avatar", h.lc.Cache(h.characterAvatar, 30*time.Minute))

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

	if len(results) == 0 {
		rw.WriteHeader(http.StatusNotFound)
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

	character, err := h.xivapi.Character(id, features)
	var herr ffxivapi.LodestoneHTTPError
	if errors.As(err, &herr) && herr == http.StatusNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Header().Add("content-type", "application/json")

	je := json.NewEncoder(rw)
	je.Encode(character)
}

func (h *HTTPApi) characterAvatar(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	character, err := h.xivapi.Character(id, 0)
	var herr ffxivapi.LodestoneHTTPError
	if errors.As(err, &herr) && herr == http.StatusNotFound {
		rw.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		rw.Write([]byte(err.Error()))
		return
	}

	http.Redirect(rw, r, character.Avatar, http.StatusFound)
}
