package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	//"os"
	"strconv"
	"time"

	"github.com/alexandrevicenzi/go-sse"
	"github.com/gorilla/mux"
)

type Handler struct {
	balance   map[int]int
	sseServer *sse.Server
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request %s %s", r.Method, r.URL)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// at-most-once notification to subscribers of certain topic
func (h *Handler) publish(channel string) {
	h.sseServer.SendMessage(fmt.Sprintf("/events/%s", channel), sse.SimpleMessage("tick"))
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	if r.Method == http.MethodOptions {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	accountId, err := strconv.Atoi(vars["account_id"])
	if err != nil {
		log.Printf("failed to parse account id %s", vars["account_id"])
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body := fmt.Sprintf("{\"balance\": %d}", h.balance[accountId])
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
	accountId, err := strconv.Atoi(vars["account_id"])
	if err != nil {
		log.Printf("failed to parse account id %s", vars["account_id"])
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	deposit, err := strconv.Atoi(vars["deposit"])
	if err != nil {
		log.Printf("failed to parse deposit %s", vars["deposit"])
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	h.balance[accountId] += deposit

	//h.publish(vars["account_id"])

	return
}

func main() {
	sseServer := sse.NewServer(&sse.Options{
		// Increase default retry interval to 10s.
		//RetryInterval: 10 * 1000,
		// CORS headers
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Keep-Alive,X-Requested-With,Cache-Control,Content-Type,Last-Event-ID",
		},
		// Custom channel name generator
		//ChannelNameFunc: func(request *http.Request) string {
		//	return request.URL.Path
		//},
		// Print debug info
		//Logger: log.New(os.Stdout, "go-sse: ", log.Ldate|log.Ltime|log.Lshortfile)
	})
	defer sseServer.Shutdown()

	router := mux.NewRouter()

	handler := Handler{
		balance:   make(map[int]int),
		sseServer: sseServer,
	}

	router.Handle("/events/{channel}", sseServer)
	router.HandleFunc("/account/{account_id}/balance", handler.GetBalance)
	router.HandleFunc("/account/{account_id}/deposit/{deposit}", handler.Deposit).Methods(http.MethodPost)
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
