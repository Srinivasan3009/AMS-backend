package handlers

import (
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"

	"AMS-backend/db"
	"AMS-backend/models"
)

// RegisterCourse (student) - registers the logged-in student for a course.
// Validations, in order:
//  1. Registration window must currently be open.
//  2. The course must actually be one of the student's available courses
//     (matches their department + computed current semester + batch, and
//     is Core, or Elective only if semester == 7) - prevents registering
//     for arbitrary/unrelated courses via a crafted request.
//  3. Not already registered for this course this term (UNIQUE constraint
//     also enforces this at the DB level as a final safeguard).
func RegisterCourse(c *fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}
	userID := int(userIDFloat)

	var req models.RegisterCourseRequest
	if err := c.BodyParser(&req); err != nil || req.CourseNo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_no is required"})
	}

	var registerNo, department, batch string
	err := db.Pool.QueryRow(context.Background(), `
		SELECT sd.register_no, sd.department, sd.batch
		FROM student_details sd
		JOIN users u ON u.register_no = sd.register_no
		WHERE u.id = $1
	`, userID).Scan(&registerNo, &department, &batch)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student not found"})
	}

	var isOpen bool
	err = db.Pool.QueryRow(context.Background(), `
		SELECT NOW() BETWEEN start_datetime AND end_datetime
		FROM registration_windows
		WHERE feature_key = 'course_registration'
	`).Scan(&isOpen)
	if err != nil || !isOpen {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "course registration is not currently open"})
	}

	semester, err := computeCurrentSemester(batch)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not compute current semester"})
	}

	var courseExists bool
	err = db.Pool.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM courses
			WHERE course_no = $1 AND department = $2 AND semester = $3 AND batch = $4
			  AND (course_category = 'Core' OR ($3 >= 7 AND course_category = 'Elective'))
		)
	`, req.CourseNo, department, semester, batch).Scan(&courseExists)
	if err != nil || !courseExists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "this course is not available for you to register"})
	}

	term := computeCurrentTerm()

	_, err = db.Pool.Exec(context.Background(), `
		INSERT INTO course_registrations (register_no, course_no, term)
		VALUES ($1, $2, $3)
	`, registerNo, req.CourseNo, term)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "you are already registered for this course"})
		}
		log.Println("register course insert error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not register for course"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "registered successfully"})
}

// GetMyRegistrations (student) - returns course_no list the student has already
// registered for in the CURRENT term, so the frontend can disable/mark those
// courses instead of letting the student attempt (and fail) a duplicate registration.
func GetMyRegistrations(c *fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}
	userID := int(userIDFloat)

	var registerNo string
	err := db.Pool.QueryRow(context.Background(),
		`SELECT register_no FROM users WHERE id = $1`, userID,
	).Scan(&registerNo)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student not found"})
	}

	term := computeCurrentTerm()

	rows, err := db.Pool.Query(context.Background(), `
		SELECT course_no FROM course_registrations
		WHERE register_no = $1 AND term = $2
	`, registerNo, term)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch registrations"})
	}
	defer rows.Close()

	var courseNos []string
	for rows.Next() {
		var courseNo string
		if err := rows.Scan(&courseNo); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read registration row"})
		}
		courseNos = append(courseNos, courseNo)
	}

	return c.JSON(fiber.Map{"term": term, "registered_courses": courseNos})
}
