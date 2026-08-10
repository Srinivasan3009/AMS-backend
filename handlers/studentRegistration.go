package handlers

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"AMS-backend/db"
	"AMS-backend/models"
)

// computeCurrentSemester derives a student's current semester from their batch
// (e.g. "2025-2029") and today's date. All batches are assumed to start in July.
//
//	July of startYear   .. Dec of startYear    -> semester 1
//	Jan of startYear+1  .. June of startYear+1  -> semester 2
//	July of startYear+1 .. Dec of startYear+1  -> semester 3
//	... and so on
func computeCurrentSemester(batch string) (int, error) {
	parts := strings.Split(batch, "-")
	if len(parts) != 2 {
		return 0, fiber.NewError(fiber.StatusInternalServerError, "invalid batch format")
	}
	startYear, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	now := time.Now()
	yearsElapsed := now.Year() - startYear
	semester := yearsElapsed * 2
	if now.Month() >= time.July {
		semester++
	}

	if semester < 1 {
		semester = 1
	}
	if semester > 8 {
		semester = 8
	}
	return semester, nil
}

// GetAvailableCoursesForRegistration (student) - only returns data if the
// course_registration window is currently open. Computes the student's current
// semester from their own department + batch, and returns matching Core courses
// (plus Elective courses if semester == 7).
func GetAvailableCoursesForRegistration(c *fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid session"})
	}
	userID := int(userIDFloat)

	// Look up the student's register_no from users, then department+batch from student_details.
	var registerNo string
	err := db.Pool.QueryRow(context.Background(),
		`SELECT register_no FROM users WHERE id = $1`, userID,
	).Scan(&registerNo)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student not found"})
	}

	var department, batch string
	err = db.Pool.QueryRow(context.Background(),
		`SELECT department, batch FROM student_details WHERE register_no = $1`, registerNo,
	).Scan(&department, &batch)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student details not found"})
	}

	// Check the registration window is open.
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

	// Core courses always included; Electives only from semester 7 onward.
	query := `
		SELECT course_no, course_name, department, semester, batch, course_type, course_category,
		       lecture_hours, tutorial_hours, practical_hours, tcp, credit, active
		FROM courses
		WHERE department = $1 AND semester = $2 AND batch = $3
		  AND (course_category = 'Core' OR ($2 >= 7 AND course_category = 'Elective'))
		ORDER BY course_no
	`
	rows, err := db.Pool.Query(context.Background(), query, department, semester, batch)
	if err != nil {
		log.Println("available courses query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch available courses"})
	}
	defer rows.Close()

	var courseList []models.Course
	for rows.Next() {
		var course models.Course
		if err := rows.Scan(&course.CourseNo, &course.CourseName, &course.Department, &course.Semester, &course.Batch,
			&course.CourseType, &course.CourseCategory, &course.LectureHours, &course.TutorialHours,
			&course.PracticalHours, &course.TCP, &course.Credit, &course.Active); err != nil {
			log.Println("available courses scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read course row"})
		}
		courseList = append(courseList, course)
	}

	return c.JSON(fiber.Map{
		"semester": semester,
		"courses":  courseList,
	})
}
