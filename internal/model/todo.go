package model

import "time"

// Todo is a model for a todo item.
type Todo struct {
	ID          int64
	Title       string
	Description *string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
