// Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Rules       []Rule                 `json:"rules"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type Rule struct {
	Condition string                 `json:"condition"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params,omitempty"`
}

type Brain struct {
	policies    map[string]*Policy
	mu          sync.RWMutex
	storeDir    string
	contextSize string
}

type DecisionRequest struct {
	Context map[string]interface{} `json:"context"`
	Query   string                 `json:"query"`
}

type DecisionResponse struct {
	Decision string                 `json:"decision"`
	Policy   string                 `json:"policy"`
	Metadata map[string]interface{} `json:"metadata"`
}

func NewBrain() *Brain {
	return &Brain{
		policies:    make(map[string]*Policy),
		storeDir:    os.Getenv("POLICY_STORE"),
		contextSize: os.Getenv("GLOBAL_CONTEXT_WINDOW"),
	}
}

func (b *Brain) loadPolicies() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	defaultPolicy := &Policy{
		ID:       "default",
		Name:     "Default Policy",
		Priority: 0,
		Enabled:  true,
		Rules: []Rule{
			{
				Condition: "always",
				Action:    "allow",
				Params:    map[string]interface{}{"log": true},
			},
		},
		CreatedAt: time.Now(),
	}
	b.policies["default"] = defaultPolicy
	return nil
}

func (b *Brain) evaluateDecision(req *DecisionRequest) *DecisionResponse {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var selectedPolicy *Policy
	highestPriority := -1

	for _, policy := range b.policies {
		if !policy.Enabled {
			continue
		}
		if policy.Priority > highestPriority {
			highestPriority = policy.Priority
			selectedPolicy = policy
		}
	}

	resp := &DecisionResponse{
		Decision: "allow",
		Policy:   "default",
		Metadata: make(map[string]interface{}),
	}

	if selectedPolicy != nil {
		resp.Policy = selectedPolicy.ID
		for _, rule := range selectedPolicy.Rules {
			if rule.Action == "deny" {
				resp.Decision = "deny"
				break
			}
		}
	}

	return resp
}

func (b *Brain) handleDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := b.evaluateDecision(&req)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (b *Brain) handlePolicyCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var policy Policy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	policy.CreatedAt = time.Now()
	if policy.ID == "" {
		policy.ID = fmt.Sprintf("policy_%d", time.Now().Unix())
	}

	b.mu.Lock()
	b.policies[policy.ID] = &policy
	b.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      policy.ID,
	})
}

func (b *Brain) handlePolicyList(w http.ResponseWriter, r *http.Request) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	policies := make([]*Policy, 0, len(b.policies))
	for _, p := range b.policies {
		policies = append(policies, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"policies": policies,
		"count":    len(policies),
	})
}

func (b *Brain) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "healthy",
		"policies":     len(b.policies),
		"context_size": b.contextSize,
	})
}

func main() {
	port := os.Getenv("BRAIN_PORT")
	if port == "" {
		port = "9101"
	}

	brain := NewBrain()
	if err := brain.loadPolicies(); err != nil {
		log.Fatalf("[BRAIN] Failed to load policies: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/decision", brain.handleDecision)
	mux.HandleFunc("/policy/create", brain.handlePolicyCreate)
	mux.HandleFunc("/policy/list", brain.handlePolicyList)
	mux.HandleFunc("/health", brain.handleHealth)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[BRAIN] Starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[BRAIN] Fatal error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[BRAIN] Shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
