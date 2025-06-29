package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// HealthCheck returns the application health status.
//
// GET /health
//
// Returns:
//   - 200: Application is healthy
//   - 503: Application has issues
func HealthCheck(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC(),
			Version:   version,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// ReadinessCheck returns whether the application is ready to serve traffic.
//
// GET /ready
//
// Returns:
//   - 200: Application is ready
//   - 503: Application is not ready
func ReadinessCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Add actual readiness checks here (database connectivity, etc.)
		ready := true

		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, err := w.Write([]byte("Not ready"))
			if err != nil {
				// Error writing response, but we've already set status
				return
			}
			return
		}

		response := HealthResponse{
			Status:    "ready",
			Timestamp: time.Now().UTC(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			// Error encoding response, but status already sent
			return
		}
	}
}
