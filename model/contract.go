package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	// User represents the structure of our resource
	Contract struct {
		ID           bson.ObjectId `json:"id" bson:"_id"`
		UserID       bson.ObjectId `json:"userid,omitempty" bson:"userid"`
		ContractType int8          `json:"contractType,omitempty" bson:"contractType"`
		StartDate    time.Time     `json:"startDate,omitempty" bson:"startDate"`
		EndDate      time.Time     `json:"endDate,omitempty" bson:"endDate"`
		CreateDate   time.Time     `json:"createDate" bson:"createDate"`
		UpdateDate   time.Time     `json:"updateDate" bson:"updateDate"`
	}
)

type Contracts []Contract
