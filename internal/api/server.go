package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/your-org/go-template-project/internal/db/sqlc"
)

type SnapshotStore interface {
	GetDashboardSnapshotByTeamSlugWindow(ctx context.Context, slug string, window int32) (sqlc.DashboardSnapshot, error)
}

type Server struct{ q SnapshotStore }

func NewServer(q SnapshotStore) *Server { return &Server{q: q} }

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/teams/", s.handleDashboard)
	return mux
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 5 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "teams" || parts[4] != "dashboard" {
		http.NotFound(w, r)
		return
	}
	slug := parts[3]
	window := int32(5)
	if param := r.URL.Query().Get("window"); param != "" {
		parsed, err := strconv.Atoi(param)
		if err != nil || parsed <= 0 {
			http.Error(w, "invalid window", http.StatusBadRequest)
			return
		}
		window = int32(parsed)
	}
	snap, err := s.q.GetDashboardSnapshotByTeamSlugWindow(r.Context(), slug, window)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "snapshot not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(snap.Payload)
}
