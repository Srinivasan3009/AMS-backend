package models

type RegistrationWindow struct {
	FeatureKey    string `json:"feature_key"`
	StartDatetime string `json:"start_datetime"`
	EndDatetime   string `json:"end_datetime"`
}

// SetWindowRequest - admin sends separate date + time fields from the form,
// combined server-side into a single timestamp.
type SetWindowRequest struct {
	StartDate string `json:"start_date"` // e.g. "2026-07-15"
	StartTime string `json:"start_time"` // e.g. "09:00"
	EndDate   string `json:"end_date"`
	EndTime   string `json:"end_time"`
}

type WindowStatusResponse struct {
	IsOpen        bool   `json:"is_open"`
	StartDatetime string `json:"start_datetime"`
	EndDatetime   string `json:"end_datetime"`
}
