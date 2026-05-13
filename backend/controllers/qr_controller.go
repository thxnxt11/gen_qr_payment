package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thxnxt11/payment_test/models"
	"github.com/thxnxt11/payment_test/services"
)

type QRController struct {
	service *services.QRService
}

func NewQRController(service *services.QRService) *QRController {
	return &QRController{service: service}
}

func (controller *QRController) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (controller *QRController) GenerateQR(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req models.QRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON body"})
		return
	}

	resp, err := controller.service.Generate(r.Context(), req)
	if err != nil {
		var serviceErr *services.ServiceError
		if errors.As(err, &serviceErr) {
			writeJSON(w, serviceErr.Status, models.ErrorResponse{Error: serviceErr.Message})
			return
		}
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
