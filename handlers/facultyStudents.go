package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"

	"AMS-backend/db"
	"AMS-backend/models"
)

// GetStudentsForCourse returns students who ACTUALLY REGISTERED for a specific
// course in a given term - not a department+batch guess. Joins course_registrations
// so only genuine enrollees show up in the roster.
func GetStudentsForCourse(c *fiber.Ctx) error {
	courseNo := c.Query("course_no")
	term := c.Query("term")

	if courseNo == "" || term == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_no and term are required"})
	}

	rows, err := db.Pool.Query(context.Background(), `
		SELECT sd.register_no, sd.name, sd.department, sd.batch, sd.email
		FROM course_registrations cr
		JOIN student_details sd ON sd.register_no = cr.register_no
		WHERE cr.course_no = $1 AND cr.term = $2
		ORDER BY sd.name
	`, courseNo, term)
	if err != nil {
		log.Println("students for course query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch course students"})
	}
	defer rows.Close()

	var students []models.CourseStudent
	for rows.Next() {
		var s models.CourseStudent
		if err := rows.Scan(&s.RegisterNo, &s.Name, &s.Department, &s.Batch, &s.Email); err != nil {
			log.Println("students for course scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read student row"})
		}
		students = append(students, s)
	}

	return c.JSON(students)
}
