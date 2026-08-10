package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"AMS-backend/db"
	"AMS-backend/models"
)

// CreateFaculty inserts into users + faculty_details in a single transaction.
func CreateFaculty(c *fiber.Ctx) error {
	var req models.FacultyCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.FacultyID == "" || req.Name == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "faculty_id, name and password are required"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
	}

	ctx := context.Background()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not start transaction"})
	}
	defer tx.Rollback(ctx) // no-op if committed

	_, err = tx.Exec(ctx,
		`INSERT INTO users (role, faculty_id, name, password_hash) VALUES ('faculty', $1, $2, $3)`,
		req.FacultyID, req.Name, string(hash),
	)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "faculty_id already exists or invalid data"})
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO faculty_details
			(faculty_id, name, date_of_birth, gender, designation, department, mobile_number, email, address_1, address_2, active, date_of_retirement)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		req.FacultyID, req.Name, req.DateOfBirth, req.Gender, req.Designation, req.Department,
		req.MobileNumber, req.Email, req.Address1, req.Address2, req.Active, req.DateOfRetirement,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create faculty details"})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not commit transaction"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "faculty created"})
}

// ListFaculty returns all faculty records for the admin table.
func ListFaculty(c *fiber.Ctx) error {
	rows, err := db.Pool.Query(context.Background(), `
		SELECT faculty_id, name, date_of_birth::text, gender, designation, department, mobile_number, email,
		       address_1, address_2, active, date_of_retirement::text
		FROM faculty_details
		ORDER BY name
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch faculty"})
	}
	defer rows.Close()

	var facultyList []models.Faculty
	for rows.Next() {
		var f models.Faculty
		if err := rows.Scan(&f.FacultyID, &f.Name, &f.DateOfBirth, &f.Gender, &f.Designation, &f.Department,
			&f.MobileNumber, &f.Email, &f.Address1, &f.Address2, &f.Active, &f.DateOfRetirement); err != nil {
			log.Println("faculty row scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read faculty row"})
		}
		facultyList = append(facultyList, f)
	}

	return c.JSON(facultyList)
}

// UpdateFaculty updates faculty_details always, and users.password_hash only if a new password was provided.
// faculty_id comes from the URL and is never modified.
func UpdateFaculty(c *fiber.Ctx) error {
	facultyID := c.Params("faculty_id")
	if facultyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "faculty_id is required in URL"})
	}

	var req models.FacultyUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ctx := context.Background()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not start transaction"})
	}
	defer tx.Rollback(ctx)

	cmdTag, err := tx.Exec(ctx, `
		UPDATE faculty_details
		SET name=$1, date_of_birth=$2, gender=$3, designation=$4, department=$5, mobile_number=$6,
		    email=$7, address_1=$8, address_2=$9, active=$10, date_of_retirement=$11, updated_at=NOW()
		WHERE faculty_id=$12
	`, req.Name, req.DateOfBirth, req.Gender, req.Designation, req.Department, req.MobileNumber,
		req.Email, req.Address1, req.Address2, req.Active, req.DateOfRetirement, facultyID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update faculty details"})
	}
	if cmdTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	// Also keep users.name in sync
	_, err = tx.Exec(ctx, `UPDATE users SET name=$1 WHERE faculty_id=$2`, req.Name, facultyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update user record"})
	}

	// Only touch password if a new one was actually entered
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
		}
		_, err = tx.Exec(ctx, `UPDATE users SET password_hash=$1 WHERE faculty_id=$2`, string(hash), facultyID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update password"})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "faculty updated"})
}
