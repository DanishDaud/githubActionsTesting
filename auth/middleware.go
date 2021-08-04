package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/gomarkho/sas-rvm-provapi/model"
)

func MiddlewareRes(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		tNow := time.Now()
		tUTC := tNow
		newTUTC := tUTC.Add(time.Minute * TokenExpiryTime)

		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*JwtUserClaim)

		claims.ExpiresAt = newTUTC.Unix()

		// Generate encoded token and send it as response.
		t, err := user.SignedString([]byte(JwtSecret))
		if err != nil {
			return err
		}

		c.Response().Header().Set("x_auth_token", t)

		appContext := model.AppContext{Context: c, ID: claims.ID, Email: claims.Email}

		if err := next(appContext); err != nil {
			c.Error(err)
		}

		return nil
	}
}
