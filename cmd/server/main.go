package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/yourusername/dungeon-crawler/internal/db"
	"github.com/yourusername/dungeon-crawler/internal/mcp"
)

// CORS middleware to allow requests from the React frontend
func corsMiddleware(allowedOrigins []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

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
	// Configure allowed origins for CORS
	allowedOriginsEnv := os.Getenv("CORS_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv != "" {
		allowedOrigins = strings.Split(allowedOriginsEnv, ",")
	} else {
		// Default origins for development
		allowedOrigins = []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}
	}

	// Apply CORS middleware
	s.router.Use(corsMiddleware(allowedOrigins))

	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET", "OPTIONS")

	// MCP endpoints
	s.router.HandleFunc("/mcp/tools", s.handleListTools).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/mcp/call", s.handleCallTool).Methods("POST", "OPTIONS")

	// REST API endpoints (for future web UI)
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/character", s.handleCreateCharacter).Methods("POST", "OPTIONS")
	api.HandleFunc("/character/{id}", s.handleGetCharacter).Methods("GET", "OPTIONS")
	api.HandleFunc("/dungeon", s.handleCreateDungeon).Methods("POST", "OPTIONS")
	api.HandleFunc("/dungeon/{id}", s.handleGetDungeon).Methods("GET", "OPTIONS")

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
