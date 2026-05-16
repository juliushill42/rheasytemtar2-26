// Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

type AuditEntry struct {
	ID        string                 `json:"id"`
	Action    string                 `json:"action"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Hash      string                 `json:"hash"`
	PrevHash  string                 `json:"prev_hash"`
}

type Ledger struct {
	entries    []*AuditEntry
	mu         sync.RWMutex
	logPath    string
	lastHash   string
}

type QueryRequest struct {
	Action    string    `json:"action,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

func NewLedger() *Ledger {
	return &Ledger{
		entries:  make([]*AuditEntry, 0),
		logPath:  os.Getenv("AUDIT_LOG_PATH"),
		lastHash: "genesis",
	}
}

func (l *Ledger) calculateHash(entry *AuditEntry) string {
	data := fmt.Sprintf("%s|%s|%s|%v",
		entry.ID, entry.Action, entry.Timestamp.Format(time.RFC3339), entry.Metadata)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (l *Ledger) appendEntry(entry *AuditEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry.PrevHash = l.lastHash
	entry.Hash = l.calculateHash(entry)
	l.entries = append(l.entries, entry)
	l.lastHash = entry.Hash

	if l.logPath != "" {
		f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		data, _ := json.Marshal(entry)
		f.Write(data)
		f.WriteString("\n")
	}

	return nil
}

func (l *Ledger) handleAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var entry AuditEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entry.ID = fmt.Sprintf("audit_%d", time.Now().UnixNano())
	entry.Timestamp = time.Now()

	if err := l.appendEntry(&entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"id":      entry.ID,
		"hash":    entry.Hash,
	})
}

func (l *Ledger) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	results := make([]*AuditEntry, 0)
	limit := req.Limit
	if limit == 0 || limit > 1000 {
		limit = 1000
	}

	for i := len(l.entries) - 1; i >= 0 && len(results) < limit; i-- {
		entry := l.entries[i]

		if req.Action != "" && entry.Action != req.Action {
			continue
		}

		if !req.StartTime.IsZero() && entry.Timestamp.Before(req.StartTime) {
			continue
		}

		if !req.EndTime.IsZero() && entry.Timestamp.After(req.EndTime) {
			continue
		}

		results = append(results, entry)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": results,
		"count":   len(results),
	})
}

func (l *Ledger) handleVerify(w http.ResponseWriter, r *http.Request) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	valid := true
	prevHash := "genesis"

	for _, entry := range l.entries {
		if entry.PrevHash != prevHash {
			valid = false
			break
		}

		expectedHash := l.calculateHash(entry)
		if entry.Hash != expectedHash {
			valid = false
			break
		}

		prevHash = entry.Hash
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   valid,
		"entries": len(l.entries),
	})
}

func (l *Ledger) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"entries": len(l.entries),
	})
}

func main() {
	port := os.Getenv("LEDGER_PORT")
	if port == "" {
		port = "9104"
	}

	ledger := NewLedger()

	mux := http.NewServeMux()
	mux.HandleFunc("/audit", ledger.handleAudit)
	mux.HandleFunc("/query", ledger.handleQuery)
	mux.HandleFunc("/verify", ledger.handleVerify)
	mux.HandleFunc("/health", ledger.handleHealth)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("[LEDGER] Starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[LEDGER] Fatal error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[LEDGER] Shutting down")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
