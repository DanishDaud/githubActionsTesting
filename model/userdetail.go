package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	UserDetail struct {
		ID                   bson.ObjectId `json:"id" bson:"_id"`
		UserID               bson.ObjectId `json:"userid" bson:"userid"`
		TotalNumbersUploaded int64         `json:"totalNumbersUploaded" bson:"totalNumbersUploaded"`
		CreateDate           time.Time     `json:"createDate" bson:"createDate"`
		UpdateDate           time.Time     `json:"updateDate" bson:"updateDate"`
	}
)
