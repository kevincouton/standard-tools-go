package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/core"
)

func NewRouter(state *AppState) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/api/v1/agent/tools", listTools)
	r.Post("/api/v1/agent/dispatch", dispatchTool(state))
	r.Get("/api/v1/market-data/{ticker}", fetchOhlcv(state))
	return r
}

func listTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, agent.ListTools())
}

func dispatchTool(state *AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Tool      string          `json:"tool"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		result, err := state.Dispatcher.Dispatch(r.Context(), agent.ToolCall{Name: req.Tool, Arguments: req.Arguments})
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, core.ErrInvalidCommand) || errors.Is(err, core.ErrInvalidTicker) || errors.Is(err, core.ErrInvalidDateRange) {
				status = http.StatusBadRequest
			}
			writeError(w, status, err)
			return
		}
		if result.Error != nil {
			writeError(w, http.StatusBadRequest, errors.New(*result.Error))
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func fetchOhlcv(state *AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker, err := core.NewTicker(chi.URLParam(r, "ticker"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		query := r.URL.Query()
		start, err := time.Parse("2006-01-02", query.Get("start"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		end, err := time.Parse("2006-01-02", query.Get("end"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		rng, err := core.NewDateRange(start, end)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		interval := core.Daily
		switch query.Get("interval") {
		case "weekly":
			interval = core.Weekly
		case "monthly":
			interval = core.Monthly
		}
		series, err := state.MarketData.Fetch(r.Context(), ticker, interval, rng, query.Get("provider"))
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err)
			return
		}
		writeJSON(w, http.StatusOK, series)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
