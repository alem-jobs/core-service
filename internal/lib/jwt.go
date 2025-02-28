package lib

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var Secret = "secret_key"

func NewToken(user_id int64, organization_id int64, organization_type string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user_id
	claims["organization_id"] = organization_id
	claims["organization_type"] = organization_type
	claims["exp"] = time.Now().Add(time.Hour * 24 * 365).Unix()

	tokenString, err := token.SignedString([]byte(Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
