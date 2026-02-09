package main

import (
	"encoding/json"
	"time"
)

// Movie represents a movie entity
type Movie struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Thumbnail   string    `json:"thumbnail"`
	VideoURL    string    `json:"videoUrl"`
	Duration    int       `json:"duration"` // in seconds
	CreatedAt   time.Time `json:"createdAt"`
}

// Room represents a watch party room
type Room struct {
	ID         string           `json:"id"`
	MovieID    string           `json:"movieId"`
	HostID     string           `json:"hostId"`
	Name       string           `json:"name"`
	Clients    map[*Client]bool `json:"-"`
	VideoState *VideoState      `json:"videoState"`
	CreatedAt  time.Time        `json:"createdAt"`
	Broadcast  chan []byte      `json:"-"`
	Register   chan *Client     `json:"-"`
	Unregister chan *Client     `json:"-"`
}

// Client represents a connected user in a room
type Client struct {
	ID       string
	Username string
	Room     *Room
	Conn     interface{} // WebSocket connection
	Send     chan []byte
}

// VideoState represents the current state of video playback
type VideoState struct {
	IsPlaying    bool      `json:"isPlaying"`
	CurrentTime  float64   `json:"currentTime"`
	LastUpdateBy string    `json:"lastUpdateBy"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// WebSocket Message Types
const (
	MessageTypeJoin     = "join"
	MessageTypeLeave    = "leave"
	MessageTypePlay     = "play"
	MessageTypePause    = "pause"
	MessageTypeSeek     = "seek"
	MessageTypeSync     = "sync"
	MessageTypeChat     = "chat"
	MessageTypeUserList = "userList"
	MessageTypeError    = "error"
	// WebRTC signaling
	MessageTypeOffer        = "offer"
	MessageTypeAnswer       = "answer"
	MessageTypeIceCandidate = "ice-candidate"
)

// Message represents a WebSocket message
type Message struct {
	Type      string          `json:"type"`
	RoomID    string          `json:"roomId,omitempty"`
	UserID    string          `json:"userId,omitempty"`
	Username  string          `json:"username,omitempty"`
	To        string          `json:"to,omitempty"` // For targeted messages (WebRTC)
	Data      json.RawMessage `json:"data,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// PlayPauseData for play/pause events
type PlayPauseData struct {
	CurrentTime float64 `json:"currentTime"`
}

// SeekData for seek events
type SeekData struct {
	Time float64 `json:"time"`
}

// ChatData for chat messages
type ChatData struct {
	Message string `json:"message"`
}

// RoomInfo for room details
type RoomInfo struct {
	ID         string      `json:"id"`
	MovieID    string      `json:"movieId"`
	Name       string      `json:"name"`
	HostID     string      `json:"hostId"`
	UserCount  int         `json:"userCount"`
	VideoState *VideoState `json:"videoState"`
	CreatedAt  time.Time   `json:"createdAt"`
}

// UserInfo for user details
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// CreateRoomRequest for creating a new room
type CreateRoomRequest struct {
	MovieID  string `json:"movieId"`
	RoomName string `json:"roomName"`
	Username string `json:"username"`
}

// CreateRoomResponse for room creation response
type CreateRoomResponse struct {
	Room   *RoomInfo `json:"room"`
	UserID string    `json:"userId"`
}

// JoinRoomRequest for joining a room
type JoinRoomRequest struct {
	Username string `json:"username"`
}
