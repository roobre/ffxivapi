package http

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/swaggo/http-swagger"
	"net/http"
	"roob.re/ffxivapi"
	"roob.re/ffxivapi/lodestone"
	"strconv"
)

type Api struct {
	*mux.Router
	xivapi *ffxivapi.FFXIVAPI
}

func New() *Api {
	return NewWithApi(ffxivapi.New())
}

func NewWithApi(api *ffxivapi.FFXIVAPI) *Api {
	h := &Api{
		Router: mux.NewRouter(),
		xivapi: api,
	}

	h.Use(logRequest)
	h.Handle("/", http.RedirectHandler("/doc/", http.StatusMovedPermanently))
	h.HandleFunc("/character/search", h.search)
	h.HandleFunc("/character/{id}", h.character)
	h.HandleFunc("/character/{id}/avatar", h.characterAvatar)

	h.Handle("/swagger.yaml", http.FileServer(http.Dir("http")))
	h.PathPrefix("/doc").Handler(httpSwagger.Handler(httpSwagger.URL("/swagger.yaml")))

	return h
}

func (h *Api) search(rw http.ResponseWriter, r *http.Request) {
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

func (h *Api) character(rw http.ResponseWriter, r *http.Request) {
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
	var herr lodestone.HTTPError
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

func (h *Api) characterAvatar(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	character, err := h.xivapi.Character(id, 0)
	var herr lodestone.HTTPError
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

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Info(request.RequestURI)
		handler.ServeHTTP(writer, request)
	})
}
