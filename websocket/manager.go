package websocket

import (
	"encoding/json"
	"log"
	"sync"

	socketio "github.com/googollee/go-socket.io"
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

// Manager จัดการการเชื่อมต่อ WebSocket
type Manager struct {
	server      *socketio.Server
	gameService *services.GameService
	sessions    map[string]map[string]socketio.Conn // sessionID -> socketID -> conn
	mu          sync.RWMutex
}

// NewManager สร้าง WebSocket manager ใหม่
func NewManager(gameService *services.GameService) (*Manager, error) {
	// สร้าง Socket.IO server
	log.Println("Starting WebSocket server...")
	server := socketio.NewServer(nil)

	manager := &Manager{
		server:      server,
		gameService: gameService,
		sessions:    make(map[string]map[string]socketio.Conn),
	}

	// ตั้งค่า event handlers
	server.OnConnect("/", func(s socketio.Conn) error {
		log.Printf("เชื่อมต่อแล้ว: %s", s.ID())
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("ยกเลิกการเชื่อมต่อแล้ว: %s (เหตุผล: %s)", s.ID(), reason)
		manager.handleDisconnect(s)
	})

	// จัดการการเข้าร่วมเกม
	server.OnEvent("/", "join_session", func(s socketio.Conn, sessionID string, userID uint, nickname string) {
		manager.handleJoinSession(s, sessionID, userID, nickname)
	})

	// จัดการการเริ่มเกม
	server.OnEvent("/", "start_game", func(s socketio.Conn, sessionID string, hostID uint) {
		manager.handleStartGame(s, sessionID, hostID)
	})

	// จัดการการส่งคำตอบ
	server.OnEvent("/", "submit_answer", func(s socketio.Conn, sessionID string, userID uint, questionID uint, choiceID uint, timeSpent float64) {
		manager.handleSubmitAnswer(s, sessionID, userID, questionID, choiceID, timeSpent)
	})

	// จัดการการจบเกม
	server.OnEvent("/", "end_game", func(s socketio.Conn, sessionID string, hostID uint) {
		manager.handleEndGame(s, sessionID, hostID)
	})

	// จัดการข้อความแชท
	server.OnEvent("/", "chat_message", func(s socketio.Conn, sessionID string, userID uint, message string) {
		manager.handleChatMessage(s, sessionID, userID, message)
	})

	return manager, nil
}

// Server คืนค่า socket.io server
func (m *Manager) Server() *socketio.Server {
	return m.server
}

// BroadcastToSession ส่งข้อความไปยังผู้เล่นทั้งหมดในห้อง
func (m *Manager) BroadcastToSession(sessionID string, message Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// ตรวจสอบว่ามีห้องนี้หรือไม่
	sessionConns, exists := m.sessions[sessionID]
	if !exists {
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
		conn.Emit("message", string(messageJSON))
	}
}

// HandleError ส่งข้อความข้อผิดพลาดไปยังผู้เล่น
func (m *Manager) HandleError(conn socketio.Conn, errorMsg string) {
	conn.Emit("error", errorMsg)
}

// handleJoinSession จัดการการเข้าร่วมเกมของผู้เล่น
func (m *Manager) handleJoinSession(s socketio.Conn, sessionID string, userID uint, nickname string) {
	// เพิ่มผู้เล่นลงในห้อง
	m.mu.Lock()
	if _, exists := m.sessions[sessionID]; !exists {
		m.sessions[sessionID] = make(map[string]socketio.Conn)
	}
	m.sessions[sessionID][s.ID()] = s
	m.mu.Unlock()

	// เข้าร่วมเกมผ่าน service
	player, err := m.gameService.JoinGameSession(sessionID, userID, nickname)
	if err != nil {
		m.HandleError(s, "ไม่สามารถเข้าร่วมเกมได้: "+err.Error())
		return
	}

	// ดึงข้อมูลเกม
	session, _ := m.gameService.GetGameSession(sessionID)

	// ส่งข้อความสำเร็จไปยังผู้เล่น
	s.Emit("joined", player)

	// ส่งข้อความไปยังผู้เล่นทั้งหมด
	m.BroadcastToSession(sessionID, Message{
		Type: EventPlayerJoined,
		Payload: map[string]interface{}{
			"session": session,
			"player":  player,
		},
	})
}

// handleStartGame จัดการการเริ่มเกม
func (m *Manager) handleStartGame(s socketio.Conn, sessionID string, hostID uint) {
	err := m.gameService.StartGameSession(sessionID, hostID)
	if err != nil {
		m.HandleError(s, "ไม่สามารถเริ่มเกมได้: "+err.Error())
		return
	}

	// ดึงข้อมูลเกมล่าสุด
	session, _ := m.gameService.GetGameSession(sessionID)

	// ส่งข้อความไปยังผู้เล่นทั้งหมดว่าเกมเริ่มแล้ว
	m.BroadcastToSession(sessionID, Message{
		Type: EventGameStarted,
		Payload: map[string]interface{}{
			"session": session,
		},
	})
}

// handleSubmitAnswer จัดการการส่งคำตอบของผู้เล่น
func (m *Manager) handleSubmitAnswer(s socketio.Conn, sessionID string, userID uint, questionID uint, choiceID uint, timeSpent float64) {
	answer, err := m.gameService.SubmitAnswer(sessionID, userID, questionID, choiceID, timeSpent)
	if err != nil {
		m.HandleError(s, "ไม่สามารถส่งคำตอบได้: "+err.Error())
		return
	}

	// ส่งข้อความสำเร็จไปยังผู้เล่น
	s.Emit("answer_submitted", answer)

	// ส่งข้อความไปยังผู้เล่นทั้งหมดว่ามีผู้เล่นส่งคำตอบแล้ว
	m.BroadcastToSession(sessionID, Message{
		Type: EventAnswerSubmitted,
		Payload: map[string]interface{}{
			"playerID":   userID,
			"questionID": questionID,
			// ไม่ควรส่งคำตอบที่แท้จริงเพื่อป้องกันการโกง
		},
	})
}

// handleEndGame จัดการการจบเกม
func (m *Manager) handleEndGame(s socketio.Conn, sessionID string, hostID uint) {
	session, err := m.gameService.EndGameSession(sessionID, hostID)
	if err != nil {
		m.HandleError(s, "ไม่สามารถจบเกมได้: "+err.Error())
		return
	}

	// ดึงรายชื่อผู้เล่น
	players, _ := m.gameService.GetPlayersBySessionID(sessionID)

	// ส่งข้อความไปยังผู้เล่นทั้งหมดว่าเกมจบแล้ว
	m.BroadcastToSession(sessionID, Message{
		Type: EventGameEnded,
		Payload: map[string]interface{}{
			"session": session,
			"players": players,
		},
	})
}

// handleChatMessage จัดการข้อความแชท
func (m *Manager) handleChatMessage(_ socketio.Conn, sessionID string, userID uint, message string) {
	// ส่งข้อความไปยังผู้เล่นทั้งหมดในห้อง
	m.BroadcastToSession(sessionID, Message{
		Type: EventChatMessage,
		Payload: map[string]interface{}{
			"userID":  userID,
			"message": message,
		},
	})
}

// handleDisconnect จัดการการยกเลิกการเชื่อมต่อของผู้เล่น
func (m *Manager) handleDisconnect(s socketio.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// หาว่าผู้เล่นอยู่ในห้องไหน
	for sessionID, clients := range m.sessions {
		if _, exists := clients[s.ID()]; exists {
			// ลบผู้เล่นออกจากห้อง
			delete(m.sessions[sessionID], s.ID())

			// ถ้าห้องว่างแล้ว ให้ลบห้อง
			if len(m.sessions[sessionID]) == 0 {
				delete(m.sessions, sessionID)
			}

			// ไม่จำเป็นต้องตรวจสอบห้องอื่นต่อ
			break
		}
	}
}
