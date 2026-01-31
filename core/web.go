package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// WebServer represents the web interface server
type WebServer struct {
	server   *http.Server
	router   *mux.Router
	agent    *Agent
	logger   *logrus.Logger
	shutdown chan struct{}
}

// WebConfig represents web server configuration
type WebConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// NewWebServer creates a new web server instance
func NewWebServer(agent *Agent, config *WebConfig) (*WebServer, error) {
	if config == nil {
		config = &WebConfig{
			Host: "localhost",
			Port: 8080,
		}
	}

	router := mux.NewRouter()
	ws := &WebServer{
		router:   router,
		agent:    agent,
		logger:   agent.logger,
		shutdown: make(chan struct{}),
	}

	// Setup routes
	ws.setupRoutes()

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	ws.server = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return ws, nil
}

// setupRoutes sets up all web routes
func (ws *WebServer) setupRoutes() {
	// API routes
	api := ws.router.PathPrefix("/api/v1").Subrouter()

	// Agent info
	api.HandleFunc("/agent", ws.getAgentInfo).Methods("GET")
	
	// Events
	api.HandleFunc("/events", ws.postEvent).Methods("POST")
	api.HandleFunc("/events", ws.getEvents).Methods("GET")
	
	// Memory
	api.HandleFunc("/memory/short", ws.getShortTermMemory).Methods("GET")
	api.HandleFunc("/memory/long", ws.getLongTermMemory).Methods("GET")
	api.HandleFunc("/memory/short", ws.postShortTermMemory).Methods("POST")
	api.HandleFunc("/memory/long", ws.postLongTermMemory).Methods("POST")
	
	// Tools
	api.HandleFunc("/tools", ws.getTools).Methods("GET")
	api.HandleFunc("/tools/{name}/execute", ws.executeTool).Methods("POST")
	
	// Health check
	ws.router.HandleFunc("/health", ws.healthCheck).Methods("GET")
	
	// Static files for web UI
	ws.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))
	
	// Serve index.html for root and SPA routes
	ws.router.HandleFunc("/", ws.serveIndex).Methods("GET")
	ws.router.HandleFunc("/dashboard", ws.serveIndex).Methods("GET")
	ws.router.HandleFunc("/memory", ws.serveIndex).Methods("GET")
	ws.router.HandleFunc("/tools", ws.serveIndex).Methods("GET")
}

// Start starts the web server
func (ws *WebServer) Start(ctx context.Context) error {
	ws.logger.Infof("Starting web server on %s", ws.server.Addr)
	
	go func() {
		if err := ws.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ws.logger.WithError(err).Error("Web server failed")
		}
	}()
	
	return nil
}

// Stop gracefully stops the web server
func (ws *WebServer) Stop() error {
	ws.logger.Info("Stopping web server")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := ws.server.Shutdown(ctx); err != nil {
		ws.logger.WithError(err).Error("Web server shutdown failed")
		return ws.server.Close()
	}
	
	close(ws.shutdown)
	return nil
}

// API Handlers

type agentInfoResponse struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Workspace   string    `json:"workspace"`
	Uptime      time.Time `json:"uptime"`
	Status      string    `json:"status"`
}

func (ws *WebServer) getAgentInfo(w http.ResponseWriter, r *http.Request) {
	resp := agentInfoResponse{
		Name:        ws.agent.config.Agent.Name,
		Version:     ws.agent.config.Agent.Version,
		Description: ws.agent.config.Agent.Description,
		Workspace:   ws.agent.config.Agent.Workspace,
		Uptime:      time.Now(), // TODO: Track actual uptime
		Status:      "running",
	}
	
	ws.writeJSON(w, resp, http.StatusOK)
}

type eventRequest struct {
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type eventResponse struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

func (ws *WebServer) postEvent(w http.ResponseWriter, r *http.Request) {
	var req eventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	// Create event and emit to event loop
	event := &Event{
		ID:        generateID(),
		Type:      EventType(req.Type),
		Timestamp: time.Now(),
		Data:      req.Data,
		Metadata:  req.Metadata,
	}
	
	if err := ws.agent.eventLoop.Emit(event); err != nil {
		ws.logger.WithError(err).Error("Failed to emit event")
		http.Error(w, "Failed to process event", http.StatusInternalServerError)
		return
	}
	
	resp := eventResponse{
		ID:        event.ID,
		Type:      string(event.Type),
		Timestamp: event.Timestamp,
		Data:      event.Data,
		Metadata:  event.Metadata,
	}
	
	ws.writeJSON(w, resp, http.StatusCreated)
}

func (ws *WebServer) getEvents(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement event history retrieval
	// For now, return empty array
	resp := []eventResponse{}
	ws.writeJSON(w, resp, http.StatusOK)
}

func (ws *WebServer) getShortTermMemory(w http.ResponseWriter, r *http.Request) {
	if ws.agent.MemoryManager == nil {
		http.Error(w, "Memory manager not available", http.StatusInternalServerError)
		return
	}
	
	// Get all short-term memory entries
	entries, err := ws.agent.MemoryManager.GetAllShortTermMemory(r.Context())
	if err != nil {
		ws.logger.WithError(err).Error("Failed to get short-term memory")
		http.Error(w, "Failed to retrieve memory", http.StatusInternalServerError)
		return
	}
	
	ws.writeJSON(w, entries, http.StatusOK)
}

func (ws *WebServer) getLongTermMemory(w http.ResponseWriter, r *http.Request) {
	if ws.agent.MemoryManager == nil {
		http.Error(w, "Memory manager not available", http.StatusInternalServerError)
		return
	}
	
	// Get all long-term memory entries
	entries, err := ws.agent.MemoryManager.GetAllLongTermMemory(r.Context())
	if err != nil {
		ws.logger.WithError(err).Error("Failed to get long-term memory")
		http.Error(w, "Failed to retrieve memory", http.StatusInternalServerError)
		return
	}
	
	ws.writeJSON(w, entries, http.StatusOK)
}

type memoryEntryRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	TTL   *int64      `json:"ttl,omitempty"` // TTL in seconds
}

func (ws *WebServer) postShortTermMemory(w http.ResponseWriter, r *http.Request) {
	var req memoryEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if ws.agent.MemoryManager == nil {
		http.Error(w, "Memory manager not available", http.StatusInternalServerError)
		return
	}
	
	var ttl time.Duration
	if req.TTL != nil {
		ttl = time.Duration(*req.TTL) * time.Second
	}
	
	err := ws.agent.MemoryManager.SetShortTermMemory(r.Context(), req.Key, req.Value, ttl)
	if err != nil {
		ws.logger.WithError(err).Error("Failed to set short-term memory")
		http.Error(w, "Failed to store memory", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

func (ws *WebServer) postLongTermMemory(w http.ResponseWriter, r *http.Request) {
	var req memoryEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if ws.agent.MemoryManager == nil {
		http.Error(w, "Memory manager not available", http.StatusInternalServerError)
		return
	}
	
	err := ws.agent.MemoryManager.SetLongTermMemory(r.Context(), req.Key, req.Value)
	if err != nil {
		ws.logger.WithError(err).Error("Failed to set long-term memory")
		http.Error(w, "Failed to store memory", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

type toolResponse struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

func (ws *WebServer) getTools(w http.ResponseWriter, r *http.Request) {
	if ws.agent.ToolManager == nil {
		http.Error(w, "Tool manager not available", http.StatusInternalServerError)
		return
	}
	
	tools := ws.agent.ToolManager.ListTools()
	resp := make([]toolResponse, len(tools))
	
	for i, tool := range tools {
		resp[i] = toolResponse{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.Parameters,
		}
	}
	
	ws.writeJSON(w, resp, http.StatusOK)
}

type toolExecuteRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
}

type toolExecuteResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error,omitempty"`
}

func (ws *WebServer) executeTool(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	toolName := vars["name"]
	
	var req toolExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if ws.agent.ToolManager == nil {
		http.Error(w, "Tool manager not available", http.StatusInternalServerError)
		return
	}
	
	call := &ToolCall{
		ID:   generateID(),
		Name: toolName,
		Args: req.Parameters,
	}
	result, err := ws.agent.ToolManager.ExecuteTool(r.Context(), call)
	resp := toolExecuteResponse{}
	
	if err != nil {
		resp.Error = err.Error()
		ws.writeJSON(w, resp, http.StatusBadRequest)
		return
	}
	
	resp.Result = result
	ws.writeJSON(w, resp, http.StatusOK)
}

func (ws *WebServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status": "healthy",
		"uptime": time.Now(),
		"version": ws.agent.config.Agent.Version,
	}
	ws.writeJSON(w, resp, http.StatusOK)
}

func (ws *WebServer) serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/static/index.html")
}

func (ws *WebServer) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		ws.logger.WithError(err).Error("Failed to write JSON response")
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}