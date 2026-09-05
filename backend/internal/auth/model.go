package auth

import "time"

type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  string    `json:"avatar_url"`
	Provider   string    `json:"provider"`
	CreatedAt  time.Time `json:"created_at"`
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}
