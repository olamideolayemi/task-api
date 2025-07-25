package models

type Task struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Details string `json:"details"`
	Done    bool   `json:"done"`
}
