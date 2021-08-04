package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type (
	// Balance represents the structure of our resource
	Balance struct {
		ID              bson.ObjectId `json:"id" bson:"_id"`
		UserID          bson.ObjectId `json:"userid" bson:"userid"`
		IsConsumedFully bool          `json:"isConsumedFully" bson:"isConsumedFully"`
		IsGhost         bool          `json:"isGhost" bson:"isGhost"`
		Amount          float32       `json:"amount" bson:"amount"`
		Consumed        float32       `json:"consumed" bson:"consumed"`
		Expired         bool          `json:"expired" bson:"expired"`
		ExpiryDate      *time.Time    `json:"expiry,omitempty" bson:"expiry,omitempty"`
		CreateDate      time.Time     `json:"createDate" bson:"createDate"`
		UpdateDate      time.Time     `json:"updateDate" bson:"updateDate"`
	}

	Balances []Balance
)
