package loader

import (
	"net/http"
	"strconv"
	"time"
)

type SessionUser struct {
	ID            int
	Authenticated bool
}

func newCookie(id int) http.Cookie {
	expiration := time.Now().Add(3 * time.Hour)
	cookie := http.Cookie{
		Name:     "_cookie",
		Value:    strconv.Itoa(id), // hardcoderd
		HttpOnly: true,
		Expires:  expiration,
	}
	return cookie
}
