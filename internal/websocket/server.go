// File: internal/websocket/server.go
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/bass4/dcs-ice/pkg/models"
)

// Server handles WebSocket connections
type Server struct {
	clients      map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	broadcast    chan interface{}
	actionChan   chan models.Action
	upgrader     websocket.Upgrader
	mutex        sync.RWMutex
	actionManager ActionManagerInterface
}

// ActionManagerInterface defines the interface for action manager interactions
type ActionManagerInterface interface {
	RegisterObserver(observer chan<- models.Action)
	UnregisterObserver(observer chan<- models.Action)
}

// NewServer creates a new WebSocket server
func NewServer(actionManager ActionManagerInterface) *Server {
	server := &Server{
		clients:      make(map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan interface{}),
		actionChan:   make(chan models.Action, 100), // Buffer 100 actions
		actionManager: actionManager,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all connections in development
				// In production, this should be more restrictive
				return true
			},
		},
	}

	// Register for action notifications
	actionManager.RegisterObserver(server.actionChan)

	// Start the server
	go server.run()

	return server
}

// run starts the WebSocket server's main loop
func (s *Server) run() {
	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.clients[client] = true
			s.mutex.Unlock()

		case client := <-s.unregister:
			s.mutex.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mutex.Unlock()

		case message := <-s.broadcast:
			s.mutex.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					// Client can't keep up with messages, remove them
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mutex.RUnlock()

		case action := <-s.actionChan:
			// Forward the action to all clients
			s.broadcast <- action
		}
	}
}

// HandleWebSocket handles WebSocket requests from clients
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		server: s,
		conn:   conn,
		send:   make(chan interface{}, 256),
	}

	// Register new client
	s.register <- client

	// Start client routines
	go client.writePump()
	go client.readPump()
}

// SendMessage broadcasts a message to all connected clients
func (s *Server) SendMessage(message interface{}) {
	s.broadcast <- message
}

// Close shuts down the WebSocket server
func (s *Server) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Close all client connections
	for client := range s.clients {
		close(client.send)
		client.conn.Close()
	}

	// Unregister from action manager
	s.actionManager.UnregisterObserver(s.actionChan)
}
