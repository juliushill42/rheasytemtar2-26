// Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Blueprint struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Template    map[string]interface{} `json:"template"`
	Hash        string                 `json:"hash"`
	CreatedAt   time.Time              `json:"created_at"`
	ReplicaCount int                   `json:"replica_count"`
}

type CloneRequest struct {
	BlueprintID string                 `json:"blueprint_id"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

type Cloning struct {
	blueprints map[string]*Blueprint
	mu         sync.RWMutex
	dir        string
	ledgerURL  string
	httpClient *http.Client
}

func NewCloning() *Cloning {
	return &Cloning{
		blueprints: make(map[string]*Blueprint),
		dir:        os.Getenv("BLUEPRINT_DIR"),
		ledgerURL:  os.Getenv("RHEA_LEDGER_URL"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Cloning) calculateHash(bp *Blueprint) string {
	data, _ := json.Marshal(bp.Template)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (c *Cloning) saveBlueprint(bp *Blueprint) error {
	if c.dir == "" {
		return nil
	}

	data, err := json.MarshalIndent(bp, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.dir, bp.ID+".json")
	return os.WriteFile(path, data, 0644)
}

func (c *Cloning) loadBlueprints() error {
	if c.dir == "" {
		return nil
	}

	if _, err := os.Stat(c.dir); os.IsNotExist(err) {
		return nil
	}

	files, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		path := filepath.Join(c.dir, file.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var bp Blueprint
		if err := json.Unmarshal(data, &bp); err != nil {
			continue
		}

		c.blueprints[bp.ID] = &bp
	}

	return nil
}

func (c *Cloning) auditLog(action, blueprintID string, metadata map[string]interface{}) {
	if c.ledgerURL == "" {
		return
	}

	payload := map[string]interface{}{
		"action":       action,
		"blueprint_id": blueprintID,
		"timestamp":    time.Now(),
		"metadata":     metadata,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.ledgerURL+"/audit", nil)
	req.Body = io.NopCloser(http.NoBody)
	c.httpClient.Do(req)
}

func (c *Cloning) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var bp Blueprint
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bp.CreatedAt = time.Now()
	if bp.ID == "" {
		bp.ID = fmt.Sprintf("bp_%d", time.Now().UnixNano())
	}
	bp.Hash = c.calculateHash(&bp)
	bp.ReplicaCount = 0

	c.mu.Lock()
	c.blueprints[bp.ID] = &bp
	c.mu.Unlock()

	c.saveBlueprint(&bp)
	c.auditLog("blueprint_created", bp.ID, map[string]interface{}{"version": bp.Version})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      bp.ID,
		"hash":    bp.Hash,
	})
}

func (c *Cloning) handleClone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.mu.RLock()
	bp, exists := c.blueprints[req.BlueprintID]
	c.mu.RUnlock()

	if !exists {
		http.Error(w, "blueprint not found", http.StatusNotFound)
		return
	}

	c.mu.Lock()
	bp.ReplicaCount++
	replicaID := fmt.Sprintf("%s_replica_%d", bp.ID, bp.ReplicaCount)
	c.mu.Unlock()

	c.auditLog("clone_created", bp.ID, map[string]interface{}{"replica_id": replicaID})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"replica_id": replicaID,
		"template":   bp.Template,
	})
}

func (c *Cloning) handleList(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blueprints := make([]*Blueprint, 0, len(c.blueprints))
	for _, bp := range c.blueprints {
		blueprints = append(blueprints, bp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blueprints": blueprints,
		"count":      len(blueprints),
	})
}

func (c *Cloning) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"blueprints": len(c.blueprints),
	})
}

func main() {
	port := os.Getenv("CLONING_PORT")
	if port == "" {
		port = "9102"
	}

	cloning := NewCloning()
	if err := cloning.loadBlueprints(); err != nil {
		log.Printf("[CLONING] Warning: Failed to load blueprints: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/create", cloning.handleCreate)
	mux.HandleFunc("/clone", cloning.handleClone)
	mux.HandleFunc("/list", cloning.handleList)
	mux.HandleFunc("/health", cloning.handleHealth)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[CLONING] Starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[CLONING] Fatal error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[CLONING] Shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
