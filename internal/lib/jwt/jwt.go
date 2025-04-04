package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/rinefica/voice_null_sso/internal/domain/model"
	"time"
)

func CreateToken(user *model.User, app *model.App, tokenTTL time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  user.Email,
		"exp":    time.Now().Add(tokenTTL).Unix(),
		"app_id": app.ID,
	})

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
