package monitor

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/example/healthcheckmonitor/config"
	"github.com/example/healthcheckmonitor/internal/logger"
	"github.com/example/healthcheckmonitor/internal/types"
)

// Monitor performs periodic health checks for a set of services.
type Monitor struct {
	services []config.ServiceConfig
	statuses map[string]types.ServiceStatus
	mu       sync.RWMutex
	client   *http.Client
	ctx      context.Context
	cancel   context.CancelFunc
}

// New creates a new Monitor instance.
func New(cfg *config.Config) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Monitor{
		services: cfg.Services,
		statuses: make(map[string]types.ServiceStatus),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		ctx:    ctx,
		cancel: cancel,
	}
	return m
}

// Start launches a goroutine per service that checks health at the configured interval.
func (m *Monitor) Start() {
	for _, svc := range m.services {
		svc := svc // capture loop variable
		go m.runServiceLoop(svc)
	}
}

// Stop signals all monitoring goroutines to exit.
func (m *Monitor) Stop() {
	m.cancel()
}

// runServiceLoop executes health checks for a single service.
func (m *Monitor) runServiceLoop(svc config.ServiceConfig) {
	ticker := time.NewTicker(svc.Interval)
	defer ticker.Stop()

	// Initial check immediately
	m.checkAndStore(svc)

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkAndStore(svc)
		}
	}
}

// checkAndStore performs a single HTTP GET request and records the result.
func (m *Monitor) checkAndStore(svc config.ServiceConfig) {
	start := time.Now()
	req, err := http.NewRequestWithContext(m.ctx, http.MethodGet, svc.URL, nil)
	if err != nil {
		m.storeStatus(svc, false, start, err.Error())
		return
	}
	resp, err := m.client.Do(req)
	if err != nil {
		m.storeStatus(svc, false, start, err.Error())
		return
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode >= 200 && resp.StatusCode < 300
	var errMsg string
	if !healthy {
		errMsg = http.StatusText(resp.StatusCode)
	}
	m.storeStatus(svc, healthy, start, errMsg)
}

// storeStatus updates the internal map with thread‑safety.
func (m *Monitor) storeStatus(svc config.ServiceConfig, healthy bool, ts time.Time, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statuses[svc.Name] = types.ServiceStatus{
		Name:        svc.Name,
		URL:         svc.URL,
		Healthy:     healthy,
		LastChecked: ts,
		Error:       errMsg,
	}
	if healthy {
		logger.Log.Infof("service %s is healthy", svc.Name)
	} else {
		logger.Log.Warnf("service %s unhealthy: %s", svc.Name, errMsg)
	}
}

// GetAll returns a snapshot of all service statuses.
func (m *Monitor) GetAll() []types.ServiceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]types.ServiceStatus, 0, len(m.statuses))
	for _, st := range m.statuses {
		out = append(out, st)
	}
	return out
}

// Get returns the status for a single service; ok == false if not found.
func (m *Monitor) Get(name string) (types.ServiceStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	st, ok := m.statuses[name]
	return st, ok
}