package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"assignment/internal/core"
	"assignment/internal/models"
	"assignment/internal/utils"
)

type Handlers struct {
	transactionService *core.TransactionService
}

func NewHandlers(transactionService *core.TransactionService) *Handlers {
	return &Handlers{
		transactionService: transactionService,
	}
}

func extractUserID(path string) string {
	// Path format: /user/{userId}/transaction or /user/{userId}/balance
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "user" {
		return parts[1]
	}
	return ""
}

func (h *Handlers) HandleTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get user ID from path
	userIDStr := extractUserID(r.URL.Path)
	userID, err := utils.ValidateUserID(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get Source-Type header
	sourceType := r.Header.Get("Source-Type")
	if sourceType == "" {
		respondError(w, http.StatusBadRequest, "Source-Type header is required")
		return
	}

	// Parse request body
	var req models.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Process transaction
	response, err := h.transactionService.ProcessTransaction(userID, req, sourceType)
	if err != nil {
		log.Printf("Error processing transaction: %v", err)
		
		// Check if it's a validation error (should return 400)
		errMsg := err.Error()
		if strings.Contains(errMsg, "invalid Source-Type header") ||
		   strings.Contains(errMsg, "invalid state") ||
		   strings.Contains(errMsg, "invalid amount format") ||
		   strings.Contains(errMsg, "invalid amount: cannot parse") ||
		   strings.Contains(errMsg, "invalid amount: cannot be negative") {
			respondError(w, http.StatusBadRequest, errMsg)
			return
		}
		
		// For other errors (like database errors), return 500
		respondError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	// Check if it's a duplicate or insufficient funds response
	if response.Message == "Duplicate transaction ignored" || response.Message == "Insufficient funds" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	respondJSON(w, response)
}

func (h *Handlers) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get user ID from path
	userIDStr := extractUserID(r.URL.Path)
	userID, err := utils.ValidateUserID(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get balance
	response, err := h.transactionService.GetBalance(userID)
	if err != nil {
		if err.Error() == "user not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("Error getting balance: %v", err)
		respondError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	respondJSON(w, response)
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

