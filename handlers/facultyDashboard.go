package handlers

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"AMS-backend/db"
	"AMS-backend/models"
)

// computeCurrentTerm returns the current teaching term as a string like "July 2026" or "Jan 2027",
// matching the format used when assignments are saved (Assign Faculty to Course).
func computeCurrentTerm() string {
	now := time.Now()
	if now.Month() >= time.July {
		return "July " + strconv.Itoa(now.Year())
	}
	return "Jan " + strconv.Itoa(now.Year())
}

// getFacultyIDFromSession resolves the logged-in user's faculty_id from their user_id JWT claim.
func getFacultyIDFromSession(c *fiber.Ctx) (string, error) {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return "", fiber.NewError(fiber.StatusUnauthorized, "invalid session")
	}
	userID := int(userIDFloat)

	var facultyID string
	err := db.Pool.QueryRow(context.Background(),
		`SELECT faculty_id FROM users WHERE id = $1`, userID,
	).Scan(&facultyID)
	return facultyID, err
}

// GetFacultyDetails returns the logged-in faculty member's own profile.
func GetFacultyDetails(c *fiber.Ctx) error {
	facultyID, err := getFacultyIDFromSession(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	var f models.Faculty
	err = db.Pool.QueryRow(context.Background(), `
		SELECT faculty_id, name, date_of_birth::text, gender, designation, department,
		       mobile_number, email, address_1, address_2, active, date_of_retirement::text
		FROM faculty_details
		WHERE faculty_id = $1
	`, facultyID).Scan(&f.FacultyID, &f.Name, &f.DateOfBirth, &f.Gender, &f.Designation, &f.Department,
		&f.MobileNumber, &f.Email, &f.Address1, &f.Address2, &f.Active, &f.DateOfRetirement)

	if err != nil {
		log.Println("faculty details fetch error:", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty details not found"})
	}

	return c.JSON(f)
}

// GetFacultyCurrentCourses returns the courses this faculty is assigned to teach in the current term.
func GetFacultyCurrentCourses(c *fiber.Ctx) error {
	facultyID, err := getFacultyIDFromSession(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	currentTerm := computeCurrentTerm()

	rows, err := db.Pool.Query(context.Background(), `
		SELECT a.term, a.department, a.semester, a.batch,
		       co.course_no, co.course_name, co.course_type, co.course_category, co.credit
		FROM faculty_course_assignments a
		JOIN courses co ON co.course_no = a.course_no
		WHERE a.faculty_id = $1 AND a.term = $2
		ORDER BY co.course_no
	`, facultyID, currentTerm)

	if err != nil {
		log.Println("current courses query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch current courses"})
	}
	defer rows.Close()

	var results []models.FacultyTeachingEntry
	for rows.Next() {
		var e models.FacultyTeachingEntry
		if err := rows.Scan(&e.Term, &e.Department, &e.Semester, &e.Batch,
			&e.CourseNo, &e.CourseName, &e.CourseType, &e.CourseCategory, &e.Credit); err != nil {
			log.Println("current courses scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read row"})
		}
		results = append(results, e)
	}

	return c.JSON(fiber.Map{
		"term":    currentTerm,
		"courses": results,
	})
}

// GetFacultyTeachingHistory returns all past-term assignments for this faculty (all-time, no limit).
func GetFacultyTeachingHistory(c *fiber.Ctx) error {
	facultyID, err := getFacultyIDFromSession(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	currentTerm := computeCurrentTerm()

	rows, err := db.Pool.Query(context.Background(), `
		SELECT a.term, a.department, a.semester, a.batch,
		       co.course_no, co.course_name, co.course_type, co.course_category, co.credit
		FROM faculty_course_assignments a
		JOIN courses co ON co.course_no = a.course_no
		WHERE a.faculty_id = $1 AND a.term != $2
		ORDER BY a.term DESC, co.course_no
	`, facultyID, currentTerm)

	if err != nil {
		log.Println("teaching history query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch teaching history"})
	}
	defer rows.Close()

	var results []models.FacultyTeachingEntry
	for rows.Next() {
		var e models.FacultyTeachingEntry
		if err := rows.Scan(&e.Term, &e.Department, &e.Semester, &e.Batch,
			&e.CourseNo, &e.CourseName, &e.CourseType, &e.CourseCategory, &e.Credit); err != nil {
			log.Println("teaching history scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read row"})
		}
		results = append(results, e)
	}

	return c.JSON(results)
}
