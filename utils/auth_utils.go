package utils

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// GetAuthenticatedUserID ดึง userID จาก context และตรวจสอบการเข้าสู่ระบบ
func GetAuthenticatedUserID(c *fiber.Ctx) (uint, int, error) {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return 0, fiber.StatusUnauthorized, errors.New("You must be logged in")
	}
	return userID, fiber.StatusOK, nil
}


// ParseIDParam แปลง ID จาก parameter
func ParseIDParam(c *fiber.Ctx, paramName string) (uint, int, error) {
	id, err := strconv.ParseUint(c.Params(paramName), 10, 32)
	if err != nil {
		return 0, fiber.StatusBadRequest, errors.New("Invalid ID parameter")
	}
	return uint(id), fiber.StatusOK, nil
}

