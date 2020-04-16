package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	balance int
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request %s %s", r.Method, r.URL)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	if r.Method == http.MethodOptions {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	body := fmt.Sprintf("{\"balance\": %d}", h.balance)
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, body); err != nil {
		log.Printf("failed to write balance result, error=%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	if r.Method == http.MethodOptions {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	deposit, err := strconv.Atoi(vars["deposit"])
	if err != nil {
		log.Printf("failed to parse deposit %s", vars["deposit"])
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.balance += deposit
	w.WriteHeader(http.StatusOK)
	return
}

func main() {
	router := mux.NewRouter()
	handler := Handler{
		balance: 42,
	}
	router.HandleFunc("/balance", handler.GetBalance)
	router.HandleFunc("/deposit/{deposit}", handler.Deposit).Methods(http.MethodPost)
	router.Use(mux.CORSMethodMiddleware(router))
	router.Use(loggingMiddleware)

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8080",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
