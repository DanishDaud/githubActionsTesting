package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type BulkVSNumber struct {
	ID               bson.ObjectId `json:"id" bson:"_id"`
	UserID           bson.ObjectId ` json:"userID" bson:"userID"`
	PhoneID          string        `json:"phoneID" bson:"phoneID"`
	ConnectionID     string        `json:"connectionID" bson:"connectionID"`
	PhoneStatus      string        `json:"phoneStatus" bson:"phoneStatus"`
	PhoneNumber      string        `json:"phoneNumber" bson:"phoneNumber"`
	ForwardingNumber string        `json:"forwardingNumber" bson:"forwardingNumber"`
	TotalInbound     int           `json:"totalInbound" bson:"totalInbound"`
	TotalOutbound    int           `json:"totalOutbound" bson:"totalOutbound"`
	InboundDuration  int           `json:"inboundDuration" bson:"inboundDuration"`
	CreateDate       time.Time     `json:"createDate" bson:"createDate"`
	UpdateDate       time.Time     `json:"updateDate" bson:"updateDate"`
}

type TelnyxNumber struct {
	ID               bson.ObjectId `json:"id" bson:"_id"`
	UserID           bson.ObjectId ` json:"userID" bson:"userID"`
	PhoneID          string        `json:"phoneID" bson:"phoneID"`
	ConnectionID     string        `json:"connectionID" bson:"connectionID"`
	PhoneStatus      string        `json:"phoneStatus" bson:"phoneStatus"`
	PhoneNumber      string        `json:"phoneNumber" bson:"phoneNumber"`
	ForwardingNumber string        `json:"forwardingNumber" bson:"forwardingNumber"`
	TotalInbound     int           `json:"totalInbound" bson:"totalInbound"`
	TotalOutbound    int           `json:"totalOutbound" bson:"totalOutbound"`
	InboundDuration  int           `json:"inboundDuration" bson:"inboundDuration"`
	CreateDate       time.Time     `json:"createDate" bson:"createDate"`
	UpdateDate       time.Time     `json:"updateDate" bson:"updateDate"`
}

type DeductBalanceMessage struct {
	UserID bson.ObjectId `json:"userid,omitempty" bson:"userid"`
	Amount float32       `json:"amount,omitempty" bson:"amount"`
}

type BulkVsApiPurchaseRequest struct {
	TN         string `json:"TN"`
	Lidb       string `json:"Lidb"`
	PortoutPin string `json:"Portout Pin"`
	TrunkGroup string `json:"Trunk Group"`
	Sms        bool   `json:"Sms"`
	Mms        bool   `json:"Mms"`
}

type ForwardingNumberResponse struct {
	ForwardingNumber string `json:"forwardingNumber"`
	Message          string `json:"message,omitempty"`
	SubMessage       string `json:"submessage,omitempty"`
}

type BulkVsApiPurchaseResponse struct {
	TN         string `json:"TN"`
	Status     string `json:"Status"`
	Lidb       string `json:"Lidb"`
	PortoutPin string `json:"Portout Pin"`
	Routing    struct {
		TrunkGroup  string `json:"Trunk Group"`
		CustomURI   string `json:"Custom URI"`
		CallForward string `json:"Call Forward"`
	} `json:"Routing"`
	Messaging struct {
		Sms bool `json:"Sms"`
		Mms bool `json:"Mms"`
	} `json:"Messaging"`
	TNDetails struct {
		RateCenter     string `json:"Rate Center"`
		State          string `json:"State"`
		Tier           int    `json:"Tier"`
		Cnam           bool   `json:"Cnam"`
		ActivationDate string `json:"Activation Date"`
	} `json:"TN Details"`
}

type TelnyxList []TelnyxNumber
