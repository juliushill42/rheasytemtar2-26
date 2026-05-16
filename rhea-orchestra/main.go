// Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type ServiceHealth struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Healthy   bool      `json:"healthy"`
	LastCheck time.Time `json:"last_check"`
	Latency   int64     `json:"latency_ms"`
}

type Orchestra struct {
	services      map[string]string
	healthStatus  map[string]*ServiceHealth
	mu            sync.RWMutex
	selfHeal      bool
	httpClient    *http.Client
}

type OrchestraRequest struct {
	Action  string                 `json:"action"`
	Target  string                 `json:"target"`
	Payload map[string]interface{} `json:"payload"`
}

type OrchestraResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Latency int64                  `json:"latency_ms"`
}

func NewOrchestra() *Orchestra {
	return &Orchestra{
		services: map[string]string{
			"brain":     os.Getenv("RHEA_BRAIN_URL"),
			"sanctuary": os.Getenv("RHEA_SANCTUARY_URL"),
			"ledger":    os.Getenv("RHEA_LEDGER_URL"),
			"cloning":   os.Getenv("RHEA_CLONING_URL"),
		},
		healthStatus: make(map[string]*ServiceHealth),
		selfHeal:     os.Getenv("SELF_HEAL_ENABLED") == "true",
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
	}
}

func (o *Orchestra) checkHealth(name, url string) *ServiceHealth {
	start := time.Now()
	health := &ServiceHealth{
		Name:      name,
		URL:       url,
		LastCheck: time.Now(),
		Healthy:   false,
	}

	resp, err := o.httpClient.Get(url + "/health")
	if err == nil && resp.StatusCode == http.StatusOK {
		health.Healthy = true
		resp.Body.Close()
	}
	health.Latency = time.Since(start).Milliseconds()
	return health
}

func (o *Orchestra) healthMonitor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for name, url := range o.services {
				health := o.checkHealth(name, url)
				o.mu.Lock()
				o.healthStatus[name] = health
				o.mu.Unlock()

				if !health.Healthy && o.selfHeal {
					log.Printf("[ORCHESTRA] Service %s unhealthy, attempting recovery", name)
				}
			}
		}
	}
}

func (o *Orchestra) routeRequest(req *OrchestraRequest) *OrchestraResponse {
	start := time.Now()
	resp := &OrchestraResponse{Success: false}

	url, exists := o.services[req.Target]
	if !exists {
		resp.Error = "unknown target service"
		return resp
	}

	o.mu.RLock()
	health := o.healthStatus[req.Target]
	o.mu.RUnlock()

	if health != nil && !health.Healthy {
		resp.Error = "target service unhealthy"
		return resp
	}

	payload, _ := json.Marshal(req.Payload)
	httpReq, err := http.NewRequest("POST", url+"/"+req.Action, nil)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}

	if len(payload) > 0 {
		httpReq.Body = io.NopCloser(http.NoBody)
	}

	httpResp, err := o.httpClient.Do(httpReq)
	if err != nil {
		resp.Error = err.Error()
		return resp
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusOK {
		resp.Success = true
		var data map[string]interface{}
		if err := json.NewDecoder(httpResp.Body).Decode(&data); err == nil {
			resp.Data = data
		}
	} else {
		resp.Error = httpResp.Status
	}

	resp.Latency = time.Since(start).Milliseconds()
	return resp
}

func (o *Orchestra) handleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OrchestraRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := o.routeRequest(&req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (o *Orchestra) handleHealth(w http.ResponseWriter, r *http.Request) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	healthy := true
	for _, health := range o.healthStatus {
		if !health.Healthy {
			healthy = false
			break
		}
	}

	status := map[string]interface{}{
		"status":   "healthy",
		"services": o.healthStatus,
	}

	if !healthy {
		status["status"] = "degraded"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func main() {
	port := os.Getenv("ORCHESTRA_PORT")
	if port == "" {
		port = "9100"
	}

	orchestra := NewOrchestra()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go orchestra.healthMonitor(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/route", orchestra.handleRoute)
	mux.HandleFunc("/health", orchestra.handleHealth)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[ORCHESTRA] Starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ORCHESTRA] Fatal error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[ORCHESTRA] Shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
