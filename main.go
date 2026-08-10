package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"AMS-backend/db"
	"AMS-backend/handlers"
	"AMS-backend/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	db.Connect()
	defer db.Pool.Close()

	app := fiber.New()

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     frontendURL,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept",
	}))

	// Public routes
	app.Post("/api/login", handlers.Login)
	app.Post("/api/logout", handlers.Logout)

	// Authenticated route (any logged-in role)
	app.Get("/api/me", middleware.Protected(), handlers.Me)

	// Role-scoped route groups
	admin := app.Group("/api/admin", middleware.Protected(), middleware.RequireRole("admin"))
	admin.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin dashboard data"})
	})

	admin.Post("/faculty", handlers.CreateFaculty)
	admin.Get("/faculty", handlers.ListFaculty)
	admin.Put("/faculty/:faculty_id", handlers.UpdateFaculty)

	admin.Post("/students", handlers.CreateStudent)
	admin.Get("/students", handlers.ListStudents)
	admin.Put("/students/:register_no", handlers.UpdateStudent)

	admin.Post("/courses", handlers.CreateCourse)
	admin.Get("/courses", handlers.ListCourses)
	admin.Put("/courses/:course_no", handlers.UpdateCourse)

	// Assignment routes: assign faculty to course and list assignments
	admin.Post("/assignments", handlers.UpsertAssignment)
	admin.Get("/assignments", handlers.ListAssignments)

	admin.Post("/menu-windows/course-registration", handlers.SetRegistrationWindow)
	admin.Get("/menu-windows/course-registration", handlers.GetRegistrationWindow)

	faculty := app.Group("/api/faculty", middleware.Protected(), middleware.RequireRole("faculty"))
	faculty.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "faculty dashboard data"})
	})

	faculty.Get("/details", handlers.GetFacultyDetails)
	faculty.Get("/current-courses", handlers.GetFacultyCurrentCourses)
	faculty.Get("/teaching-history", handlers.GetFacultyTeachingHistory)
	faculty.Get("/course-students", handlers.GetStudentsForCourse)
	faculty.Get("/assigned-courses", handlers.GetAssignedCourses)
	faculty.Get("/assigned-terms", handlers.GetAssignedTerms)
	faculty.Get("/grade-roster", handlers.GetGradeRoster)
	faculty.Post("/submit-grades", handlers.SubmitGrades)

	student := app.Group("/api/student", middleware.Protected(), middleware.RequireRole("student"))
	student.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "student dashboard data"})
	})
	student.Get("/details", handlers.GetStudentDetails)
	student.Get("/academic-record", handlers.GetAcademicRecord)
	student.Get("/registration-window", handlers.GetRegistrationWindowStatus)
	student.Get("/available-courses", handlers.GetAvailableCoursesForRegistration)
	student.Post("/register-course", handlers.RegisterCourse)
	student.Get("/my-registrations", handlers.GetMyRegistrations)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
