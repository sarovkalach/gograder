package jwt

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

type Claims struct {
	tokenType string
	jwt.StandardClaims
}

var (
	SignKey               = "SecretKey"
	refreshTokenValidTime = time.Hour * 24 * 30
	acessTokenValidTime   = time.Hour * 1
)

func NewJwtToken(secret string) *JwtToken {
	return &JwtToken{Secret: []byte(secret)}
}

func (j *JwtToken) parseSecretGetter(token *jwt.Token) (interface{}, error) {
	method, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok || method.Alg() != "HS256" {
		return nil, fmt.Errorf("bad sign method")
	}
	return j.Secret, nil
}

func (j *JwtToken) Check(inputToken string) (bool, error) {
	payload := &Claims{}
	_, err := jwt.ParseWithClaims(inputToken, payload, j.parseSecretGetter)
	if err != nil {
		return false, fmt.Errorf("cant parse jwt token: %v", err)
	}
	if payload.Valid() != nil {
		return false, fmt.Errorf("invalid jwt token: %v", err)
	}
	fmt.Println("Payload ID: ", payload.Id)
	return true, nil
}

func (j *JwtToken) Create(userID int, tokenType string) string {
	var timeLive int64
	if tokenType == "refreshToken" {
		timeLive = time.Now().Add(refreshTokenValidTime).Unix()
	} else {
		timeLive = time.Now().Add(acessTokenValidTime).Unix()
	}
	claims := &Claims{
		tokenType: tokenType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: timeLive,
			Id:        strconv.Itoa(userID),
		},
	}
	// claims := &jwt.StandardClaims{
	// 	ExpiresAt: timeLive,
	// 	Id:        strconv.Itoa(id),
	// }
	// timestamp := strconv.FormatInt(timeLive, 10)
	// // Set fields of jwt
	// claims := make(jwt.MapClaims)
	// claims["iss"] = "http://127.0.0.1"
	// // claims["type"] = "refreshToken"
	// claims["type"] = tokenType
	// // claims["type"] = "access"
	// // claims["admin"] = u.Admin
	// // claims["sub"] = u.Email
	// // claims["sub"] = id
	// claims["exp"] = timestamp

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
