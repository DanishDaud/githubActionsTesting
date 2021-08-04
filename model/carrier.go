package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Carrier struct {
	ID           bson.ObjectId `json:"id" bson:"_id"`
	IP           string        `json:"ip" bson:"ip"`
	DefaultValue bool          `json:"default_value" bson:"default"`
	Name         string        `json:"name,omitempty" bson:"name"`
	CreateDate   time.Time     `json:"createDate" bson:"createDate"`
	UpdateDate   time.Time     `json:"updateDate" bson:"updateDate"`
}
