package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/yourusername/dungeon-crawler/internal/db"
	"github.com/yourusername/dungeon-crawler/internal/mcp"
)

type Server struct {
	db        *db.DB
	mcpServer *mcp.Server
	router    *mux.Router
}

func main() {
	// Get database path from env or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./dungeon-crawler.db"
	}

	// Initialize database
	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize MCP server
	mcpServer := mcp.NewServer()

	// Create server
	server := &Server{
		db:        database,
		mcpServer: mcpServer,
		router:    mux.NewRouter(),
	}

	// Setup routes
	server.setupRoutes()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting dungeon crawler server on port %s", port)
	log.Printf("Database: %s", dbPath)
	log.Fatal(http.ListenAndServe(":"+port, server.router))
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// MCP endpoints
	s.router.HandleFunc("/mcp/tools", s.handleListTools).Methods("GET")
	s.router.HandleFunc("/mcp/call", s.handleCallTool).Methods("POST")

	// REST API endpoints (for future web UI)
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/character", s.handleCreateCharacter).Methods("POST")
	api.HandleFunc("/character/{id}", s.handleGetCharacter).Methods("GET")
	api.HandleFunc("/dungeon", s.handleCreateDungeon).Methods("POST")
	api.HandleFunc("/dungeon/{id}", s.handleGetDungeon).Methods("GET")

	// Serve static files (future frontend)
	// s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
}

// Health check handler
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// MCP tool listing handler
func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	tools := s.mcpServer.ListTools()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": tools,
	})
}

// MCP tool call handler
func (s *Server) handleCallTool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.mcpServer.CallTool(req.Name, req.Arguments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// REST API handlers (stubs)

func (s *Server) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement character creation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Character creation not yet implemented",
	})
}

func (s *Server) handleGetCharacter(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement character retrieval
	vars := mux.Vars(r)
	characterID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Character retrieval not yet implemented",
		"character_id": characterID,
	})
}

func (s *Server) handleCreateDungeon(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement dungeon generation
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Dungeon generation not yet implemented",
	})
}

func (s *Server) handleGetDungeon(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement dungeon retrieval
	vars := mux.Vars(r)
	dungeonID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Dungeon retrieval not yet implemented",
		"dungeon_id": dungeonID,
	})
}
