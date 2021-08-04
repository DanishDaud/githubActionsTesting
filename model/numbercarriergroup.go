package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type TelnyxNumbersGroup struct {
	ID              bson.ObjectId `json:"id" bson:"_id"`
	UserID          bson.ObjectId `json:"userID" bson:"userID"`
	NumberID        bson.ObjectId `json:"numberID" bson:"numberID"`
	ContactListID   bson.ObjectId `json:"contactListID" bson:"contactListID"`
	ContactGroupID  bson.ObjectId `json:"contactGroupID" bson:"contactGroupID"`
	TotalInbound    int           `json:"totalInbound" bson:"totalInbound"`
	TotalOutbound   int           `json:"totalOutbound" bson:"totalOutbound"`
	InboundDuration int           `json:"inboundDuration" bson:"inboundDuration"`
	CreateDate      time.Time     `json:"createDate" bson:"createDate"`
	UpdateDate      time.Time     `json:"updateDate" bson:"updateDate"`
}

type TelnyxNumbersGroupAggr struct {
	ID              bson.ObjectId  `json:"id" bson:"_id"`
	UserID          bson.ObjectId  `json:"userID" bson:"userID"`
	NumberID        bson.ObjectId  `json:"numberID" bson:"numberID"`
	ContactListID   bson.ObjectId  `json:"contactListID" bson:"contactListID"`
	ContactGroupID  bson.ObjectId  `json:"contactGroupID" bson:"contactGroupID"`
	TotalInbound    int            `json:"totalInbound" bson:"totalInbound"`
	TotalOutbound   int            `json:"totalOutbound" bson:"totalOutbound"`
	InboundDuration int            `json:"inboundDuration" bson:"inboundDuration"`
	CreateDate      time.Time      `json:"createDate" bson:"createDate"`
	UpdateDate      time.Time      `json:"updateDate" bson:"updateDate"`
	ContactList     []ContactList  `json:"contactList" bson:"contactList"`
	Number          []TelnyxNumber `json:"numbers" bson:"numbers"`
}

type ContactListStats struct {
	ID            bson.ObjectId `json:"id" bson:"_id"`
	TotalInbound  int32         `json:"totalInbound" bson:"totalInbound"`
	TotalOutbound int32         `json:"totalOutbound" bson:"totalOutbound"`
	TotalDuration int32         `json:"inboundDuration" bson:"inboundDuration"`
	//ContactGroupID bson.ObjectId `json:"contactGroupID" bson:"contactGroupID"`
	//TotalInbound int			`json:"totalInbound" bson:"totalInbound"`
	//TotalOutbound	 int	`json:"totalOutbound" bson:"totalOutbound"`
	//InboundDuration		int `json:"inboundDuration" bson:"inboundDuration"`
	//CreateDate     time.Time     `json:"createDate" bson:"createDate"`
	//UpdateDate     time.Time     `json:"updateDate" bson:"updateDate"`
	//ContactList    []ContactList  `json:"contactList" bson:"contactList"`
	//TelnyxNumber   []TelnyxNumber `json:"telnyxNumber" bson:"telnyxNumber"`
}

type TelnyxContactGroup struct {
	Number         string        `json:"phone_numbers"`
	UserID         bson.ObjectId `bson:"userid"`
	ContactGroupID bson.ObjectId `json:"contactListID" bson:"contactListID"`
}
