package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	BillingHistoryPaid    = 1
	BillingHistoryPending = 2
	BillingHistoryError   = 3
)

type BillingHistory struct {
	ID         bson.ObjectId `json:"id" bson:"_id"`
	Status     int8          `json:"status" bson:"status"`
	UserID     bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	Amount     float32       `json:"amount" bson:"amount"`
	CreateDate time.Time     `json:"createDate" bson:"createDate"`
	UpdateDate time.Time     `json:"updateDate" bson:"updateDate"`
}

type BillingHistories []BillingHistory
