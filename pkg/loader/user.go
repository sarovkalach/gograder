package loader

import "time"

type User struct {
	ID       int
	Email    string
	Password string
	// FirstName string
	// LastName  string
	// CreatedAt time.Time
}

type Session struct {
	Id        int
	Uuid      string
	Email     string
	UserId    int
	CreatedAt time.Time
}
