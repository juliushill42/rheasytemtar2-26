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

type ThreatLevel int

const (
	LevelLow ThreatLevel = iota
	LevelMedium
	LevelHigh
	LevelCritical
)

type QuarantineEntry struct {
	ID          string                 `json:"id"`
	Source      string                 `json:"source"`
	ThreatLevel ThreatLevel            `json:"threat_level"`
	Payload     map[string]interface{} `json:"payload"`
	Hash        string                 `json:"hash"`
	Isolated    bool                   `json:"isolated"`
	Timestamp   time.Time              `json:"timestamp"`
	Notes       string                 `json:"notes,omitempty"`
}

type Sanctuary struct {
	quarantine map[string]*QuarantineEntry
	mu         sync.RWMutex
	dir        string
	ledgerURL  string
	httpClient *http.Client
}

type QuarantineRequest struct {
	Source      string                 `json:"source"`
	ThreatLevel int                    `json:"threat_level"`
	Payload     map[string]interface{} `json:"payload"`
	Notes       string                 `json:"notes,omitempty"`
}

func NewSanctuary() *Sanctuary {
	return &Sanctuary{
		quarantine: make(map[string]*QuarantineEntry),
		dir:        os.Getenv("QUARANTINE_DIR"),
		ledgerURL:  os.Getenv("RHEA_LEDGER_URL"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *Sanctuary) calculateHash(payload map[string]interface{}) string {
	data, _ := json.Marshal(payload)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *Sanctuary) saveEntry(entry *QuarantineEntry) error {
	if s.dir == "" {
		return nil
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, entry.ID+".json")
	return os.WriteFile(path, data, 0644)
}

func (s *Sanctuary) auditLog(action, entryID string, metadata map[string]interface{}) {
	if s.ledgerURL == "" {
		return
	}

	payload := map[string]interface{}{
		"action":    action,
		"entry_id":  entryID,
		"timestamp": time.Now(),
		"metadata":  metadata,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", s.ledgerURL+"/audit", nil)
	req.Body = io.NopCloser(http.NoBody)
	s.httpClient.Do(req)
}

func (s *Sanctuary) handleQuarantine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QuarantineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entry := &QuarantineEntry{
		ID:          fmt.Sprintf("quar_%d", time.Now().UnixNano()),
		Source:      req.Source,
		ThreatLevel: ThreatLevel(req.ThreatLevel),
		Payload:     req.Payload,
		Hash:        s.calculateHash(req.Payload),
		Isolated:    true,
		Timestamp:   time.Now(),
		Notes:       req.Notes,
	}

	s.mu.Lock()
	s.quarantine[entry.ID] = entry
	s.mu.Unlock()

	s.saveEntry(entry)
	s.auditLog("quarantine_created", entry.ID, map[string]interface{}{
		"threat_level": entry.ThreatLevel,
		"source":       entry.Source,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      entry.ID,
		"hash":    entry.Hash,
	})
}

func (s *Sanctuary) handleRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	entry, exists := s.quarantine[req.ID]
	if exists {
		entry.Isolated = false
	}
	s.mu.Unlock()

	if !exists {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}

	s.saveEntry(entry)
	s.auditLog("quarantine_released", entry.ID, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      entry.ID,
	})
}

func (s *Sanctuary) handleList(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]*QuarantineEntry, 0, len(s.quarantine))
	for _, entry := range s.quarantine {
		entries = append(entries, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}

func (s *Sanctuary) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	isolated := 0
	for _, entry := range s.quarantine {
		if entry.Isolated {
			isolated++
		}
	}
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"total":    len(s.quarantine),
		"isolated": isolated,
	})
}

func main() {
	port := os.Getenv("SANCTUARY_PORT")
	if port == "" {
		port = "9103"
	}

	sanctuary := NewSanctuary()

	mux := http.NewServeMux()
	mux.HandleFunc("/quarantine", sanctuary.handleQuarantine)
	mux.HandleFunc("/release", sanctuary.handleRelease)
	mux.HandleFunc("/list", sanctuary.handleList)
	mux.HandleFunc("/health", sanctuary.handleHealth)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[SANCTUARY] Starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[SANCTUARY] Fatal error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SANCTUARY] Shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
