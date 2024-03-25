package main

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func isAuthenticated() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Get JWT token from the request cookies
		cookie := ctx.Cookies("jwt")
		if cookie == "" {
			// JWT token not found, redirect to login page
			return ctx.Redirect("/login")
		}

		// Parse JWT token

		token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method used
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			// Return the secret key used to sign the token
			return []byte("scrubPosts"), nil
		})

		// Check for parsing errors
		if err != nil {
			return ctx.Redirect("/login")
		}
		// Check if the token is valid
		if !token.Valid {
			return ctx.Redirect("/login")
		}
		// User is authenticated, continue to the next middleware or handler
		return ctx.Next()

	}
}
