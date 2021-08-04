package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type HostedCheckout struct {
	RedirectURI string `json:"redirect_uri" bson:"redirect_uri"`
	CheckoutURI string `json:"checkout_uri" bson:"checkout_uri"`
}

type WepayCheckout struct {
	ID               bson.ObjectId  `json:"id" bson:"_id"`
	UserID           bson.ObjectId  `json:"userid" bson:"userid"`
	CheckoutId       int64          `json:"checkout_id" bson:"checkout_id"`
	AccountId        int64          `json:"account_id" bson:"account_id"`
	Amount           float32        `json:"amount" bson:"amount"`
	Type             string         `json:"type" bson:"type"`
	Currency         string         `json:"currency" bson:"currency"`
	ShortDescription string         `json:"short_description" bson:"short_description"`
	HostedCheckout   HostedCheckout `json:"hosted_checkout" bson:"hosted_checkout"`
	CreateDate       time.Time      `json:"createDate" bson:"createDate"`
	UpdateDate       time.Time      `json:"updateDate" bson:"updateDate"`
	CallbackUri      string         `json:"callback_uri" bson:"callback_uri"`
	Consumed         bool           `json:"consumed" bson:"consumed"`
}
