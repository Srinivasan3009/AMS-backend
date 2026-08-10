package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"AMS-backend/db"
	"AMS-backend/models"
)

// CreateCourse inserts a new course. TCP is always recalculated server-side as L+T+P,
// never trusted from the client, so it can't be tampered with or drift out of sync.
func CreateCourse(c *fiber.Ctx) error {
	var req models.CourseCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.CourseNo == "" || req.CourseName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_no and course_name are required"})
	}

	tcp := req.LectureHours + req.TutorialHours + req.PracticalHours

	_, err := db.Pool.Exec(context.Background(), `
		INSERT INTO courses
			(course_no, course_name, department, semester, batch, course_type, course_category,
			 lecture_hours, tutorial_hours, practical_hours, tcp, credit, active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, req.CourseNo, req.CourseName, req.Department, req.Semester, req.Batch, req.CourseType, req.CourseCategory,
		req.LectureHours, req.TutorialHours, req.PracticalHours, tcp, req.Credit, req.Active)

	if err != nil {
		log.Println("course insert error:", err)
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "course_no already exists or invalid data"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "course created"})
}

// ListCourses returns all courses, or filters by department + semester (+ optional batch, category)
// when provided (used by "View Course" and "Assign Faculty to Course").
func ListCourses(c *fiber.Ctx) error {
	department := c.Query("department")
	semester := c.Query("semester")
	batch := c.Query("batch")
	category := c.Query("category")

	baseQuery := `
		SELECT course_no, course_name, department, semester, batch, course_type, course_category,
		       lecture_hours, tutorial_hours, practical_hours, tcp, credit, active
		FROM courses
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if department != "" {
		baseQuery += fmt.Sprintf(" AND department=$%d", argIdx)
		args = append(args, department)
		argIdx++
	}
	if semester != "" {
		baseQuery += fmt.Sprintf(" AND semester=$%d", argIdx)
		args = append(args, semester)
		argIdx++
	}
	if batch != "" {
		baseQuery += fmt.Sprintf(" AND batch=$%d", argIdx)
		args = append(args, batch)
		argIdx++
	}
	if category != "" {
		baseQuery += fmt.Sprintf(" AND course_category=$%d", argIdx)
		args = append(args, category)
		argIdx++
	}
	baseQuery += " ORDER BY course_no"

	rows, err := db.Pool.Query(context.Background(), baseQuery, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not fetch courses"})
	}
	defer rows.Close()

	var courseList []models.Course
	for rows.Next() {
		var course models.Course
		if err := rows.Scan(&course.CourseNo, &course.CourseName, &course.Department, &course.Semester, &course.Batch,
			&course.CourseType, &course.CourseCategory, &course.LectureHours, &course.TutorialHours,
			&course.PracticalHours, &course.TCP, &course.Credit, &course.Active); err != nil {
			log.Println("course row scan error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not read course row"})
		}
		courseList = append(courseList, course)
	}

	return c.JSON(courseList)
}

// UpdateCourse updates a course by course_no (from URL, immutable). TCP is always recalculated server-side.
func UpdateCourse(c *fiber.Ctx) error {
	courseNo := c.Params("course_no")
	if courseNo == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "course_no is required in URL"})
	}

	var req models.CourseUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	tcp := req.LectureHours + req.TutorialHours + req.PracticalHours

	cmdTag, err := db.Pool.Exec(context.Background(), `
		UPDATE courses
		SET course_name=$1, department=$2, semester=$3, batch=$4, course_type=$5, course_category=$6,
		    lecture_hours=$7, tutorial_hours=$8, practical_hours=$9, tcp=$10, credit=$11, active=$12, updated_at=NOW()
		WHERE course_no=$13
	`, req.CourseName, req.Department, req.Semester, req.Batch, req.CourseType, req.CourseCategory,
		req.LectureHours, req.TutorialHours, req.PracticalHours, tcp, req.Credit, req.Active, courseNo)

	if err != nil {
		log.Println("course update error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not update course"})
	}
	if cmdTag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
	}

	return c.JSON(fiber.Map{"message": "course updated"})
}
