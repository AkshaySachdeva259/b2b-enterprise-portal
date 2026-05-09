package models

import "time"

type Page struct {
	ID        int       `json:"id" db:"id"`
	PageName  string    `json:"page_name" db:"page_name"`
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
