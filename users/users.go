package users

import (
	"github.com/gofiber/fiber/v2"
)

func SignUp(c *fiber.Ctx) error {
	return c.SendString("Sign up endpoint")
}
