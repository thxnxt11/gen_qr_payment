package routes

import (
	"net/http"

	"github.com/thxnxt11/payment_test/controllers"
)

func NewRouter(corsOrigin string, controller *controllers.QRController) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", controller.Health)
	mux.HandleFunc("/api/qr", controller.GenerateQR)
	return withCORS(corsOrigin, mux)
}
