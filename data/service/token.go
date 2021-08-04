package service

import (
	"time"

	"github.com/dchest/passwordreset"
)

const secret = "c2d1c33a-1f42-42d9-b126-a669da826202"

type Token struct {
}

func NewTokenService() *Token {
	return &Token{}
}

func (t *Token) GenerateToken(email string) string {
	if email == "" {
		return email
	}

	pwdVal, _ := getPasswordHash(email)
	return passwordreset.NewToken(email, time.Minute*60, []byte(pwdVal), []byte(secret))
}

func (t *Token) Verify(token string) (string, error) {
	return passwordreset.VerifyToken(token, getPasswordHash, []byte(secret))
}

func getPasswordHash(login string) ([]byte, error) {
	return []byte(login), nil
}
