package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	// User represents the structure of our resource
	Payment struct {
		ID         bson.ObjectId `json:"id" bson:"_id"`
		UserID     bson.ObjectId `json:"userid,omitempty" bson:"userid"`
		Amount     float32       `json:"amount" bson:"amount"`
		Comments   string        `json:"comments" bson:"comments"`
		CreateDate time.Time     `json:"createDate" bson:"createDate"`
		UpdateDate time.Time     `json:"updateDate" bson:"updateDate"`
	}
)
type Payments []Payment
