package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/core"
)

const dateFormat = "2006-01-02"

func NewRouter(state *AppState) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/api/v1/agent/tools", listTools)
	r.Post("/api/v1/agent/dispatch", dispatchTool(state))
	r.Get("/api/v1/market-data/{ticker}", fetchOhlcv(state))
	r.Get("/a2a/agent.json", a2aAgentCard)
	r.Post("/a2a/tasks", a2aDispatchTask(state))
	r.Get("/mcp/capabilities", mcpCapabilities)
	r.Post("/mcp/tools/list", mcpListTools)
	r.Post("/mcp/tools/call", mcpCallTool(state))
	return r
}

func listTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, agent.ListTools())
}

func dispatchTool(state *AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeToolCall(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.Tool == "" {
			writeError(w, http.StatusBadRequest, errors.New("tool is required"))
			return
		}
		result, err := state.Dispatcher.Dispatch(r.Context(), agent.ToolCall{Name: req.Tool, Arguments: req.Arguments})
		if err != nil {
			writeError(w, domainErrorStatus(err), err)
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
		start, err := time.Parse(dateFormat, query.Get("start"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		end, err := time.Parse(dateFormat, query.Get("end"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		rng, err := core.NewDateRange(start, end)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		interval, err := core.ParseBarInterval(query.Get("interval"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		series, err := state.MarketData.Fetch(r.Context(), ticker, interval, rng, query.Get("provider"))
		if err != nil {
			writeError(w, domainErrorStatus(err), err)
			return
		}
		writeJSON(w, http.StatusOK, series)
	}
}

// domainErrorStatus maps core domain errors to HTTP status codes.
//
// Mapping rules:
//   - Invalid input errors -> 400 Bad Request
//   - Not found -> 404 Not Found
//   - Provider/data-quality problems from upstream -> 502 Bad Gateway / 503 Service Unavailable
//   - Internal and any unrecognized errors -> 500 Internal Server Error
func domainErrorStatus(err error) int {
	switch {
	case errors.Is(err, core.ErrInvalidCommand),
		errors.Is(err, core.ErrInvalidTicker),
		errors.Is(err, core.ErrInvalidDateRange):
		return http.StatusBadRequest
	case errors.Is(err, core.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, core.ErrProviderNotAvailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, core.ErrDataQuality):
		return http.StatusBadGateway
	case errors.Is(err, core.ErrInternal):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// errorResponse is the canonical error response body.
type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{
		Error: err.Error(),
		Code:  errorCode(status),
	})
}

func errorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	default:
		return "UNKNOWN"
	}
}
