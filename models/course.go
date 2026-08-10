package models

type Course struct {
	ID             int    `json:"id"`
	CourseNo       string `json:"course_no"`
	CourseName     string `json:"course_name"`
	Department     string `json:"department"`
	Semester       int    `json:"semester"`
	Batch          string `json:"batch"`
	CourseType     string `json:"course_type"`
	CourseCategory string `json:"course_category"`
	LectureHours   int    `json:"lecture_hours"`
	TutorialHours  int    `json:"tutorial_hours"`
	PracticalHours int    `json:"practical_hours"`
	TCP            int    `json:"tcp"`
	Credit         int    `json:"credit"`
	Active         bool   `json:"active"`
}

// CourseCreateRequest - TCP is NOT accepted from client, always recalculated server-side as L+T+P.
type CourseCreateRequest struct {
	CourseNo       string `json:"course_no"`
	CourseName     string `json:"course_name"`
	Department     string `json:"department"`
	Semester       int    `json:"semester"`
	Batch          string `json:"batch"`
	CourseType     string `json:"course_type"`
	CourseCategory string `json:"course_category"`
	LectureHours   int    `json:"lecture_hours"`
	TutorialHours  int    `json:"tutorial_hours"`
	PracticalHours int    `json:"practical_hours"`
	Credit         int    `json:"credit"`
	Active         bool   `json:"active"`
}

// CourseUpdateRequest - course_no excluded (comes from URL, immutable). TCP recalculated server-side too.
type CourseUpdateRequest struct {
	CourseName     string `json:"course_name"`
	Department     string `json:"department"`
	Semester       int    `json:"semester"`
	Batch          string `json:"batch"`
	CourseType     string `json:"course_type"`
	CourseCategory string `json:"course_category"`
	LectureHours   int    `json:"lecture_hours"`
	TutorialHours  int    `json:"tutorial_hours"`
	PracticalHours int    `json:"practical_hours"`
	Credit         int    `json:"credit"`
	Active         bool   `json:"active"`
}
