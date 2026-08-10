package models

type User struct {
	ID           int    `json:"id"`
	Role         string `json:"role"`
	RegisterNo   string `json:"register_no,omitempty"`
	FacultyID    string `json:"faculty_id,omitempty"`
	Username     string `json:"username,omitempty"`
	Name         string `json:"name"`
	PasswordHash string `json:"-"`
}

// LoginRequest: no role field — the identifier alone determines the role.
type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Role    string `json:"role"`
	Name    string `json:"name"`
}
