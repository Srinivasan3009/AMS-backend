package models

type Faculty struct {
	ID               int     `json:"id"`
	FacultyID        string  `json:"faculty_id"`
	Name             string  `json:"name"`
	DateOfBirth      string  `json:"date_of_birth"`
	Gender           string  `json:"gender"`
	Designation      string  `json:"designation"`
	Department       string  `json:"department"`
	MobileNumber     string  `json:"mobile_number"`
	Email            string  `json:"email"`
	Address1         string  `json:"address_1"`
	Address2         string  `json:"address_2"`
	Active           bool    `json:"active"`
	DateOfRetirement *string `json:"date_of_retirement"`
}

// FacultyCreateRequest is used for POST (create) - includes password since a users row is created too.
type FacultyCreateRequest struct {
	FacultyID        string  `json:"faculty_id"`
	Name             string  `json:"name"`
	DateOfBirth      string  `json:"date_of_birth"`
	Gender           string  `json:"gender"`
	Designation      string  `json:"designation"`
	Department       string  `json:"department"`
	MobileNumber     string  `json:"mobile_number"`
	Email            string  `json:"email"`
	Address1         string  `json:"address_1"`
	Address2         string  `json:"address_2"`
	Password         string  `json:"password"`
	Active           bool    `json:"active"`
	DateOfRetirement *string `json:"date_of_retirement"`
}

// FacultyUpdateRequest is used for PUT (modify) - password optional, faculty_id excluded (comes from URL, immutable).
type FacultyUpdateRequest struct {
	Name             string  `json:"name"`
	DateOfBirth      string  `json:"date_of_birth"`
	Gender           string  `json:"gender"`
	Designation      string  `json:"designation"`
	Department       string  `json:"department"`
	MobileNumber     string  `json:"mobile_number"`
	Email            string  `json:"email"`
	Address1         string  `json:"address_1"`
	Address2         string  `json:"address_2"`
	Password         string  `json:"password"` // optional - only update if non-empty
	Active           bool    `json:"active"`
	DateOfRetirement *string `json:"date_of_retirement"`
}
