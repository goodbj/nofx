package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Create router
	r := mux.NewRouter()

	// Add health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Add proxy endpoints for futures API
	r.HandleFunc("/fapi/v2/balance", handleBalance).Methods("GET")
	r.HandleFunc("/fapi/v2/account", handleAccount).Methods("GET")
	r.HandleFunc("/fapi/v2/positionRisk", handlePositions).Methods("GET")
	r.HandleFunc("/fapi/v1/order", handleOrder).Methods("POST", "DELETE", "GET")
	r.HandleFunc("/fapi/v1/openOrders", handleOpenOrders).Methods("GET")
	r.HandleFunc("/fapi/v1/allOrders", handleAllOrders).Methods("GET")
	r.HandleFunc("/fapi/v1/ticker/price", handleTickerPrice).Methods("GET")
	r.HandleFunc("/fapi/v1/ticker/bookTicker", handleBookTicker).Methods("GET")
	r.HandleFunc("/fapi/v1/klines", handleKlines).Methods("GET")

	// Add catch-all handler for any other endpoints
	r.PathPrefix("/").HandlerFunc(handleGeneric)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Default port
	}

	// Create HTTP server with timeout settings
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("🚀 Binance Proxy Server starting on port %s", port)
	log.Printf("📊 Health check available at: http://localhost:%s/health", port)
	log.Printf("🔗 Default Target API URL: %s", defaultTargetAPIURL)

	// Start server
	log.Fatal(server.ListenAndServe())
}
