package models

type CourseRegistration struct {
	ID           int    `json:"id"`
	RegisterNo   string `json:"register_no"`
	CourseNo     string `json:"course_no"`
	Term         string `json:"term"`
	RegisteredAt string `json:"registered_at"`
}

type RegisterCourseRequest struct {
	CourseNo string `json:"course_no"`
}
