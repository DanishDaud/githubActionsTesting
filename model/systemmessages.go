package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type EnumType int

const (
	CampaignCreate EnumType = 1 + iota
	SoundFileCreate
)

type SystemMessages struct {
	ID          bson.ObjectId `json:"_id" bson:"_id"`
	Message     string        `json:"message" bson:"message"`
	Description string        `json:"description" bson:"description"`
	Type        EnumType      `json:"type" bson:"type"`
	CreateDate  time.Time     `json:"createDate" bson:"createDate"`
	UpdateDate  time.Time     `json:"updateDate" bson:"updateDate"`
}
