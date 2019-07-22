package grader

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
)

type User struct {
	ID           int
	Email        string
	Password     string
	RefreshToken string
	// LastName  string
	// CreatedAt time.Time
}

type UserHandler struct {
	DBCon *sql.DB
	tm    *TokenManager
}

func (u *User) UpdateToken(DBCon *sql.DB) error {
	_, err := DBCon.Exec("UPDATE users SET refreshtoken=?  WHERE id=?", u.RefreshToken, u.ID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil

}

func getUserByID(DBCon *sql.DB, id int) *User {
	row := DBCon.QueryRow("SELECT * FROM users WHERE id = ?", id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken)
	if err != nil {
		//check User exists
		return nil
	}
	return user
}

func getUserByRefreshToken(DBCon *sql.DB, refreshtoken int) *User {
	row := DBCon.QueryRow("SELECT * FROM users WHERE refreshtoken = ?", refreshtoken)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken)
	if err != nil {
		//check User exists
		return nil
	}
	return user
}

func getUser(DBCon *sql.DB, email string, passwd string) (*User, error) {
	row := DBCon.QueryRow("SELECT * FROM users WHERE email = ?", email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken)
	if err != nil {
		//check User exists
		return nil, errBadLogin
	}
	if Encrypt(passwd) != user.Password {
		return nil, errBadPass
	}
	return user, nil
}

func Encrypt(passwd string) string {
	hashedPasswd := fmt.Sprintf("%x", sha256.Sum256([]byte(passwd)))
	return hashedPasswd
}
