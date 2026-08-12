package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"

	"AMS-backend/db"
	"AMS-backend/models"
)

// UpsertAssignment creates or updates the faculty assigned to a course for a given term.

func UpsertAssignment(c *fiber.Ctx) error {
	var req models.AssignmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.CourseNo == "" || req.FacultyID == "" || req.Term == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_no, faculty_id and term are required"})
	}

	_, err := db.Pool.Exec(context.Background(), `
		INSERT INTO faculty_course_assignments (course_no, faculty_id, term, department, semester, batch)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (course_no, term)
		DO UPDATE SET faculty_id = EXCLUDED.faculty_id, updated_at = NOW()
	`, req.CourseNo, req.FacultyID, req.Term, req.Department, req.Semester, req.Batch)

	if err != nil {
		log.Println("assignment upsert error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not save assignment"})
	}

	return c.JSON(fiber.Map{"message": "faculty assigned"})
}

// ListAssignments returns assignments filtered by term + department + semester (+ optional batch)
func ListAssignments(c *fiber.Ctx) error {
	term := c.Query("term")
	department := c.Query("department")
	semester := c.Query("semester")

	if term == "" || department == "" || semester == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "term, department and semester are required"})
	}

	rows, err := db.Pool.Query(context.Background(), `
		SELECT a.course_no, a.faculty_id, f.name, a.term, a.department, a.semester, a.batch
		FROM faculty_course_assignments a
		JOIN faculty_details f ON f.faculty_id = a.faculty_id
		WHERE a.term = $1 AND a.department = $2 AND a.semester = $3
	`, term, department, semester)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch assignments"})
	}
	defer rows.Close()

	var assignments []models.Assignment
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(&a.CourseNo, &a.FacultyID, &a.FacultyName, &a.Term, &a.Department, &a.Semester, &a.Batch); err != nil {
			log.Println("assignment row scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read assignment row"})
		}
		assignments = append(assignments, a)
	}

	return c.JSON(assignments)
}
