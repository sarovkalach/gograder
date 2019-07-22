package grader

import (
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// type JwtToken struct {
// 	Secret []byte
// }

type TokenManager struct {
	Secret []byte
}

var (
	signKey               = "SecretKey"
	refreshTokenValidTime = time.Hour * 24 * 30
	acessTokenValidTime   = time.Hour * 1
)

func newTokenManager(secret string) *TokenManager {
	return &TokenManager{Secret: []byte(secret)}
}

func (j *TokenManager) refreshToken(u *User) string {
	var timeLive int64
	// if tokenType == "refreshToken" {
	// 	timeLive = time.Now().Add(refreshTokenValidTime).Unix()
	// } else {
	// 	timeLive = time.Now().Add(acessTokenValidTime).Unix()
	// }

	timestamp := strconv.FormatInt(timeLive, 10)
	// Set fields of jwt
	claims := make(jwt.MapClaims)
	claims["iss"] = "http://127.0.0.1"
	claims["type"] = "refreshToken"
	claims["type"] = "access"
	// claims["admin"] = u.Admin
	claims["sub"] = u.Email
	claims["exp"] = timestamp

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// sign token by secret key
	tokenString, _ := token.SignedString(j.Secret)
	return tokenString
}

func (j *TokenManager) accessToken(refreshToken string) {
	// user := getUserByRefreshToken(refreshToken)
}
