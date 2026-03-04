package domain

import "time"

type Post struct {
	ID        string    `json:"id" redis:"id"`
	Title     string    `json:"title" redis:"title"`
	Author    string    `json:"author" redis:"author"`
	Content   string    `json:"content" redis:"content"`
	Tags      []string  `json:"tags" redis:"tags"`
	CreatedAt time.Time `json:"created_at" redis:"created_at"`
}
