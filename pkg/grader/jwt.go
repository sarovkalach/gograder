package grader

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// type JwtToken struct {
// 	Secret []byte
// }
var (
	errBadToken = errors.New("Bad Token")
)

type JwtToken struct {
	Secret []byte
}

var (
	signKey               = "SecretKey"
	refreshTokenValidTime = time.Hour * 24 * 30
	acessTokenValidTime   = time.Hour * 1
)

func newJwtToken(secret string) *JwtToken {
	return &JwtToken{Secret: []byte(secret)}
}

func (j *JwtToken) parseSecretGetter(token *jwt.Token) (interface{}, error) {
	method, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok || method.Alg() != "HS256" {
		return nil, fmt.Errorf("bad sign method")
	}
	return j.Secret, nil
}

func (j *JwtToken) Create(id int, tokenType string) string {
	var timeLive int64
	if tokenType == "refreshToken" {
		timeLive = time.Now().Add(refreshTokenValidTime).Unix()
	} else {
		timeLive = time.Now().Add(acessTokenValidTime).Unix()
	}
	timestamp := strconv.FormatInt(timeLive, 10)
	// Set fields of jwt
	claims := make(jwt.MapClaims)
	claims["iss"] = "http://127.0.0.1"
	// claims["type"] = "refreshToken"
	claims["type"] = tokenType
	// claims["type"] = "access"
	// claims["admin"] = u.Admin
	// claims["sub"] = u.Email
	claims["sub"] = id
	claims["exp"] = timestamp

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// sign token by secret key
	tokenString, _ := token.SignedString(j.Secret)
	return tokenString
}

// func (j *TokenManager) accessToken(DBCon *sql.DB, refreshToken string) (string, error) {
// 	user, err := getUserByRefreshToken(DBCon, refreshToken)
// 	if err != nil {
// 		return "", errBadToken
// 	}
// 	fmt.Println(user)
// 	return "", nil
// }
