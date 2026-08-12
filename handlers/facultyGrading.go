package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"

	"AMS-backend/db"
	"AMS-backend/models"
	"AMS-backend/utils"
)

var validGrades = map[string]bool{
	"O": true, "A+": true, "A": true, "B+": true, "B": true,
	"C": true, "U": true, "RA": true, "SA": true, "W": true,
}

// GetAssignedCourses (faculty) - distinct courses this faculty is assigned to teach,
// optionally filtered by department, for the "Course No" dropdown.
func GetAssignedCourses(c *fiber.Ctx) error {
	facultyID, err := getFacultyIDFromSession(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	department := c.Query("department")

	query := `
		SELECT DISTINCT co.course_no, co.course_name
		FROM faculty_course_assignments a
		JOIN courses co ON co.course_no = a.course_no
		WHERE a.faculty_id = $1
	`
	args := []interface{}{facultyID}
	if department != "" {
		query += " AND a.department = $2"
		args = append(args, department)
	}
	query += " ORDER BY co.course_no"

	rows, err := db.Pool.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch assigned courses"})
	}
	defer rows.Close()

	var options []models.AssignedCourseOption
	for rows.Next() {
		var o models.AssignedCourseOption
		if err := rows.Scan(&o.CourseNo, &o.CourseName); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read course row"})
		}
		options = append(options, o)
	}

	return c.JSON(options)
}

// GetAssignedTerms (faculty) - distinct terms this faculty taught a specific course.
func GetAssignedTerms(c *fiber.Ctx) error {
	facultyID, err := getFacultyIDFromSession(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	courseNo := c.Query("course_no")
	if courseNo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_no is required"})
	}

	rows, err := db.Pool.Query(context.Background(), `
		SELECT DISTINCT term FROM faculty_course_assignments
		WHERE faculty_id = $1 AND course_no = $2
		ORDER BY term DESC
	`, facultyID, courseNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch terms"})
	}
	defer rows.Close()

	var terms []string
	for rows.Next() {
		var term string
		if err := rows.Scan(&term); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read term row"})
		}
		terms = append(terms, term)
	}

	return c.JSON(terms)
}

func isAssignedToCourse(facultyID, courseNo, term string) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1 FROM faculty_course_assignments
			WHERE faculty_id = $1 AND course_no = $2 AND term = $3
		)
	`, facultyID, courseNo, term).Scan(&exists)
	return exists, err
}

// GetGradeRoster (faculty) - registered students for course_no+term, with their existing
// grade if already submitted (null otherwise). Ownership-checked.
func GetGradeRoster(c *fiber.Ctx) error {
	facultyID, err := getFacultyIDFromSession(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	courseNo := c.Query("course_no")
	term := c.Query("term")
	if courseNo == "" || term == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_no and term are required"})
	}

	assigned, err := isAssignedToCourse(facultyID, courseNo, term)
	if err != nil || !assigned {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you are not assigned to this course for this term"})
	}

	rows, err := db.Pool.Query(context.Background(), `
		SELECT sd.register_no, sd.name, g.grade
		FROM course_registrations cr
		JOIN student_details sd ON sd.register_no = cr.register_no
		LEFT JOIN grades g ON g.register_no = cr.register_no AND g.course_no = cr.course_no AND g.term = cr.term
		WHERE cr.course_no = $1 AND cr.term = $2
		ORDER BY sd.name
	`, courseNo, term)
	if err != nil {
		log.Println("grade roster query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch grade roster"})
	}
	defer rows.Close()

	var roster []models.GradeRosterEntry
	for rows.Next() {
		var r models.GradeRosterEntry
		if err := rows.Scan(&r.RegisterNo, &r.Name, &r.Grade); err != nil {
			log.Println("grade roster scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read roster row"})
		}
		roster = append(roster, r)
	}

	return c.JSON(roster)
}

// gradedStudent holds what we need post-commit to send a notification email.
type gradedStudent struct {
	RegisterNo string
	Name       string
	Email      string
	Grade      string
}

// SubmitGrades (faculty) - bulk upserts grades for course_no+term.
// Every row is validated individually: invalid grade values or unregistered
// students are reported back in "failed" with a reason, WITHOUT failing the
// rest of the batch. All other valid rows are still saved.
// After a successful commit, one notification email is sent per successfully
// graded student; any email send failures are reported separately and do NOT
// roll back the saved grades (the grade is the source of truth - email is best-effort).
func SubmitGrades(c *fiber.Ctx) error {
	facultyID, err := getFacultyIDFromSession(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	var req models.SubmitGradesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.CourseNo == "" || req.Term == "" || len(req.Grades) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_no, term and grades are required"})
	}

	assigned, err := isAssignedToCourse(facultyID, req.CourseNo, req.Term)
	if err != nil || !assigned {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you are not assigned to this course for this term"})
	}

	var courseName string
	err = db.Pool.QueryRow(context.Background(),
		`SELECT course_name FROM courses WHERE course_no = $1`, req.CourseNo,
	).Scan(&courseName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not resolve course details"})
	}

	ctx := context.Background()
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not start transaction"})
	}
	defer tx.Rollback(ctx)

	var failed []models.FailedGradeRow
	var toNotify []gradedStudent

	for _, entry := range req.Grades {
		if entry.RegisterNo == "" || entry.Grade == "" {
			failed = append(failed, models.FailedGradeRow{RegisterNo: entry.RegisterNo, Reason: "missing register number or grade"})
			continue
		}
		if !validGrades[entry.Grade] {
			failed = append(failed, models.FailedGradeRow{RegisterNo: entry.RegisterNo, Reason: "invalid grade value: " + entry.Grade})
			continue
		}

		// Confirm the student is registered for this course+term, and fetch their
		// name+email in the same query (covers "unknown roll number" too, since
		// a nonexistent register_no simply won't match any row here).
		var name, email string
		err = tx.QueryRow(ctx, `
			SELECT sd.name, sd.email
			FROM course_registrations cr
			JOIN student_details sd ON sd.register_no = cr.register_no
			WHERE cr.register_no = $1 AND cr.course_no = $2 AND cr.term = $3
		`, entry.RegisterNo, req.CourseNo, req.Term).Scan(&name, &email)

		if err != nil {
			failed = append(failed, models.FailedGradeRow{RegisterNo: entry.RegisterNo, Reason: "student not registered for this course/term, or unknown register number"})
			continue
		}

		// Check the existing grade (if any) BEFORE writing. If it's identical to what's
		// being submitted, skip entirely - no DB write, no email. This is what actually
		// prevents "editing one student re-notifies everyone" - it applies regardless of
		// whether the unchanged rows came from an individual edit or a re-uploaded CSV
		// that still contains everyone's previously-existing grades.
		var existingGrade *string
		err = tx.QueryRow(ctx, `
			SELECT grade FROM grades WHERE register_no = $1 AND course_no = $2 AND term = $3
		`, entry.RegisterNo, req.CourseNo, req.Term).Scan(&existingGrade)
		// err == pgx.ErrNoRows just means no prior grade exists yet - that's fine, proceed.

		if existingGrade != nil && *existingGrade == entry.Grade {
			continue // no change - skip silently, not an error
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO grades (register_no, course_no, term, grade, graded_by)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (register_no, course_no, term)
			DO UPDATE SET grade = EXCLUDED.grade, graded_by = EXCLUDED.graded_by, updated_at = NOW()
		`, entry.RegisterNo, req.CourseNo, req.Term, entry.Grade, facultyID)

		if err != nil {
			log.Println("grade upsert error:", err)
			failed = append(failed, models.FailedGradeRow{RegisterNo: entry.RegisterNo, Reason: "database error saving this grade"})
			continue
		}

		toNotify = append(toNotify, gradedStudent{
			RegisterNo: entry.RegisterNo,
			Name:       name,
			Email:      email,
			Grade:      entry.Grade,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not commit transaction"})
	}

	// Grades are now durably saved. Email is sent AFTER commit and is best-effort:
	// a failed email never un-saves a grade, since the portal is always the
	// source of truth and stays in sync regardless of email delivery.
	var emailFailures []models.FailedGradeRow

	for _, s := range toNotify {
		err := utils.SendGradeNotification(
			s.Email,
			s.Name,
			req.CourseNo,
			courseName,
			req.Term,
			s.Grade,
		)

		// Default: email was successfully accepted by Resend.
		status := "success"
		var failureReason *string

		if err != nil {
			status = "failure"

			reason := err.Error()
			failureReason = &reason

			log.Println(
				"grade email send error for",
				s.RegisterNo,
				":",
				err,
			)

			emailFailures = append(emailFailures, models.FailedGradeRow{
				RegisterNo: s.RegisterNo,
				Reason:     "email could not be sent",
			})
		}

		// Save the email notification attempt to email_logs.
		_, logErr := db.Pool.Exec(ctx, `
		INSERT INTO email_logs
			(register_no, course_no, term, grade, status, failure_reason, sent_by)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
	`,
			s.RegisterNo,
			req.CourseNo,
			req.Term,
			s.Grade,
			status,
			failureReason,
			facultyID,
		)

		if logErr != nil {
			log.Println(
				"failed to save email log for",
				s.RegisterNo,
				":",
				logErr,
			)
		}
	}

	return c.JSON(models.SubmitGradesResponse{
		Message:        "grade submission processed",
		SucceededCount: len(toNotify),
		Failed:         failed,
		EmailFailures:  emailFailures,
	})
}

// GetEmailLog (faculty) - returns this faculty member's own email notification
// history (success and failure), most recent first, joined with student name
// and course name for readability.
func GetEmailLog(c *fiber.Ctx) error {
	facultyID, err := getFacultyIDFromSession(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "faculty not found"})
	}

	rows, err := db.Pool.Query(context.Background(), `
		SELECT e.id, e.register_no, sd.name, e.course_no, co.course_name, e.term, e.grade,
		       e.status, e.failure_reason, e.sent_at::text
		FROM email_logs e
		JOIN student_details sd ON sd.register_no = e.register_no
		JOIN courses co ON co.course_no = e.course_no
		WHERE e.sent_by = $1
		ORDER BY e.sent_at DESC
	`, facultyID)
	if err != nil {
		log.Println("email log query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch email log"})
	}
	defer rows.Close()

	var logs []models.EmailLogEntry
	for rows.Next() {
		var e models.EmailLogEntry
		if err := rows.Scan(&e.ID, &e.RegisterNo, &e.StudentName, &e.CourseNo, &e.CourseName,
			&e.Term, &e.Grade, &e.Status, &e.FailureReason, &e.SentAt); err != nil {
			log.Println("email log scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read email log row"})
		}
		logs = append(logs, e)
	}

	return c.JSON(logs)
}
