package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/patiphanak/league-of-quiz/services"
)

// EventType กำหนดประเภทของเหตุการณ์ WebSocket
type EventType string

// กำหนดประเภทของเหตุการณ์
const (
	EventPlayerJoined    EventType = "player_joined"
	EventGameStarted     EventType = "game_started"
	EventQuestionStarted EventType = "question_started"
	EventAnswerSubmitted EventType = "answer_submitted"
	EventQuestionEnded   EventType = "question_ended"
	EventGameEnded       EventType = "game_ended"
	EventChatMessage     EventType = "chat_message"
)

// Message แทนข้อความ WebSocket
type Message struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// ClientMessage แทนข้อความที่รับจาก client
type ClientMessage struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// JoinSessionPayload แทนข้อมูลที่ใช้ในการเข้าร่วมเกม
type JoinSessionPayload struct {
	SessionID string `json:"sessionId"`
	UserID    uint   `json:"userId"`
	Nickname  string `json:"nickname"`
}

// GameActionPayload แทนข้อมูลที่ใช้ในการควบคุมเกม
type GameActionPayload struct {
	SessionID string `json:"sessionId"`
	UserID    uint   `json:"userId"`
}

// SubmitAnswerPayload แทนข้อมูลที่ใช้ในการส่งคำตอบ
type SubmitAnswerPayload struct {
	SessionID  string  `json:"sessionId"`
	UserID     uint    `json:"userId"`
	QuestionID uint    `json:"questionId"`
	ChoiceID   uint    `json:"choiceId"`
	TimeSpent  float64 `json:"timeSpent"`
}

// ChatMessagePayload แทนข้อมูลข้อความแชท
type ChatMessagePayload struct {
	SessionID string `json:"sessionId"`
	UserID    uint   `json:"userId"`
	Message   string `json:"message"`
}

// ErrorResponse แทนข้อความแจ้งข้อผิดพลาด
type ErrorResponse struct {
	Error string `json:"error"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Manager จัดการการเชื่อมต่อ WebSocket
type Manager struct {
	gameService *services.GameService
	sessions    map[string]map[string]*websocket.Conn // sessionID -> connID -> conn
	users       map[string]uint                      // connID -> userID
	mu          sync.RWMutex
}

// NewManager สร้าง WebSocket manager ใหม่
func NewManager(gameService *services.GameService) (*Manager, error) {
	if gameService == nil {
		return nil, fmt.Errorf("gameService cannot be nil")
	}
	
	return &Manager{
		gameService: gameService,
		sessions:    make(map[string]map[string]*websocket.Conn),
		users:       make(map[string]uint),
	}, nil
}

// generateConnID สร้าง ID ที่ไม่ซ้ำกันสำหรับการเชื่อมต่อ
func generateConnID() string {
	return uuid.New().String()
}

// HandleHTTP คืนค่า http.Handler สำหรับการจัดการ WebSocket
func (m *Manager) HandleHTTP() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.HandleWebSocket(w, r)
	})
}

// HandleWebSocket จัดการการเชื่อมต่อ WebSocket ใหม่
func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	
	connID := generateConnID()
	log.Printf("Client connected: %s", connID)
	
	// เริ่มต้นอ่านข้อความจากการเชื่อมต่อ
	go m.readPump(conn, connID)
}

// readPump อ่านข้อความจากการเชื่อมต่อ WebSocket และจัดการข้อความ
func (m *Manager) readPump(conn *websocket.Conn, connID string) {
	defer func() {
		m.handleDisconnect(connID)
		conn.Close()
		log.Printf("Connection closed: %s", connID)
	}()
	
	// ตั้งค่าพารามิเตอร์การเชื่อมต่อ
	conn.SetReadLimit(4096) // 4KB max message size
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		// อ่านข้อความจากการเชื่อมต่อ
		_, rawMessage, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, 
				websocket.CloseGoingAway, 
				websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}
		
		// แปลงข้อความเป็น ClientMessage
		var message ClientMessage
		err = json.Unmarshal(rawMessage, &message)
		if err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			m.sendError(conn, "Invalid message format")
			continue
		}
		
		// จัดการข้อความตาม action
		switch message.Action {
		case "join_session":
			var payload JoinSessionPayload
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				m.sendError(conn, "Invalid join_session payload")
				continue
			}
			m.handleJoinSession(conn, connID, payload.SessionID, payload.UserID, payload.Nickname)
			
		case "start_game":
			var payload GameActionPayload
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				m.sendError(conn, "Invalid start_game payload")
				continue
			}
			m.handleStartGame(conn, connID, payload.SessionID, payload.UserID)
			
		case "submit_answer":
			var payload SubmitAnswerPayload
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				m.sendError(conn, "Invalid submit_answer payload")
				continue
			}
			m.handleSubmitAnswer(conn, connID, payload.SessionID, payload.UserID, 
				payload.QuestionID, payload.ChoiceID, payload.TimeSpent)
			
		case "end_game":
			var payload GameActionPayload
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				m.sendError(conn, "Invalid end_game payload")
				continue
			}
			m.handleEndGame(conn, connID, payload.SessionID, payload.UserID)
			
		case "chat_message":
			var payload ChatMessagePayload
			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				m.sendError(conn, "Invalid chat_message payload")
				continue
			}
			m.handleChatMessage(conn, connID, payload.SessionID, payload.UserID, payload.Message)
			
		default:
			m.sendError(conn, "Unknown action: "+message.Action)
		}
	}
}

// sendMessage ส่งข้อความไปยังการเชื่อมต่อ WebSocket
func (m *Manager) sendMessage(conn *websocket.Conn, message interface{}) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return err
	}
	
	return conn.WriteMessage(websocket.TextMessage, messageJSON)
}

// sendError ส่งข้อความข้อผิดพลาดไปยังการเชื่อมต่อ WebSocket
func (m *Manager) sendError(conn *websocket.Conn, errorMsg string) {
	log.Printf("Sending error to client: %s", errorMsg)
	m.sendMessage(conn, ErrorResponse{Error: errorMsg})
}

// BroadcastToSession ส่งข้อความไปยังผู้เล่นทั้งหมดในห้อง
func (m *Manager) BroadcastToSession(sessionID string, message Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// ตรวจสอบว่ามีห้องนี้หรือไม่
	sessionConns, exists := m.sessions[sessionID]
	if !exists {
		log.Printf("Cannot broadcast to session %s: session not found", sessionID)
		return
	}
	
	// แปลงข้อความเป็น JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	
	// ส่งข้อความไปยังผู้เล่นทั้งหมดในห้อง
	for _, conn := range sessionConns {
		if err := conn.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

// BroadcastToSessionExcept ส่งข้อความไปยังผู้เล่นทั้งหมดในห้องยกเว้นผู้เล่นที่ระบุ
func (m *Manager) BroadcastToSessionExcept(sessionID string, exceptConnID string, message Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// ตรวจสอบว่ามีห้องนี้หรือไม่
	sessionConns, exists := m.sessions[sessionID]
	if !exists {
		log.Printf("Cannot broadcast to session %s: session not found", sessionID)
		return
	}
	
	// แปลงข้อความเป็น JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	
	// ส่งข้อความไปยังผู้เล่นทั้งหมดในห้องยกเว้นผู้เล่นที่ระบุ
	for connID, conn := range sessionConns {
		if connID != exceptConnID {
			if err := conn.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}

// handleJoinSession จัดการการเข้าร่วมเกมของผู้เล่น
func (m *Manager) handleJoinSession(conn *websocket.Conn, connID string, sessionID string, userID uint, nickname string) {
	// เพิ่มผู้เล่นลงในห้อง
	m.mu.Lock()
	if _, exists := m.sessions[sessionID]; !exists {
		m.sessions[sessionID] = make(map[string]*websocket.Conn)
	}
	m.sessions[sessionID][connID] = conn
	m.users[connID] = userID
	m.mu.Unlock()
	
	// เข้าร่วมเกมผ่าน service
	player, err := m.gameService.JoinGameSession(sessionID, userID, nickname)
	if err != nil {
		m.sendError(conn, "Cannot join game: "+err.Error())
		return
	}
	
	// ดึงข้อมูลเกม
	session, _ := m.gameService.GetGameSession(sessionID)
	
	// ส่งข้อความสำเร็จไปยังผู้เล่น
	m.sendMessage(conn, map[string]interface{}{
		"type": "joined",
		"payload": map[string]interface{}{
			"session": session,
			"player":  player,
		},
	})
	
	// ส่งข้อความไปยังผู้เล่นทั้งหมด
	m.BroadcastToSession(sessionID, Message{
		Type: EventPlayerJoined,
		Payload: map[string]interface{}{
			"sessionId": sessionID,
			"player":    player,
		},
	})
}

// handleStartGame จัดการการเริ่มเกม
func (m *Manager) handleStartGame(conn *websocket.Conn, _ string, sessionID string, hostID uint) {
	err := m.gameService.StartGameSession(sessionID, hostID)
	if err != nil {
		m.sendError(conn, "Cannot start game: "+err.Error())
		return
	}
	
	// ดึงข้อมูลเกมล่าสุด
	session, _ := m.gameService.GetGameSession(sessionID)
	
	// ส่งข้อความไปยังผู้เล่นทั้งหมดว่าเกมเริ่มแล้ว
	m.BroadcastToSession(sessionID, Message{
		Type: EventGameStarted,
		Payload: map[string]interface{}{
			"sessionId": sessionID,
			"session":   session,
		},
	})
}

// handleSubmitAnswer จัดการการส่งคำตอบของผู้เล่น
func (m *Manager) handleSubmitAnswer(conn *websocket.Conn, _ string, sessionID string, userID uint, questionID uint, choiceID uint, timeSpent float64) {
	answer, err := m.gameService.SubmitAnswer(sessionID, userID, questionID, choiceID, timeSpent)
	if err != nil {
		m.sendError(conn, "Cannot submit answer: "+err.Error())
		return
	}
	
	// ส่งข้อความสำเร็จไปยังผู้เล่น
	m.sendMessage(conn, map[string]interface{}{
		"type": "answer_submitted",
		"payload": answer,
	})
	
	// ส่งข้อความไปยังผู้เล่นทั้งหมดว่ามีผู้เล่นส่งคำตอบแล้ว
	m.BroadcastToSession(sessionID, Message{
		Type: EventAnswerSubmitted,
		Payload: map[string]interface{}{
			"sessionId":  sessionID,
			"playerId":   userID,
			"questionId": questionID,
			// ไม่ควรส่งคำตอบที่แท้จริงเพื่อป้องกันการโกง
		},
	})
}

// handleEndGame จัดการการจบเกม
func (m *Manager) handleEndGame(conn *websocket.Conn, _ string, sessionID string, hostID uint) {
	session, err := m.gameService.EndGameSession(sessionID, hostID)
	if err != nil {
		m.sendError(conn, "Cannot end game: "+err.Error())
		return
	}
	
	// ดึงรายชื่อผู้เล่น
	players, _ := m.gameService.GetPlayersBySessionID(sessionID)
	
	// ส่งข้อความไปยังผู้เล่นทั้งหมดว่าเกมจบแล้ว
	m.BroadcastToSession(sessionID, Message{
		Type: EventGameEnded,
		Payload: map[string]interface{}{
			"sessionId": sessionID,
			"session":   session,
			"players":   players,
		},
	})
}

// handleChatMessage จัดการข้อความแชท
func (m *Manager) handleChatMessage(_ *websocket.Conn, _ string, sessionID string, userID uint, message string) {
	// ส่งข้อความไปยังผู้เล่นทั้งหมดในห้อง
	m.BroadcastToSession(sessionID, Message{
		Type: EventChatMessage,
		Payload: map[string]interface{}{
			"sessionId": sessionID,
			"userID":    userID,
			"message":   message,
		},
	})
}

// handleDisconnect จัดการการยกเลิกการเชื่อมต่อของผู้เล่น
func (m *Manager) handleDisconnect(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// หาว่าผู้เล่นอยู่ในห้องไหน
	var sessionIDToRemove string
	for sessionID, clients := range m.sessions {
		if _, exists := clients[connID]; exists {
			// พบห้องที่มีการเชื่อมต่อนี้
			sessionIDToRemove = sessionID
			break
		}
	}
	
	// ถ้าพบว่าผู้เล่นอยู่ในห้อง
	if sessionIDToRemove != "" {
		// ลบผู้เล่นออกจากห้อง
		delete(m.sessions[sessionIDToRemove], connID)
		
		// ถ้าห้องว่างแล้ว ให้ลบห้อง
		if len(m.sessions[sessionIDToRemove]) == 0 {
			delete(m.sessions, sessionIDToRemove)
		}
	}
	
	// ลบการเชื่อมโยงกับ userID
	delete(m.users, connID)
}