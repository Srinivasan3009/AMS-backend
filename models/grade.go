package models

type GradeRosterEntry struct {
	RegisterNo string  `json:"register_no"`
	Name       string  `json:"name"`
	Grade      *string `json:"grade"` // null if not graded yet
}

type GradeSubmitEntry struct {
	RegisterNo string `json:"register_no"`
	Grade      string `json:"grade"`
}

type SubmitGradesRequest struct {
	CourseNo string             `json:"course_no"`
	Term     string             `json:"term"`
	Grades   []GradeSubmitEntry `json:"grades"`
}

type FailedGradeRow struct {
	RegisterNo string `json:"register_no"`
	Reason     string `json:"reason"`
}

type SubmitGradesResponse struct {
	Message        string           `json:"message"`
	SucceededCount int              `json:"succeeded_count"`
	Failed         []FailedGradeRow `json:"failed"`
	EmailFailures  []FailedGradeRow `json:"email_failures"`
}

type AssignedCourseOption struct {
	CourseNo   string `json:"course_no"`
	CourseName string `json:"course_name"`
}
type EmailLogEntry struct {
	ID            int     `json:"id"`
	RegisterNo    string  `json:"register_no"`
	StudentName   string  `json:"student_name"`
	CourseNo      string  `json:"course_no"`
	CourseName    string  `json:"course_name"`
	Term          string  `json:"term"`
	Grade         string  `json:"grade"`
	Status        string  `json:"status"`
	FailureReason *string `json:"failure_reason"`
	SentAt        string  `json:"sent_at"`
}
