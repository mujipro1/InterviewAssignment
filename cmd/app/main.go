package main

import (
	"log"
	"net/http"
	"os"

	"assignment/internal/core"
	"assignment/internal/db"
	handlers "assignment/internal/http"
)

func main() {
	// Get database connection string from environment
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "host=postgres user=postgres password=postgres dbname=assignment sslmode=disable"
	}

	// Initialize database
	database, err := db.NewDB(connStr)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize services
	transactionService := core.NewTransactionService(database.DB)

	// Initialize handlers
	h := handlers.NewHandlers(transactionService)

	// Setup routes with custom router
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		// POST /user/{userId}/transaction
		if method == "POST" {
			if len(path) > 14 && path[:6] == "/user/" && path[len(path)-12:] == "/transaction" {
				h.HandleTransaction(w, r)
				return
			}
		}

		// GET /user/{userId}/balance
		if method == "GET" {
			if len(path) > 7 && path[:6] == "/user/" && path[len(path)-8:] == "/balance" {
				h.HandleGetBalance(w, r)
				return
			}
			// GET /health
			if path == "/health" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
				return
			}
		}

		// 404 for unmatched routes
		http.NotFound(w, r)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

