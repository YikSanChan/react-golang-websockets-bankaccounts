package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"nhooyr.io/websocket"
)

// webSocketServer enables broadcasting to a set of subscribers.
type webSocketServer struct {
	// subscriberMessageBuffer controls the max number
	// of messages that can be queued for a subscriber
	// before it is kicked.
	//
	// Defaults to 16.
	subscriberMessageBuffer int

	// publishLimiter controls the rate limit applied to the publish endpoint.
	//
	// Defaults to one publish every 100ms with a burst of 8.
	publishLimiter *rate.Limiter

	// logf controls where logs are sent.
	// Defaults to log.Printf.
	logf func(f string, v ...interface{})

	subscribersMu sync.Mutex
	subscribers   map[*subscriber]struct{}
}

// newWebSocketServer constructs a webSocketServer with the defaults.
func newWebSocketServer() *webSocketServer {
	return &webSocketServer{
		subscriberMessageBuffer: 16,
		logf:                    log.Printf,
		subscribers:             make(map[*subscriber]struct{}),
		publishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
	}
}

// subscriber represents a subscriber.
// Messages are sent on the msgs channel and if the client
// cannot keep up with the messages, closeSlow is called.
type subscriber struct {
	msgs      chan []byte
	closeSlow func()
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		h.wss.logf("%v", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	err = h.wss.subscribe(r.Context(), c)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		h.wss.logf("%v", err)
		return
	}
}

// subscribe subscribes the given WebSocket to all broadcast messages.
// It creates a subscriber with a buffered msgs chan to give some room to slower
// connections and then registers the subscriber. It then listens for all messages
// and writes them to the WebSocket. If the context is cancelled or
// an error occurs, it returns and deletes the subscription.
//
// It uses CloseRead to keep reading from the connection to process control
// messages and cancel the context if the connection drops.
func (wss *webSocketServer) subscribe(ctx context.Context, conn *websocket.Conn) error {
	ctx = conn.CloseRead(ctx)

	s := &subscriber{
		msgs: make(chan []byte, wss.subscriberMessageBuffer),
		closeSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}
	wss.addSubscriber(s)
	defer wss.deleteSubscriber(s)

	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, time.Second*5, conn, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// publish publishes the msg to all subscribers.
// It never blocks and so messages to slow subscribers
// are dropped.
func (wss *webSocketServer) publish(msg []byte) {
	wss.subscribersMu.Lock()
	defer wss.subscribersMu.Unlock()

	wss.publishLimiter.Wait(context.Background())

	for s := range wss.subscribers {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}

// addSubscriber registers a subscriber.
func (wss *webSocketServer) addSubscriber(s *subscriber) {
	wss.subscribersMu.Lock()
	wss.subscribers[s] = struct{}{}
	log.Printf("%d subscribers", len(wss.subscribers))
	wss.subscribersMu.Unlock()
}

// deleteSubscriber deletes the given subscriber.
func (wss *webSocketServer) deleteSubscriber(s *subscriber) {
	wss.subscribersMu.Lock()
	log.Println("Deleting subscriber")
	delete(wss.subscribers, s)
	wss.subscribersMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, conn *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return conn.Write(ctx, websocket.MessageText, msg)
}
