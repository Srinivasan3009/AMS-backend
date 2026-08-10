package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"AMS-backend/db"
	"AMS-backend/models"
)

// CreateStudent inserts into users + student_details in a single transaction.
func CreateStudent(c *fiber.Ctx) error {
	var req models.StudentCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.RegisterNo == "" || req.Name == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "register_no, name and password are required"})
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
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO users (role, register_no, name, password_hash) VALUES ('student', $1, $2, $3)`,
		req.RegisterNo, req.Name, string(hash),
	)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "register_no already exists or invalid data"})
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO student_details
			(register_no, name, date_of_birth, gender, father_name, mother_name, degree, department, batch, joining_year, mobile_number, email, active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		req.RegisterNo, req.Name, req.DateOfBirth, req.Gender, req.FatherName, req.MotherName,
		req.Degree, req.Department, req.Batch, req.JoiningYear, req.MobileNumber, req.Email, req.Active,
	)
	if err != nil {
		log.Println("student_details insert error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not create student details"})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not commit transaction"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "student created"})
}

// ListStudents returns all student records for the admin table.
func ListStudents(c *fiber.Ctx) error {
	rows, err := db.Pool.Query(context.Background(), `
		SELECT register_no, name, date_of_birth::text, gender, father_name, mother_name,
		       degree, department, batch, joining_year, mobile_number, email, active
		FROM student_details
		ORDER BY name
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch students"})
	}
	defer rows.Close()

	var studentList []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.RegisterNo, &s.Name, &s.DateOfBirth, &s.Gender, &s.FatherName, &s.MotherName,
			&s.Degree, &s.Department, &s.Batch, &s.JoiningYear, &s.MobileNumber, &s.Email, &s.Active); err != nil {
			log.Println("student row scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read student row"})
		}
		studentList = append(studentList, s)
	}

	return c.JSON(studentList)
}

// UpdateStudent updates student_details always, and users.password_hash only if a new password was provided.
// register_no comes from the URL and is never modified.
func UpdateStudent(c *fiber.Ctx) error {
	registerNo := c.Params("register_no")
	if registerNo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "register_no is required in URL"})
	}

	var req models.StudentUpdateRequest
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
		UPDATE student_details
		SET name=$1, date_of_birth=$2, gender=$3, father_name=$4, mother_name=$5,
		    degree=$6, department=$7, batch=$8, joining_year=$9, mobile_number=$10, email=$11, active=$12, updated_at=NOW()
		WHERE register_no=$13
	`, req.Name, req.DateOfBirth, req.Gender, req.FatherName, req.MotherName,
		req.Degree, req.Department, req.Batch, req.JoiningYear, req.MobileNumber, req.Email, req.Active, registerNo)

	if err != nil {
		log.Println("student_details update error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update student details"})
	}
	if cmdTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student not found"})
	}

	_, err = tx.Exec(ctx, `UPDATE users SET name=$1 WHERE register_no=$2`, req.Name, registerNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update user record"})
	}

	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
		}
		_, err = tx.Exec(ctx, `UPDATE users SET password_hash=$1 WHERE register_no=$2`, string(hash), registerNo)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update password"})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "student updated"})
}

// GetStudentDetails returns the logged-in student's own profile, including degree.
func GetStudentDetails(c *fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	userID, ok := userIDRaw.(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user id"})
	}

	var details models.StudentDetails
	row := db.Pool.QueryRow(context.Background(), `
		SELECT sd.id, sd.register_no, sd.name, sd.date_of_birth::text, sd.gender,
		       sd.father_name, sd.mother_name, sd.degree, sd.department, sd.batch,
		       sd.joining_year, sd.mobile_number, sd.email
		FROM student_details sd
		JOIN users u ON u.register_no = sd.register_no
		WHERE u.id = $1 AND u.role = 'student'
	`, int(userID))

	if err := row.Scan(&details.ID, &details.RegisterNo, &details.Name, &details.DateOfBirth,
		&details.Gender, &details.FatherName, &details.MotherName, &details.Degree,
		&details.Department, &details.Batch, &details.JoiningYear,
		&details.MobileNumber, &details.Email); err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student details not found"})
		}
		log.Println("student details query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch student details"})
	}

	return c.JSON(details)
}

// GetAcademicRecord returns the student's profile plus ONLY the courses they have
// actually registered for (via course_registrations) - not every course offered
// to their department+batch. This replaces the earlier department+batch approximation.
func GetAcademicRecord(c *fiber.Ctx) error {
	userIDRaw := c.Locals("user_id")
	userID, ok := userIDRaw.(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid user id"})
	}

	var details models.StudentDetails
	row := db.Pool.QueryRow(context.Background(), `
		SELECT sd.id, sd.register_no, sd.name, sd.date_of_birth::text, sd.gender,
		       sd.father_name, sd.mother_name, sd.degree, sd.department, sd.batch,
		       sd.joining_year, sd.mobile_number, sd.email
		FROM student_details sd
		JOIN users u ON u.register_no = sd.register_no
		WHERE u.id = $1 AND u.role = 'student'
	`, int(userID))

	if err := row.Scan(&details.ID, &details.RegisterNo, &details.Name, &details.DateOfBirth,
		&details.Gender, &details.FatherName, &details.MotherName, &details.Degree,
		&details.Department, &details.Batch, &details.JoiningYear,
		&details.MobileNumber, &details.Email); err != nil {
		if err == pgx.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student details not found"})
		}
		log.Println("student details query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch student details"})
	}

	rows, err := db.Pool.Query(context.Background(), `
		SELECT cr.term, co.course_no, co.course_name, co.department, co.semester, co.batch,
		       co.course_type, co.course_category, co.credit, g.grade
		FROM course_registrations cr
		JOIN courses co ON co.course_no = cr.course_no
		LEFT JOIN grades g ON g.register_no = cr.register_no AND g.course_no = cr.course_no AND g.term = cr.term
		WHERE cr.register_no = $1
		ORDER BY cr.term DESC, co.semester, co.course_no
	`, details.RegisterNo)
	if err != nil {
		log.Println("academic record query error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch academic record"})
	}
	defer rows.Close()

	var records []models.StudentAcademicRecord
	for rows.Next() {
		var rec models.StudentAcademicRecord
		if err := rows.Scan(&rec.Term, &rec.CourseNo, &rec.CourseName, &rec.Department, &rec.Semester,
			&rec.Batch, &rec.CourseType, &rec.CourseCategory, &rec.Credit, &rec.Grade); err != nil {
			log.Println("academic record scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read academic record"})
		}
		records = append(records, rec)
	}

	return c.JSON(fiber.Map{"student": details, "courses": records})
}
