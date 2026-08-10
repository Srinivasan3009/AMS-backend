package models

type FacultyTeachingEntry struct {
	Term           string `json:"term"`
	Department     string `json:"department"`
	Semester       int    `json:"semester"`
	Batch          string `json:"batch"`
	CourseNo       string `json:"course_no"`
	CourseName     string `json:"course_name"`
	CourseType     string `json:"course_type"`
	CourseCategory string `json:"course_category"`
	Credit         int    `json:"credit"`
}