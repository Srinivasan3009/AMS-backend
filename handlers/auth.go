package handlers

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"AMS-backend/db"
	"AMS-backend/models"
	"AMS-backend/utils"
)

func Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "identifier and password are required"})
	}

	// Single identifier field is matched against all three role columns.
	// Whichever column matches determines the user's role.
	query := `
		SELECT id, role, name, password_hash
		FROM users
		WHERE register_no = $1 OR faculty_id = $1 OR username = $1
	`

	var user models.User
	row := db.Pool.QueryRow(context.Background(), query, identifier)
	err := row.Scan(&user.ID, &user.Role, &user.Name, &user.PasswordHash)

	if err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "server error"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate token"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		Secure:   false, // set true in production (HTTPS only)
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(models.LoginResponse{
		Message: "login successful",
		Role:    user.Role,
		Name:    user.Name,
	})
}

func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		HTTPOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})
	return c.JSON(fiber.Map{"message": "logged out"})
}

func Me(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"user_id": c.Locals("user_id"),
		"role":    c.Locals("role"),
	})
}
