package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type CustomField struct {
	ID          bson.ObjectId `json:"_id" bson:"_id"`
	UserID      bson.ObjectId `json:"userid" bson:"userid"`
	Name        string        `json:"name" bson:"name"`
	CreatedDate time.Time     `json:"created_date" bson:"created_date"`
}

type CustomFilter struct {
	UserId bson.ObjectId
	Search string
	Page   int
	Limit  int
}
