package api

import (
	"encoding/json"
	"net/http"

	"github.com/dvjn/sorcerer/pkg/models"
)

func sendError(w http.ResponseWriter, code string, message string, status int) {
	errorResponse := models.ErrorResponse{
		Errors: []models.Error{
			{
				Code:    code,
				Message: message,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse)
}
