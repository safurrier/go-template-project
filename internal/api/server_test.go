package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/your-org/go-template-project/internal/db/sqlc"
)

type fakeStore struct {
	snap sqlc.DashboardSnapshot
	err  error
}

func (f fakeStore) GetDashboardSnapshotByTeamSlugWindow(context.Context, string, int32) (sqlc.DashboardSnapshot, error) {
	return f.snap, f.err
}

func TestDashboardEndpoint(t *testing.T) {
	s := NewServer(fakeStore{snap: sqlc.DashboardSnapshot{Payload: []byte(`{"team_slug":"arizona-mbb"}`)}})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/arizona-mbb/dashboard?window=5", nil)
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestDashboardNotFound(t *testing.T) {
	s := NewServer(fakeStore{err: pgx.ErrNoRows})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/arizona-mbb/dashboard", nil)
	s.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("code=%d", rr.Code)
	}
}
