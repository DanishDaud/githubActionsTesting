package model

import (
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

type AppContext struct {
	echo.Context
	ID    bson.ObjectId `json:"id" bson:"_id"`
	Role  int8          `json:"role" bson:"email"`
	Email string        `json:"email" bson:"email"`
}
