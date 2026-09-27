package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/healthcheckmonitor/config"
)

func TestMonitor_HealthCheck(t *testing.T) {
	// Create a mock HTTP server that always returns 200 OK.
	mockSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockSrv.Close()

	// Build a minimal config pointing at the mock server.
	cfg := &config.Config{
		Services: []config.ServiceConfig{
			{
				Name:     "mock",
				URL:      mockSrv.URL,
				Interval: 100 * time.Millisecond,
			},
		},
	}

	mon := New(cfg)
	mon.Start()
	defer mon.Stop()

	// Allow a couple of intervals to pass.
	time.Sleep(250 * time.Millisecond)

	st, ok := mon.Get("mock")
	if !ok {
		t.Fatalf("expected service 'mock' to be present")
	}
	if !st.Healthy {
		t.Errorf("expected mock service to be healthy, got unhealthy: %s", st.Error)
	}
	if st.Error != "" {
		t.Errorf("expected no error message, got: %s", st.Error)
	}
}