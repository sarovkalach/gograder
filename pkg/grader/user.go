package grader

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/sarovkalach/gograder/pkg/jwt"
)

type User struct {
	ID           int
	Email        string
	Password     string
	RefreshToken string
	// LastName  string
	// CreatedAt time.Time
}

var (
	errTokenNotValid = errors.New("Token not valid")
)

type UserHandler struct {
	DBCon *sql.DB
	token *jwt.JwtToken
}

func (u *User) UpdateToken(DBCon *sql.DB) error {
	_, err := DBCon.Exec("UPDATE users SET refreshtoken=?  WHERE id=?", u.RefreshToken, u.ID)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil

}
func (u *User) CheckToken(DBCon *sql.DB, token string) error {
	row := DBCon.QueryRow("SELECT refreshtoken FROM users WHERE refreshtoken = ?", token)
	var refreshToken string
	err := row.Scan(&refreshToken)
	if err != nil {
		//check User exists
		return errTokenNotValid
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

func getUserByEmail(DBCon *sql.DB, email string) (*User, error) {
	row := DBCon.QueryRow("SELECT * FROM users WHERE refreshtoken = ?", email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken)
	if err != nil {
		//check User exists
		return nil, err
	}
	return user, nil
}

func getUserByRefreshToken(DBCon *sql.DB, refreshtoken string) (*User, error) {
	row := DBCon.QueryRow("SELECT * FROM users WHERE refreshtoken = ?", refreshtoken)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.RefreshToken)
	if err != nil {
		//check User exists
		return nil, err
	}
	return user, nil
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
