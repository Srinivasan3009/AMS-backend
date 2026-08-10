package models

type Assignment struct {
	ID          int    `json:"id"`
	CourseNo    string `json:"course_no"`
	FacultyID   string `json:"faculty_id"`
	FacultyName string `json:"faculty_name"`
	Term        string `json:"term"`
	Department  string `json:"department"`
	Semester    int    `json:"semester"`
	Batch       string `json:"batch"`
}

type AssignmentRequest struct {
	CourseNo   string `json:"course_no"`
	FacultyID  string `json:"faculty_id"`
	Term       string `json:"term"`
	Department string `json:"department"`
	Semester   int    `json:"semester"`
	Batch      string `json:"batch"`
}
