package model

import "gopkg.in/mgo.v2/bson"

// User represents the structure of our resource
type ModuleType struct {
	ID             bson.ObjectId `json:"id" bson:"_id"`
	Type           int8          `json:"type" bson:"type"`
	Name           string        `json:"name" bson:"name"`
	Desc           string        `json:"desc" bson:"desc"`
	CampaignAccess []int8        `json:"campaignAccess" bson:"campaignAccess"`
}
