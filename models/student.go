package models

type Student struct {
	ID           int    `json:"id"`
	RegisterNo   string `json:"register_no"`
	Name         string `json:"name"`
	DateOfBirth  string `json:"date_of_birth"`
	Gender       string `json:"gender"`
	FatherName   string `json:"father_name"`
	MotherName   string `json:"mother_name"`
	Degree       string `json:"degree"`
	Department   string `json:"department"`
	Batch        string `json:"batch"`
	JoiningYear  string `json:"joining_year"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Active       bool   `json:"active"`
}

// StudentCreateRequest is used for POST (create) - includes password since a users row is created too.
type StudentCreateRequest struct {
	RegisterNo   string `json:"register_no"`
	Name         string `json:"name"`
	DateOfBirth  string `json:"date_of_birth"`
	Gender       string `json:"gender"`
	FatherName   string `json:"father_name"`
	MotherName   string `json:"mother_name"`
	Degree       string `json:"degree"`
	Department   string `json:"department"`
	Batch        string `json:"batch"`
	JoiningYear  string `json:"joining_year"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Active       bool   `json:"active"`
}

// StudentUpdateRequest is used for PUT (modify) - password optional, register_no excluded (comes from URL, immutable).
type StudentUpdateRequest struct {
	Name         string `json:"name"`
	DateOfBirth  string `json:"date_of_birth"`
	Gender       string `json:"gender"`
	FatherName   string `json:"father_name"`
	MotherName   string `json:"mother_name"`
	Degree       string `json:"degree"`
	Department   string `json:"department"`
	Batch        string `json:"batch"`
	JoiningYear  string `json:"joining_year"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Active       bool   `json:"active"`
}

type StudentDetails struct {
	ID           int    `json:"id"`
	RegisterNo   string `json:"register_no"`
	Name         string `json:"name"`
	DateOfBirth  string `json:"date_of_birth"`
	Gender       string `json:"gender"`
	FatherName   string `json:"father_name"`
	MotherName   string `json:"mother_name"`
	Degree       string `json:"degree"` 
	Department   string `json:"department"`
	Batch        string `json:"batch"`
	JoiningYear  string `json:"joining_year"`
	MobileNumber string `json:"mobile_number"`
	Email        string `json:"email"`
}

type StudentAcademicRecord struct {
	Term           string `json:"term"`
	CourseNo       string `json:"course_no"`
	CourseName     string `json:"course_name"`
	Department     string `json:"department"`
	Semester       int    `json:"semester"`
	Batch          string `json:"batch"`
	CourseType     string `json:"course_type"`
	CourseCategory string `json:"course_category"`
	Credit         int    `json:"credit"`
	Grade *string `json:"grade"` // null if not graded yet
}

type CourseStudent struct {
	RegisterNo string `json:"register_no"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Batch      string `json:"batch"`
	Email      string `json:"email"`
}
