package users

import (
	"database/sql"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Define a struct to represent the claims in the JWT
type Payload struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type Users struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// function that renders a signup page template
func SignUpPage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("signUp", fiber.Map{})
	}
}

func SignUp(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "POST" {
			// process signup form
			email := c.FormValue("email")
			password := c.FormValue("password")

			// check if the user already exists
			var userCount int

			row := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", email)
			if err := row.Scan(&userCount); err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Error checking if user exists")
			}

			if userCount > 0 {
				return c.Status(fiber.StatusBadRequest).SendString("User already exists")

			}

			// check data isn't empty on signing up form
			if email == "" || password == "" {
				return c.Status(fiber.StatusBadRequest).SendString("Email and password are required")
			}

			// hash the password for security.
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

			if err != nil {
				return c.Status(fiber.StatusBadRequest).SendString("Error hashing the password")
			}

			// insert the user into the database
			_, err = db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", email, string(hashedPassword))
			if err != nil {
				c.Status(fiber.StatusBadRequest).SendString("Error creating user and inserting into db.")
			}

			// Redirect to the main page after successful signup
			return c.Redirect("/viewPost", fiber.StatusSeeOther)
		}

		// render signup form
		return c.Render("signUp", fiber.Map{})
	}
}

// function to generate a JWT
func generateJWT(userId string) (string, error) {
	// Define the expiration time for the token
	expirationTime := time.Now().Add(24 * time.Hour)
	// Define the claims for the JWT
	payload := &Payload{
		UserID: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		}}
	// Create the token with the payload
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	// Sign the token with a secret key
	tokenString, err := token.SignedString([]byte("scrubPosts"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// function that renders a login page template
func Login(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{})
}

func LoginSubmit(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		if c.Method() == "POST" {
			email := c.FormValue("email")
			password := c.FormValue("password")

			// check data isn't empty on signing up form
			if email == "" || password == "" {
				return c.Status(fiber.StatusBadRequest).SendString("Email and password are required")
			}

			// Query the database to get the user
			var storedPassword string
			err := db.QueryRow("SELECT password FROM users WHERE email =$1", email).Scan(&storedPassword)
			if err != nil {
				if err == sql.ErrNoRows {
					return c.Status(fiber.StatusUnauthorized).SendString("Invalid email or password during login")
				}
				return err
			}

			// Generate JWT token
			token, err := generateJWT(email)
			if err != nil {
				return err
			}
			// Set JWT token in cookie
			c.Cookie(&fiber.Cookie{
				Name:     "jwt",
				Value:    token,
				Expires:  time.Now().Add(24 * time.Hour),
				HTTPOnly: true,
			})
			// Redirect to homepage successful login
			return c.Redirect("/viewPost")
		}
		// Render login page
		return c.Render("login", fiber.Map{})
	}
}
