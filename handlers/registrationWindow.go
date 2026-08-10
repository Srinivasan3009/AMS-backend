package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"AMS-backend/db"
	"AMS-backend/models"
)

const courseRegistrationKey = "course_registration"

// SetRegistrationWindow (admin) - creates or overwrites the course_registration window.
// Combines separate date+time inputs into full timestamps.
func SetRegistrationWindow(c *fiber.Ctx) error {
	var req models.SetWindowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.StartDate == "" || req.StartTime == "" || req.EndDate == "" || req.EndTime == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "start date/time and end date/time are required"})
	}

	startDatetime := fmt.Sprintf("%s %s:00", req.StartDate, req.StartTime)
	endDatetime := fmt.Sprintf("%s %s:00", req.EndDate, req.EndTime)

	_, err := db.Pool.Exec(context.Background(), `
		INSERT INTO registration_windows (feature_key, start_datetime, end_datetime)
		VALUES ($1, $2, $3)
		ON CONFLICT (feature_key)
		DO UPDATE SET start_datetime = EXCLUDED.start_datetime, end_datetime = EXCLUDED.end_datetime, updated_at = NOW()
	`, courseRegistrationKey, startDatetime, endDatetime)

	if err != nil {
		log.Println("set registration window error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not save registration window"})
	}

	return c.JSON(fiber.Map{"message": "registration window saved"})
}

// GetRegistrationWindow (admin) - fetches the current setting, used to pre-fill the Menu Enable form.
func GetRegistrationWindow(c *fiber.Ctx) error {
	var w models.RegistrationWindow
	err := db.Pool.QueryRow(context.Background(), `
		SELECT feature_key, start_datetime::text, end_datetime::text
		FROM registration_windows
		WHERE feature_key = $1
	`, courseRegistrationKey).Scan(&w.FeatureKey, &w.StartDatetime, &w.EndDatetime)

	if err != nil {
		// No window set yet - not an error, just empty.
		return c.JSON(fiber.Map{"feature_key": courseRegistrationKey, "start_datetime": nil, "end_datetime": nil})
	}

	return c.JSON(w)
}

// GetRegistrationWindowStatus (student) - checks if the course_registration window is currently open.
func GetRegistrationWindowStatus(c *fiber.Ctx) error {
	var w models.RegistrationWindow
	err := db.Pool.QueryRow(context.Background(), `
		SELECT feature_key, start_datetime::text, end_datetime::text
		FROM registration_windows
		WHERE feature_key = $1
	`, courseRegistrationKey).Scan(&w.FeatureKey, &w.StartDatetime, &w.EndDatetime)

	if err != nil {
		return c.JSON(models.WindowStatusResponse{IsOpen: false})
	}

	var isOpen bool
	checkErr := db.Pool.QueryRow(context.Background(), `
		SELECT NOW() BETWEEN start_datetime AND end_datetime
		FROM registration_windows
		WHERE feature_key = $1
	`, courseRegistrationKey).Scan(&isOpen)

	if checkErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not check window status"})
	}

	return c.JSON(models.WindowStatusResponse{
		IsOpen:        isOpen,
		StartDatetime: w.StartDatetime,
		EndDatetime:   w.EndDatetime,
	})
}
