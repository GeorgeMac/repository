package models

import "time"

type Repository struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	FetchedAt time.Time `json:"fetchedAt"`
}
