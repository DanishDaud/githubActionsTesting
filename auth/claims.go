package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gomarkho/sas-rvm-provapi/model"
	"gopkg.in/mgo.v2/bson"
)

// jwt token claims which contains info regarding user
type JwtUserClaim struct {
	ID        bson.ObjectId `json:"id" bson:"_id"`
	FirstName string        `json:"firstName" bson:"firstName"`
	LastName  string        `json:"lastName" bson:"lastName"`
	Email     string        `json:"email" bson:"email"`
	Role      int8          `json:"role" bson:"role"`
	jwt.StandardClaims
}

func Token(user model.User) (string, error) {

	tNow := time.Now()
	tUTC := tNow
	newTUTC := tUTC.Add(time.Minute * TokenExpiryTime)

	// Set custom claims
	claims := &JwtUserClaim{
		user.ID,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Role,
		jwt.StandardClaims{
			ExpiresAt: newTUTC.Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(JwtSecret))
	if err != nil {
		return "", err
	}

	return t, nil
}
