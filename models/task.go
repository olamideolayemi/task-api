package models

import "github.com/google/uuid"

type Task struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Details  string    `json:"details"`
	Done     bool      `json:"done"`
	ImageURL string    `json:"image_url"`
	UserID   uuid.UUID `json:"user_id"`
}
