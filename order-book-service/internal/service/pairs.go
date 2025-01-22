package service

import (
	"encoding/json"
	"net/http"
	"order-book-service/internal/repository"

	"github.com/gorilla/mux"
)

type PairsAPI struct {
	repo repository.Repository
	r    *mux.Router
}

func NewPairsAPI(repo repository.Repository) *PairsAPI {
	api := &PairsAPI{
		repo: repo,
		r:    mux.NewRouter(),
	}
	api.setupRoutes()
	return api
}

func (api *PairsAPI) setupRoutes() {
	api.r.HandleFunc("/v1/exchange/{exchange}/pairs", api.getPairs).Methods("GET")
	api.r.HandleFunc("/v1/exchange/{exchange}/pairs", api.updatePairs).Methods("PUT")
}

func (api *PairsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.r.ServeHTTP(w, r)
}

type PairsResponse struct {
	Exchange string   `json:"exchange"`
	Pairs    []string `json:"pairs"`
}

func (api *PairsAPI) getPairs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exchange := vars["exchange"]

	pairs, err := api.repo.GetPairs(r.Context(), exchange)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(PairsResponse{
		Exchange: exchange,
		Pairs:    pairs,
	})
}

func (api *PairsAPI) updatePairs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exchange := vars["exchange"]

	var req PairsResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := api.repo.SavePairs(r.Context(), exchange, req.Pairs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
