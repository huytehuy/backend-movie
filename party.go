package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	rooms      = make(map[string]*Room)
	roomsMutex sync.RWMutex
	upgrader   = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}
)

// Run starts the room's message handling loop
func (room *Room) Run() {
	for {
		select {
		case client := <-room.Register:
			room.Clients[client] = true
			log.Printf("Client %s joined room %s", client.Username, room.ID)

			// Send current video state to new client
			room.sendVideoStateToClient(client)

			// Broadcast user list update
			room.broadcastUserList()

		case client := <-room.Unregister:
			if _, ok := room.Clients[client]; ok {
				delete(room.Clients, client)
				close(client.Send)
				log.Printf("Client %s left room %s", client.Username, room.ID)

				// Broadcast user list update
				room.broadcastUserList()
			}

		case message := <-room.Broadcast:
			for client := range room.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(room.Clients, client)
				}
			}
		}
	}
}

func (room *Room) sendVideoStateToClient(client *Client) {
	syncMsg := Message{
		Type:      MessageTypeSync,
		RoomID:    room.ID,
		Data:      mustMarshal(room.VideoState),
		Timestamp: time.Now(),
	}

	msgBytes := mustMarshal(syncMsg)
	select {
	case client.Send <- msgBytes:
	default:
	}
}

func (room *Room) broadcastUserList() {
	users := make([]UserInfo, 0, len(room.Clients))
	for client := range room.Clients {
		users = append(users, UserInfo{
			ID:       client.ID,
			Username: client.Username,
		})
	}

	msg := Message{
		Type:      MessageTypeUserList,
		RoomID:    room.ID,
		Data:      mustMarshal(users),
		Timestamp: time.Now(),
	}

	room.Broadcast <- mustMarshal(msg)
}

// CreateRoom creates a new watch party room
func CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	roomID := uuid.New().String()[:8]
	userID := uuid.New().String()[:8]

	room := &Room{
		ID:      roomID,
		MovieID: req.MovieID,
		HostID:  userID,
		Name:    req.RoomName,
		Clients: make(map[*Client]bool),
		VideoState: &VideoState{
			IsPlaying:   false,
			CurrentTime: 0,
			UpdatedAt:   time.Now(),
		},
		CreatedAt:  time.Now(),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	roomsMutex.Lock()
	rooms[roomID] = room
	roomsMutex.Unlock()

	// Start room goroutine
	go room.Run()

	log.Printf("Room created: %s for movie %s by %s", roomID, req.MovieID, req.Username)

	resp := CreateRoomResponse{
		Room: &RoomInfo{
			ID:         room.ID,
			MovieID:    room.MovieID,
			Name:       room.Name,
			HostID:     room.HostID,
			UserCount:  0,
			VideoState: room.VideoState,
			CreatedAt:  room.CreatedAt,
		},
		UserID: userID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetRoom returns room information
func GetRoom(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomID := params["id"]

	roomsMutex.RLock()
	room, exists := rooms[roomID]
	roomsMutex.RUnlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	roomInfo := RoomInfo{
		ID:         room.ID,
		MovieID:    room.MovieID,
		Name:       room.Name,
		HostID:     room.HostID,
		UserCount:  len(room.Clients),
		VideoState: room.VideoState,
		CreatedAt:  room.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomInfo)
}

// HandleWebSocket handles WebSocket connections for watch party
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	roomID := params["id"]
	username := r.URL.Query().Get("username")

	if username == "" {
		username = "Anonymous"
	}

	roomsMutex.RLock()
	room, exists := rooms[roomID]
	roomsMutex.RUnlock()

	if !exists {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:       uuid.New().String()[:8],
		Username: username,
		Room:     room,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	room.Register <- client

	// Start goroutines for this client
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.(*websocket.Conn).Close()
	}()

	conn := c.Conn.(*websocket.Conn)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// writePump pumps messages from hub to WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.(*websocket.Conn).Close()
	}()

	conn := c.Conn.(*websocket.Conn)

	for {
		select {
		case message, ok := <-c.Send:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(messageBytes []byte) {
	var msg Message
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		return
	}

	msg.UserID = c.ID
	msg.Username = c.Username
	msg.Timestamp = time.Now()

	switch msg.Type {
	case MessageTypePlay:
		var data PlayPauseData
		json.Unmarshal(msg.Data, &data)

		c.Room.VideoState.IsPlaying = true
		c.Room.VideoState.CurrentTime = data.CurrentTime
		c.Room.VideoState.LastUpdateBy = c.Username
		c.Room.VideoState.UpdatedAt = time.Now()

		c.Room.Broadcast <- mustMarshal(msg)
		log.Printf("Room %s: %s played at %.2f", c.Room.ID, c.Username, data.CurrentTime)

	case MessageTypePause:
		var data PlayPauseData
		json.Unmarshal(msg.Data, &data)

		c.Room.VideoState.IsPlaying = false
		c.Room.VideoState.CurrentTime = data.CurrentTime
		c.Room.VideoState.LastUpdateBy = c.Username
		c.Room.VideoState.UpdatedAt = time.Now()

		c.Room.Broadcast <- mustMarshal(msg)
		log.Printf("Room %s: %s paused at %.2f", c.Room.ID, c.Username, data.CurrentTime)

	case MessageTypeSeek:
		var data SeekData
		json.Unmarshal(msg.Data, &data)

		c.Room.VideoState.CurrentTime = data.Time
		c.Room.VideoState.LastUpdateBy = c.Username
		c.Room.VideoState.UpdatedAt = time.Now()

		c.Room.Broadcast <- mustMarshal(msg)
		log.Printf("Room %s: %s seeked to %.2f", c.Room.ID, c.Username, data.Time)

	case MessageTypeChat:
		c.Room.Broadcast <- mustMarshal(msg)
		log.Printf("Room %s: %s: %s", c.Room.ID, c.Username, string(msg.Data))

	// WebRTC signaling - targeted messages
	case MessageTypeOffer, MessageTypeAnswer, MessageTypeIceCandidate:
		c.sendToClient(msg)
		log.Printf("Room %s: WebRTC %s from %s to %s", c.Room.ID, msg.Type, c.Username, msg.To)

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// sendToClient sends a message to a specific client in the room
func (c *Client) sendToClient(msg Message) {
	if msg.To == "" {
		log.Printf("Warning: targeted message without 'to' field")
		return
	}

	// Find the target client
	for client := range c.Room.Clients {
		if client.ID == msg.To {
			msgBytes := mustMarshal(msg)
			select {
			case client.Send <- msgBytes:
				log.Printf("Sent %s message to client %s", msg.Type, msg.To)
			default:
				log.Printf("Failed to send message to client %s (channel full)", msg.To)
			}
			return
		}
	}
	log.Printf("Target client %s not found in room %s", msg.To, c.Room.ID)
}

// Helper function to marshal JSON
func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("Marshal error: %v", err)
		return []byte("{}")
	}
	return b
}
