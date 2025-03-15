package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/patiphanak/league-of-quiz/websocket"
)

// SetupWebSocketRoute ลงทะเบียน WebSocket route
func SetupWebSocketRoute(app *fiber.App, wsManager *websocket.Manager) {
	// แปลง http.Handler เป็น fiber.Handler ด้วย adaptor
	socketHandler := adaptor.HTTPHandler(wsManager.Server())

	// ลงทะเบียน WebSocket routes
	app.Get("/socket.io/*", socketHandler)
	app.Post("/socket.io/*", socketHandler)
}
