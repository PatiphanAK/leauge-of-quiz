package middleware

import (
	"encoding/json"
	"regexp"

	"github.com/gofiber/fiber/v2"
)

// TransformResponse transforms the response data to add "localhost:3000" to image URLs
func TransformResponse() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Call next handler
		if err := c.Next(); err != nil {
			return err
		}

		// Get response body
		body := c.Response().Body()

		// Check if the response is JSON
		if !json.Valid(body) || len(body) == 0 {
			return nil
		}

		// Convert to string for easier manipulation
		bodyStr := string(body)

		// Regex to find image URLs
		// This will match URLs starting with /storage/ in JSON strings
		re := regexp.MustCompile(`"(\/storage\/[^"]+)"`)
		transformedBody := re.ReplaceAllString(bodyStr, `"localhost:3000$1"`)

		// Set transformed body back to response
		c.Response().SetBody([]byte(transformedBody))

		return nil
	}
}
