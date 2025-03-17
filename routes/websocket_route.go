package routes

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/patiphanak/league-of-quiz/websocket"
)

func SetupWebSocketRoute(app *fiber.App, wsManager *websocket.Manager) {
	// แปลง http.Handler เป็น fiber.Handler ด้วย adaptor
	log.Println("Setting up WebSocket routes")
	socketHandler := adaptor.HTTPHandler(wsManager.Server())

	// ลงทะเบียน WebSocket routes
	app.Get("/socket.io/*", socketHandler)
	app.Post("/socket.io/*", socketHandler)
}
