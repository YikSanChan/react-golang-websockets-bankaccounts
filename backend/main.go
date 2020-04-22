package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	balance map[string]int
	hub     *Hub
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
	vars := mux.Vars(r)
	accountId := vars["account_id"]
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
	accountId := vars["account_id"]
	deposit, err := strconv.Atoi(vars["deposit"])
	if err != nil {
		log.Printf("failed to parse deposit %s", vars["deposit"])
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newBalance := h.balance[accountId] + deposit
	body := fmt.Sprintf("{\"balance\": %d}", newBalance)
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, body); err != nil {
		log.Printf("failed to write balance result, error=%s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//for client, ok := range h.hub.clients {
	//	log.Println("client")
	//	if !ok {
	//		continue
	//	}
	//	client.hub.broadcast <- []byte("data")
	//}

	return
}

// serveWs handles websocket requests from the peer.
func (h *Handler) ServeWs(w http.ResponseWriter, r *http.Request) {
	active := 0
	for _, ok := range h.hub.clients {
		if ok {
			active += 1
		}
	}
	fmt.Printf("Serving ws to %d active clients\n", active) // TODO: figure out why ...
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	//defer conn.Close()
	client := &Client{hub: h.hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
}

func main() {
	router := mux.NewRouter()
	hub := newHub()
	handler := Handler{
		balance: make(map[string]int),
		hub:     hub,
	}
	go hub.run()

	router.HandleFunc("/ws", handler.ServeWs)
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
